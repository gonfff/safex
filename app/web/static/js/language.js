(function () {
  const LANG_RU = "ru";
  const LANG_EN = "en";
  const STORAGE_KEY = "safex-language";
  const translations = {
    [LANG_EN]: {
      hero: {
        kicker: "Safe secrets exchange",
      },
      nav: {
        source: "Source code",
        faq: "FAQ",
      },
      guide: {
        intro: {
          title: "Safex is a lightweight secure sharing service",
          body: "Safex is a simple service whose only mission is safe data exchange. Encryption runs on the client, so even the developers cannot read your message. The server merely forwards the payload, and after the recipient opens the link both the link and the data self-destruct.",
        },
        steps: {
          title: "How to use Safex",
          one: "Attach a file or paste the text of your secret.",
          two: "Come up with a six-digit PIN code.",
          three: "Choose how long the link should remain available.",
          four: "Copy the generated link and share it with the recipient.",
          five: "Send the PIN via another communication channel.",
        },
        docs: {
          title: "Documentation",
          body: "Find detailed guides and API references in our documentation <a href='https://gonfff.github.io/safex/' target='_blank' class='link link-primary'>here</a>.",
        },
      },
      cta: {
        primary: "Generate link",
      },
      form: {
        title: "Create a secure message",
        subtitle:
          "Share sensitive information that will self-destruct after it is read.",
        newSecretBtn: "Create another secret",
        result: {
          badge: "Secret created",
          instructions:
            "Share the link with the recipient and send the PIN via another channel.",
          linkLabel: "Link",
          copyLink: "Copy link",
          pinLabel: "PIN",
          ttlLabel: "Available",
          ttlUnit: "minutes",
          fileLabel: "File",
          sizeLabel: "Size",
          sizeUnit: "bytes",
        },
        fields: {
          message: {
            label: "Message text or file",
            placeholder: "Paste secret text or drop a file below",
          },
          file: {
            dropPrimary: "Drop a file",
            dropSecondary: "or click to select",
            attachedLabel: "Attached:",
            clear: "Clear file",
          },
          pin: {
            label: "6-digit PIN",
            placeholder: "For example, 842119",
            helper: "The recipient must enter the PIN before decryption.",
          },
          expiry: {
            label: "Expiration",
            helper:
              "After the timer ends, the message is deleted automatically.",
            minutes: "minutes",
            hours: "hours",
            days: "days",
          },
        },
        cta: "Generate link",
      },
      retrieve: {
        form: {
          title: "Retrieve the message",
          subtitle: "Enter the PIN to decrypt the secret payload.",
          pinLabel: "PIN",
          pinPlaceholder: "For example, 842119",
          cta: "Unlock message",
        },
        result: {
          badge: "Message decrypted",
          title: "Secret available",
          subtitle:
            "The payload has been destroyed on the server. Save a copy if you still need it.",
          textLabel: "Message text",
          fileLabel: "Encrypted file",
          sizeLabel: "Size",
          copy: "Copy",
          download: "Download",
          sizeUnits: {
            b: "B",
            kb: "KB",
            mb: "MB",
            gb: "GB",
            tb: "TB",
            pb: "PB",
          },
        },
        errors: {
          invalidPin: "File deleted or PIN is wrong",
          finishPin: "Failed to complete PIN verification.",
          decryptionFailed: "Failed to decrypt message",
          initPin: "Failed to initiate PIN verification.",
          serverResponse: "Invalid server response.",
          noSecretId: "Secret identifier not specified",
          requestFailed: "Failed to execute request",
          defaultError: "Error",
          legacy: {
            invalidPin: "Invalid PIN or file already deleted",
          },
        },
      },
      limits: {
        size: " — Max size:",
        rate: " — Rate limit:",
        "time-dimension": "per minute",
      },
    },
    [LANG_RU]: {
      hero: {
        kicker: "Безопасный обмен секретами",
      },
      nav: {
        source: "Исходный код",
        faq: "FAQ",
      },
      guide: {
        intro: {
          title: "Safex — простой сервис для безопасного обмена",
          body: "Safex — это простой сервис, единственная цель которого — безопасный обмен данными. Шифрование происходит на стороне клиента, поэтому даже у разработчиков нет возможности прочитать сообщение. Сервер лишь передаёт информацию между пользователями, а после прочтения ссылка и данные саморазрушаются.",
        },
        docs: {
          title: "Документация",
          body: "Подробные руководства вы найдёте в документации <a href='https://gonfff.github.io/safex/' target='_blank' class='link link-primary'>здесь</a>.",
        },
        steps: {
          title: "Как пользоваться",
          one: "Прикрепите файл или введите текст сообщения.",
          two: "Придумайте шестизначный PIN-код.",
          three: "Определите, сколько времени ссылка будет доступна.",
          four: "Скопируйте сгенерированную ссылку и отправьте её получателю.",
          five: "Передайте PIN-код по другому каналу связи.",
        },
      },
      cta: {
        primary: "Создать ссылку",
      },
      form: {
        title: "Создать защищенное сообщение",
        subtitle:
          "Поделитесь чувствительной информацией, которая самоуничтожится после прочтения.",
        newSecretBtn: "Создать новый секрет",
        result: {
          badge: "Секрет создан",
          instructions:
            "Отправьте ссылку получателю и PIN по другому каналу связи.",
          linkLabel: "Ссылка",
          copyLink: "Скопировать ссылку",
          pinLabel: "PIN",
          ttlLabel: "Доступно",
          ttlUnit: "минут",
          fileLabel: "Файл",
          sizeLabel: "Размер",
          sizeUnit: "байт",
        },
        fields: {
          message: {
            label: "Текст сообщения или файл",
            placeholder: "Вставьте секретный текст или перетащите файл ниже",
          },
          file: {
            dropPrimary: "Перетащите файл",
            dropSecondary: "или нажмите, чтобы выбрать",
            attachedLabel: "Прикреплено:",
            clear: "Очистить файл",
          },
          pin: {
            label: "6-значный PIN",
            placeholder: "Например, 842119",
            helper: "Получателю понадобится PIN перед расшифровкой.",
          },
          expiry: {
            label: "Срок действия",
            helper: "После истечения срока сообщение удалится автоматически.",
            minutes: "минуты",
            hours: "часы",
            days: "дни",
          },
        },
        cta: "Создать ссылку",
      },
      retrieve: {
        form: {
          title: "Получить сообщение",
          subtitle: "Введите PIN-код, чтобы расшифровать секретное сообщение.",
          pinLabel: "PIN-код",
          pinPlaceholder: "Например, 842119",
          cta: "Получить сообщение",
        },
        result: {
          badge: "Сообщение расшифровано",
          title: "Секрет доступен",
          subtitle: "Сообщение удалено с сервера, сохраните его, если нужно.",
          textLabel: "Текст сообщения",
          fileLabel: "Зашифрованный файл",
          sizeLabel: "Размер",
          copy: "Скопировать",
          download: "Скачать",
          sizeUnits: {
            b: "Б",
            kb: "КБ",
            mb: "МБ",
            gb: "ГБ",
            tb: "ТБ",
            pb: "ПБ",
          },
        },
        errors: {
          invalidPin: "Файл удален или неверный пин-код",
          finishPin: "Не удалось завершить проверку PIN.",
          decryptionFailed: "Не удалось расшифровать сообщение",
          initPin: "Не удалось инициировать проверку PIN.",
          serverResponse: "Некорректный ответ сервера.",
          noSecretId: "Не указан идентификатор секрета",
          requestFailed: "Не удалось выполнить запрос",
          defaultError: "Ошибка",
          legacy: {
            invalidPin: "Неправильный PIN или файл уже удален",
          },
        },
      },
      limits: {
        size: " — Максимальный размер:",
        rate: " — Лимит запросов:",
        "time-dimension": "в минуту",
      },
    },
  };

  const SIZE_UNIT_KEYS = ["b", "kb", "mb", "gb", "tb", "pb"];

  function getSizeUnits(lang) {
    const fallback = translations[LANG_EN]?.retrieve?.result?.sizeUnits || {};
    return translations[lang]?.retrieve?.result?.sizeUnits || fallback;
  }

  function formatLocalizedSize(bytes, lang) {
    if (!Number.isFinite(bytes) || bytes < 0) {
      return "";
    }
    const units = getSizeUnits(lang);
    if (bytes < 1024) {
      const unit = units.b || "B";
      return `${bytes} ${unit}`;
    }
    let value = bytes;
    let unitIndex = 0;
    while (value >= 1024 && unitIndex < SIZE_UNIT_KEYS.length - 1) {
      value /= 1024;
      unitIndex += 1;
    }
    const unitKey = SIZE_UNIT_KEYS[unitIndex] || "pb";
    const label = units[unitKey] || units.pb || "";
    return `${value.toFixed(1)} ${label}`.trim();
  }

  const langButtons = document.querySelectorAll("[data-lang]");

  function resolveTranslation(dict, path) {
    return path
      .split(".")
      .reduce((acc, part) => (acc ? acc[part] : undefined), dict);
  }

  function updateLanguageButtons(activeLang) {
    langButtons.forEach((btn) => {
      const isActive = btn.dataset.lang === activeLang;
      btn.classList.toggle("btn-active", isActive);
      btn.setAttribute("aria-pressed", String(isActive));
    });
  }

  function applyLanguage(lang) {
    const languageKey = translations[lang] ? lang : LANG_EN;
    const activeDictionary = translations[languageKey];
    currentLanguage = languageKey;
    document.querySelectorAll("[data-i18n]").forEach((node) => {
      const key = node.getAttribute("data-i18n");
      const value = resolveTranslation(activeDictionary, key);
      if (typeof value === "string") {
        if (node.hasAttribute("data-i18n-html")) {
          node.innerHTML = value;
        } else {
          node.textContent = value;
        }
      }
    });
    document.querySelectorAll("[data-i18n-placeholder]").forEach((node) => {
      const key = node.getAttribute("data-i18n-placeholder");
      const value = resolveTranslation(activeDictionary, key);
      if (typeof value === "string") {
        node.setAttribute("placeholder", value);
      }
    });
    document.querySelectorAll("[data-i18n-size]").forEach((node) => {
      const bytes = Number(node.dataset.bytes);
      if (Number.isFinite(bytes)) {
        node.textContent = formatLocalizedSize(bytes, languageKey);
      }
    });
    document.documentElement.lang = languageKey === LANG_RU ? "ru" : "en";
    localStorage.setItem(STORAGE_KEY, languageKey);
    updateLanguageButtons(languageKey);
  }

  let currentLanguage = LANG_EN;

  function detectInitialLanguage() {
    const saved = localStorage.getItem(STORAGE_KEY);
    if (translations[saved]) {
      return saved;
    }
    const browserLang =
      navigator.language && navigator.language.toLowerCase().startsWith("ru")
        ? LANG_RU
        : LANG_EN;
    return translations[browserLang] ? browserLang : LANG_EN;
  }

  function initLanguage() {
    const initialLang = detectInitialLanguage();
    applyLanguage(initialLang);
    document.addEventListener("safex:refresh-language", () => {
      applyLanguage(currentLanguage);
    });
    langButtons.forEach((btn) => {
      btn.addEventListener("click", () => {
        applyLanguage(btn.dataset.lang);
      });
    });
  }

  initLanguage();

  // Export functions for use in other modules
  window.safexLang = {
    LANG_RU,
    LANG_EN,
    getDocumentLanguage: () => {
      const langAttr = document.documentElement.lang?.toLowerCase();
      if (langAttr && langAttr.startsWith(LANG_RU)) {
        return LANG_RU;
      }
      return LANG_EN;
    },
    translate: (key, specificLang = null) => {
      const lang = specificLang || window.safexLang.getDocumentLanguage();
      return (
        resolveTranslation(translations[lang] || translations[LANG_EN], key) ||
        ""
      );
    },
  };
})();
