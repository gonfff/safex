import init, { decrypt, finish_login, start_login } from "/static/wasm/safex_wasm.js";

const LANG_RU = "ru";
const LANG_EN = "en";
const TEXT = {
  invalidPin: {
    [LANG_RU]: "Файл удален или неверный пин-код",
    [LANG_EN]: "File deleted or PIN is wrong",
  },
  finishPin: {
    [LANG_RU]: "Не удалось завершить проверку PIN.",
    [LANG_EN]: "Failed to complete PIN verification.",
  },
};
const LEGACY_INVALID_PIN_MESSAGES = [
  "Неправильный PIN или файл уже удален",
  "Invalid PIN or file already deleted",
];
const SERVER_INVALID_PIN_MESSAGES = [
  TEXT.invalidPin[LANG_RU],
  TEXT.invalidPin[LANG_EN],
  ...LEGACY_INVALID_PIN_MESSAGES,
];

const getDocumentLanguage = () => {
  const langAttr = document.documentElement.lang?.toLowerCase();
  if (langAttr && langAttr.startsWith(LANG_RU)) {
    return LANG_RU;
  }
  return LANG_EN;
};

const translate = (key) => {
  const lang = getDocumentLanguage();
  return TEXT[key]?.[lang] || TEXT[key]?.[LANG_EN] || "";
};

class RevealFlowError extends Error {
  constructor(message, options = {}) {
    super(message);
    this.invalidPin = Boolean(options.invalidPin);
  }
}

