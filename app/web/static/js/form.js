import init, {
  encrypt,
  finish_registration,
  start_registration,
} from "/static/wasm/safex_rust.js";

(async () => {
  await init();

  (() => {
    const messageInput = document.getElementById("secret-message");
    const fileInput = document.getElementById("secret-file");
    const dropzone = document.getElementById("secret-file-dropzone");
    const fileInfo = document.getElementById("secret-file-info");
    const fileNameLabel = document.getElementById("secret-file-name");
    const clearFileButton = document.getElementById("clear-file-button");
    const form = document.getElementById("create-form");
    const pinInput = document.getElementById("pin-code");
    const createCard = document.getElementById("create-card");
    const resultContainer = document.getElementById("create-result");

    if (!messageInput || !fileInput || !dropzone || !form || !pinInput) {
      return;
    }

    const maxFileBytes = Number(fileInput.dataset.maxBytes || "0");
    const maxFileMB = Number(fileInput.dataset.maxMb || "0");
    const formatMaxSizeLabel = () => {
      if (maxFileMB > 0) {
        return `${maxFileMB} MB`;
      }
      if (maxFileBytes > 0) {
        return `${(maxFileBytes / (1024 * 1024)).toFixed(1)} MB`;
      }
      return "";
    };

    const toggleFileArea = (disabled) => {
      fileInput.disabled = disabled;
      dropzone.classList.toggle("opacity-50", disabled);
      dropzone.classList.toggle("pointer-events-none", disabled);
      dropzone.setAttribute("aria-disabled", disabled ? "true" : "false");
    };

    const toggleMessageArea = (disabled) => {
      messageInput.disabled = disabled;
      messageInput.classList.toggle("textarea-disabled", disabled);
      messageInput.classList.toggle("bg-base-200", disabled);
      messageInput.classList.toggle("bg-base-100", !disabled);
    };

    const hideFileInfo = () => {
      if (fileInfo) {
        fileInfo.classList.add("hidden");
      }
      if (fileNameLabel) {
        fileNameLabel.textContent = "";
      }
      clearFileButton?.classList.add("hidden");
    };

    const showFileInfo = (name) => {
      if (fileNameLabel) {
        fileNameLabel.textContent = name;
      }
      if (fileInfo) {
        fileInfo.classList.remove("hidden");
      }
      clearFileButton?.classList.remove("hidden");
    };

    const resetFileInput = () => {
      fileInput.value = "";
      hideFileInfo();
    };

    const hasMessage = () => messageInput.value.trim().length > 0;

    messageInput.addEventListener("input", () => {
      const textPresent = hasMessage();
      toggleFileArea(textPresent);

      if (!textPresent) {
        toggleMessageArea(false);
        return;
      }

      if (fileInput.value) {
        resetFileInput();
        toggleMessageArea(false);
      }
    });

    fileInput.addEventListener("change", () => {
      const file = fileInput.files && fileInput.files[0];
      const hasFile = Boolean(file);
      toggleMessageArea(hasFile);

      if (!hasFile) {
        toggleFileArea(hasMessage());
        resetFileInput();
        return;
      }

      messageInput.value = "";
      toggleFileArea(true);
      showFileInfo(file.name);
    });

    clearFileButton?.addEventListener("click", () => {
      resetFileInput();
      toggleMessageArea(false);
      toggleFileArea(hasMessage());
      messageInput.focus();
    });

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

    const registerOpaqueSecret = async (pin) => {
      const registrationStart = start_registration(pin);
      const handle = registrationStart.handle;
      const clientMessage = registrationStart.message;
      registrationStart.free();

      const registerResponse = await fetch("/opaque/register/start", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          request: bytesToBase64(clientMessage),
        }),
      });

      let registerPayload = null;
      try {
        registerPayload = await registerResponse.json();
      } catch (error) {
        // ignore json parsing errors to provide a generic error below
      }

      if (!registerResponse.ok || !registerPayload) {
        const message =
          registerPayload?.error || "Error during opaque registration start";
        throw new Error(message);
      }
      const { secretId, response } = registerPayload;
      if (!secretId || !response) {
        throw new Error("Invalid response from server");
      }

      const serverResponseBytes = base64ToBytes(response);
      const registrationFinish = finish_registration(
        handle,
        pin,
        serverResponseBytes
      );
      try {
        return {
          secretId,
          exportKey: registrationFinish.export_key,
          opaqueUpload: registrationFinish.upload,
        };
      } finally {
        registrationFinish.free();
      }
    };

    const PAYLOAD_TYPE_FIELD = "payload_type";
    const PayloadType = {
      FILE: "file",
      TEXT: "text",
    };

    // Intercept form submission for encryption
    form.addEventListener("submit", async (e) => {
      e.preventDefault();

      const pin = pinInput.value;
      if (!pin || pin.length !== 6) {
        alert("PIN should be 6 characters long");
        return;
      }

      const formData = new FormData(form);
      const file = fileInput.files && fileInput.files[0];
      const message = messageInput.value.trim();

      if (!file && !message) {
        alert("You need to select a file or enter a message");
        return;
      }

      try {
        const registration = await registerOpaqueSecret(pin);
        formData.set("secret_id", registration.secretId);
        formData.set("opaque_upload", bytesToBase64(registration.opaqueUpload));

        if (file) {
          // Encrypt file
          const fileBytes = new Uint8Array(await file.arrayBuffer());
          const encryptedBytes = encrypt(registration.exportKey, fileBytes);

          // Create a new blob with encrypted data
          const encryptedBlob = new Blob([encryptedBytes], {
            type: "application/octet-stream",
          });
          formData.set("file", encryptedBlob, file.name + ".encrypted");
          formData.set(PAYLOAD_TYPE_FIELD, PayloadType.FILE);
        } else {
          // Encrypt text message
          const messageBytes = new TextEncoder().encode(message);
          const encryptedBytes = encrypt(registration.exportKey, messageBytes);

          // Create a blob for the encrypted message
          const encryptedBlob = new Blob([encryptedBytes], {
            type: "application/octet-stream",
          });
          formData.set("file", encryptedBlob, "message.encrypted");
          formData.delete("message");
          formData.set(PAYLOAD_TYPE_FIELD, PayloadType.TEXT);
        }

        // Delete PIN from form data to avoid sending it to the server
        formData.delete("pin");

        // Send encrypted data to the server
        const response = await fetch("/secrets", {
          method: "POST",
          body: formData,
        });

        const result = await response.text();
        if (resultContainer) {
          resultContainer.innerHTML = result;
          document.dispatchEvent(new Event("safex:refresh-language"));
        }

        if (response.ok && response.status === 201) {
          createCard?.classList.add("hidden");
          resultContainer?.scrollIntoView({
            behavior: "smooth",
            block: "start",
          });
        } else {
          createCard?.classList.remove("hidden");
        }
      } catch (error) {
        console.error("Error during encryption:", error);
        const errorMessage =
          error instanceof Error && error.message
            ? error.message
            : "Error during data encryption";
        alert(errorMessage);
        createCard?.classList.remove("hidden");
      }
    });
  })();
})();
