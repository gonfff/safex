(() => {
  const form = document.getElementById("unlock-form");
  const pinInput = document.getElementById("unlock-pin");
  const messageCard = document.getElementById("message-card");
  const messageContent = document.getElementById("message-content");
  const messageWrapper = document.getElementById("message-content-wrapper");
  const fileWrapper = document.getElementById("file-content-wrapper");
  const fileNameLabel = document.getElementById("file-name");
  const copyButton = document.getElementById("copy-button");
  const downloadButton = document.getElementById("download-button");

  if (!form || !pinInput || !messageCard || !copyButton || !downloadButton) {
    return;
  }

  const DEFAULT_TEXT =
    messageContent?.textContent.trim() ||
    "Пароль от хранилища: 9Sdq921!"; // Placeholder until API integration
  let activeFileUrl = "";
  let activeFileName = "";

  const wait = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

  const setLoadingState = (isLoading) => {
    const submitButton = form.querySelector("button[type='submit']");
    if (!submitButton) {
      return;
    }
    submitButton.classList.toggle("loading", isLoading);
    submitButton.toggleAttribute("disabled", isLoading);
  };

  const toggleResultMode = (mode) => {
    const isFile = mode === "file";
    messageWrapper?.classList.toggle("hidden", isFile);
    copyButton?.classList.toggle("hidden", isFile);
    fileWrapper?.classList.toggle("hidden", !isFile);
    downloadButton?.classList.toggle("hidden", !isFile);
  };

  const renderResult = ({ type, payload, fileName, fileUrl }) => {
    if (!messageCard) {
      return;
    }
    if (type === "file") {
      toggleResultMode("file");
      activeFileUrl = fileUrl || "";
      activeFileName = fileName || "secret.bin";
      if (fileNameLabel) {
        fileNameLabel.textContent = activeFileName;
      }
    } else {
      toggleResultMode("text");
      if (messageContent) {
        messageContent.textContent = payload || DEFAULT_TEXT;
      }
    }
    messageCard.classList.remove("hidden");
    messageCard.scrollIntoView({ behavior: "smooth", block: "center" });
  };

  const fetchSecret = async (pin) => {
    // Placeholder for future API request. Replace with real fetch when backend is ready.
    await wait(600);
    const responseType = form.dataset.responseType === "file" ? "file" : "text";
    if (responseType === "file") {
      return {
        type: "file",
        fileName: form.dataset.fileName || "secret.bin",
        fileUrl: form.dataset.fileUrl || "",
      };
    }
    return {
      type: "text",
      payload: form.dataset.textPayload || DEFAULT_TEXT,
    };
  };

  const handleSubmit = async (event) => {
    event.preventDefault();
    if (!pinInput.value.trim()) {
      pinInput.focus();
      return;
    }
    setLoadingState(true);
    try {
      const result = await fetchSecret(pinInput.value.trim());
      renderResult(result);
    } catch (error) {
      console.error("Unable to unlock the message", error);
    } finally {
      setLoadingState(false);
    }
  };

  const flashButtonState = (button, stateClass) => {
    if (!button) {
      return;
    }
    const original = button.dataset.originalClass || button.className;
    if (!button.dataset.originalClass) {
      button.dataset.originalClass = original;
    }
    button.className = `${original} ${stateClass}`;
    setTimeout(() => {
      button.className = button.dataset.originalClass;
    }, 1600);
  };

  const copyToClipboard = async () => {
    if (!messageContent) {
      return;
    }
    const payload = messageContent.textContent.trim();
    if (!payload) {
      return;
    }
    try {
      if (navigator.clipboard?.writeText) {
        await navigator.clipboard.writeText(payload);
      } else {
        const temp = document.createElement("textarea");
        temp.value = payload;
        temp.setAttribute("readonly", "");
        temp.style.position = "absolute";
        temp.style.left = "-9999px";
        document.body.appendChild(temp);
        temp.select();
        document.execCommand("copy");
        temp.remove();
      }
      flashButtonState(copyButton, "btn-success");
    } catch (error) {
      console.error("Clipboard copy failed", error);
      flashButtonState(copyButton, "btn-error");
    }
  };

  const downloadFile = () => {
    const name = activeFileName || "secret.bin";
    let url = activeFileUrl;
    let revokeUrl = false;
    if (!url) {
      const blob = new Blob(
        ["Safex placeholder file. Replace with backend response."],
        { type: "application/octet-stream" }
      );
      url = URL.createObjectURL(blob);
      revokeUrl = true;
    }
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = name;
    document.body.appendChild(anchor);
    anchor.click();
    anchor.remove();
    if (revokeUrl) {
      URL.revokeObjectURL(url);
    }
    flashButtonState(downloadButton, "btn-success");
  };

  form.addEventListener("submit", handleSubmit);
  copyButton.addEventListener("click", copyToClipboard);
  downloadButton.addEventListener("click", downloadFile);
})();