(async () => {
  await init();

  const form = document.getElementById("unlock-form");
  const pinInput = document.getElementById("unlock-pin");
  const secretIdInput = document.querySelector("#unlock-form input[name='secret_id']");
  const resultContainer = document.getElementById("reveal-result");
  const indicator = document.getElementById("reveal-indicator");
  const submitButton = form?.querySelector("button[type='submit']");

  if (!form || !pinInput || !secretIdInput || !resultContainer) {
    return;
  }

  const textDecoder = new TextDecoder();

  const bytesToBase64 = (bytes) => {
    if (!bytes || bytes.length === 0) {
      return "";
    }
    let binary = "";
    const chunkSize = 0x8000;
    for (let i = 0; i < bytes.length; i += chunkSize) {
      const chunk = bytes.subarray(i, i + chunkSize);
      binary += String.fromCharCode.apply(null, chunk);
    }
    return btoa(binary);
  };

  const base64ToBytes = (base64) => {
    if (!base64) {
      return new Uint8Array();
    }
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i += 1) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  };

  const setLoading = (loading) => {
    if (loading) {
      indicator?.classList.remove("hidden");
    } else {
      indicator?.classList.add("hidden");
    }
    pinInput.disabled = loading;
    if (submitButton) {
      submitButton.disabled = loading;
    }
  };

  const renderErrorCard = (message, { invalidPin = false } = {}) => {
    const card = document.createElement("div");
    card.className = "card bg-base-200 border border-base-300 shadow-xl";

    const body = document.createElement("div");
    body.className = "card-body space-y-4";

    const alert = document.createElement("div");
    alert.className = "alert alert-error";

    const span = document.createElement("span");
    if (invalidPin) {
      span.dataset.i18n = "retrieve.errors.invalidPin";
      span.textContent = translate("invalidPin");
    } else {
      span.textContent = message || "Ошибка";
    }

    alert.appendChild(span);
    body.appendChild(alert);
    card.appendChild(body);

    resultContainer.innerHTML = "";
    resultContainer.appendChild(card);
    document.dispatchEvent(new Event("safex:refresh-language"));
    pinInput.focus();
  };

  const renderRevealHtml = (html) => {
    const wrapper = document.createElement("div");
    wrapper.innerHTML = html;
    const swapNode = wrapper.querySelector("#unlock-card[hx-swap-oob]");
    let card = null;

    if (swapNode) {
      const existingCard = document.getElementById("unlock-card");
      if (existingCard) {
        existingCard.replaceWith(swapNode);
        resultContainer.innerHTML = "";
      } else {
        resultContainer.replaceChildren(swapNode);
      }
      card = swapNode;
    } else {
      const nodes = Array.from(wrapper.childNodes);
      resultContainer.replaceChildren(...nodes);
    }

    document.dispatchEvent(new Event("safex:refresh-language"));
    pinInput.focus();
    return card;
  };

  const decryptAndRender = (card, exportKey) => {
    if (!card || !exportKey) {
      return;
    }
    const downloadButton = card.querySelector('[data-action="download"]');
    const encryptedBase64 = downloadButton?.dataset.downloadBase64 || card.dataset.secretPayload || "";
    if (!encryptedBase64) {
      return;
    }

    let decryptedBytes;
    try {
      decryptedBytes = decrypt(exportKey, base64ToBytes(encryptedBase64));
    } catch (error) {
      console.error("Decryption failed", error);
      renderErrorCard("Не удалось расшифровать сообщение");
      return;
    }

    const decryptedBase64 = bytesToBase64(decryptedBytes);
    if (downloadButton) {
      downloadButton.dataset.downloadBase64 = decryptedBase64;
      const originalName = downloadButton.dataset.downloadName || "secret.bin";
      const sanitizedName = originalName.replace(/\.encrypted$/i, "") || originalName;
      downloadButton.dataset.downloadName = sanitizedName;
      const fileNameLabel = card.querySelector("#file-name");
      if (fileNameLabel) {
        fileNameLabel.textContent = sanitizedName;
      }
    }

    if (card.dataset.secretType === "text") {
      const wrapper = card.querySelector("#message-content-wrapper");
      const messageNode = card.querySelector("#message-content");
      if (wrapper && messageNode) {
        wrapper.classList.remove("hidden");
        try {
          messageNode.textContent = textDecoder.decode(decryptedBytes);
        } catch (error) {
          console.error("Text decoding failed", error);
          messageNode.textContent = "";
        }
      }
    }
  };

  const startOpaqueLogin = async (pin, secretId) => {
    let loginStart;
    try {
      loginStart = start_login(pin);
    } catch (error) {
      throw new RevealFlowError("Не удалось инициировать проверку PIN.");
    }

    const handle = loginStart.handle;
    const clientMessage = loginStart.message;
    loginStart.free();

    const response = await fetch("/opaque/login/start", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        secretId,
        request: bytesToBase64(clientMessage),
      }),
    });

    let payload = null;
    try {
      payload = await response.json();
    } catch (error) {
      payload = null;
    }

    if (!response.ok || !payload) {
      const message = payload?.error || "Ошибка инициализации протокола";
      const normalized = typeof message === "string" ? message.trim() : "";
      const invalidPin = normalized && SERVER_INVALID_PIN_MESSAGES.includes(normalized);
      const displayMessage = invalidPin ? translate("invalidPin") : message;
      throw new RevealFlowError(displayMessage, { invalidPin });
    }

    const { sessionId, response: serverResponse } = payload;
    if (!sessionId || !serverResponse) {
      throw new RevealFlowError("Некорректный ответ сервера.");
    }
    return { handle, serverResponse, sessionId };
  };

  const finishOpaqueLogin = (handle, pin, serverResponse) => {
    if (!serverResponse) {
      throw new RevealFlowError("Некорректный ответ сервера.");
    }
    const serverBytes = base64ToBytes(serverResponse);
    let loginFinish;
    try {
      loginFinish = finish_login(handle, pin, serverBytes);
    } catch (error) {
      throw new RevealFlowError(translate("finishPin"));
    }

    const exportKey = loginFinish.export_key;
    const finalizationB64 = bytesToBase64(loginFinish.finalization);
    loginFinish.free();
    return { exportKey, finalizationB64 };
  };

  const revealSecret = async (pin) => {
    const secretId = secretIdInput.value.trim();
    if (!secretId) {
      throw new RevealFlowError("Не указан идентификатор секрета");
    }

    const loginInit = await startOpaqueLogin(pin, secretId);
    const { exportKey, finalizationB64 } = finishOpaqueLogin(
      loginInit.handle,
      pin,
      loginInit.serverResponse,
    );

    const formData = new FormData();
    formData.set("secret_id", secretId);
    formData.set("session_id", loginInit.sessionId || "");
    formData.set("finalization", finalizationB64);

    const response = await fetch("/secrets/reveal", {
      method: "POST",
      body: formData,
    });

    const html = await response.text();
    return { html, exportKey };
  };

  form.addEventListener("submit", async (event) => {
    event.preventDefault();
    if (!form.reportValidity()) {
      return;
    }

    const pin = pinInput.value.trim();
    setLoading(true);
    try {
      const { html, exportKey } = await revealSecret(pin);
      const card = renderRevealHtml(html);
      if (card) {
        decryptAndRender(card, exportKey);
      }
    } catch (error) {
      console.error("Reveal failed", error);
      if (error instanceof RevealFlowError) {
        renderErrorCard(error.message, { invalidPin: error.invalidPin });
      } else {
        renderErrorCard("Не удалось выполнить запрос");
      }
    } finally {
      setLoading(false);
    }
  });

  const flashButtonState = (button, stateClass) => {
    if (!button) {
      return;
    }
    button.classList.add(stateClass);
    setTimeout(() => button.classList.remove(stateClass), 1200);
  };

  const copyFromTarget = async (targetId, button) => {
    const node = document.getElementById(targetId);
    if (!node) {
      return;
    }
    const payload = node.textContent?.trim();
    if (!payload) {
      return;
    }
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(payload);
      } else {
        const temp = document.createElement("textarea");
        temp.value = payload;
        temp.setAttribute("readonly", "readonly");
        temp.style.position = "absolute";
        temp.style.left = "-9999px";
        document.body.appendChild(temp);
        temp.select();
        document.execCommand("copy");
        temp.remove();
      }
      flashButtonState(button, "btn-success");
    } catch (error) {
      console.error("Clipboard copy failed", error);
      flashButtonState(button, "btn-error");
    }
  };

  const downloadFromButton = (button) => {
    const base64 = button.dataset.downloadBase64 || "";
    const fileName = button.dataset.downloadName || "secret.bin";
    const contentType = button.dataset.downloadType || "application/octet-stream";
    if (!base64) {
      return;
    }
    const byteString = atob(base64);
    const buffer = new Uint8Array(byteString.length);
    for (let i = 0; i < byteString.length; i += 1) {
      buffer[i] = byteString.charCodeAt(i);
    }
    const blob = new Blob([buffer], { type: contentType });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    URL.revokeObjectURL(url);
    flashButtonState(button, "btn-success");
  };

  document.addEventListener("click", (event) => {
    const copyBtn = event.target.closest('[data-action="copy"]');
    if (copyBtn) {
      event.preventDefault();
      copyFromTarget(copyBtn.dataset.copyTarget, copyBtn);
      return;
    }
    const downloadBtn = event.target.closest('[data-action="download"]');
    if (downloadBtn) {
      event.preventDefault();
      downloadFromButton(downloadBtn);
    }
  });
})();
