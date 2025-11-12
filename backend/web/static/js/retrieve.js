(() => {
	const form = document.getElementById("unlock-form");
	const pinInput = document.getElementById("unlock-pin");
	if (!form || !pinInput) {
		return;
	}

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

	document.body.addEventListener("htmx:afterSwap", (event) => {
		if (event.target?.id === "reveal-result") {
			pinInput.focus();
		}
	});
})();
