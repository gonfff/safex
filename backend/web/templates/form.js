(() => {
  const messageInput = document.getElementById("secret-message");
  const fileInput = document.getElementById("secret-file");
  const dropzone = document.getElementById("secret-file-dropzone");
  const fileInfo = document.getElementById("secret-file-info");
  const fileNameLabel = document.getElementById("secret-file-name");
  const clearFileButton = document.getElementById("clear-file-button");

  if (!messageInput || !fileInput || !dropzone) {
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
})();
