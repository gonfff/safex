(function () {
    const THEME_LIGHT = 'winter';
    const THEME_DARK = 'dim';
    const STORAGE_KEY = 'safex-theme';
    const root = document.documentElement;
    const toggle = document.getElementById('theme-toggle');

    if (!toggle) {
        return;
    }

    const savedTheme = localStorage.getItem(STORAGE_KEY);
    const initialTheme = savedTheme === THEME_DARK ? THEME_DARK : THEME_LIGHT;
    toggle.checked = initialTheme === THEME_DARK;
    root.setAttribute('data-theme', initialTheme);

    toggle.addEventListener('change', () => {
        const nextTheme = toggle.checked ? THEME_DARK : THEME_LIGHT;
        root.setAttribute('data-theme', nextTheme);
        localStorage.setItem(STORAGE_KEY, nextTheme);
    });
})();
