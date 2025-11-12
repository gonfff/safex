import init, { encrypt } from "/static/pkg/safex_wasm.js";

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

    const PAYLOAD_TYPE_FIELD = "payload_type";
    const PayloadType = {
      FILE: "file",
      TEXT: "text",
    };

    // Перехватываем отправку формы для шифрования
    form.addEventListener("submit", async (e) => {
      e.preventDefault();

      const pin = pinInput.value;
      if (!pin || pin.length !== 6) {
        alert("Пожалуйста, введите 6-значный пинкод");
        return;
      }

      const formData = new FormData(form);
      const file = fileInput.files[0];
      const message = messageInput.value.trim();

      try {
        if (file) {
          // Шифруем файл
          const fileBytes = new Uint8Array(await file.arrayBuffer());
          const encryptedBytes = encrypt(pin, fileBytes);

          // Создаем новый blob с зашифрованными данными
          const encryptedBlob = new Blob([encryptedBytes], {
            type: "application/octet-stream",
          });
          formData.set("file", encryptedBlob, file.name + ".encrypted");
          formData.set(PAYLOAD_TYPE_FIELD, PayloadType.FILE);
        } else if (message) {
          // Шифруем текстовое сообщение
          const messageBytes = new TextEncoder().encode(message);
          const encryptedBytes = encrypt(pin, messageBytes);

          // Создаем blob для зашифрованного сообщения
          const encryptedBlob = new Blob([encryptedBytes], {
            type: "application/octet-stream",
          });
          formData.set("file", encryptedBlob, "message.encrypted");
          formData.delete("message");
          formData.set(PAYLOAD_TYPE_FIELD, PayloadType.TEXT);
        } else {
          alert("Нужно выбрать файл или ввести сообщение");
          return;
        }

        // Удаляем пинкод из formData (не отправляем на сервер)
        formData.delete("pin");

        // Отправляем зашифрованные данные
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
        console.error("Ошибка шифрования:", error);
        alert("Ошибка при шифровании данных");
        createCard?.classList.remove("hidden");
      }
    });
  })();
})();
