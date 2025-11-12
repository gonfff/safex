(function () {
  const navbarRoot = document.getElementById("navbar-root");
  if (!navbarRoot) {
    return;
  }

  navbarRoot.innerHTML = `
        <header class="navbar bg-base-200 px-6 py-3">
            <div class="navbar-start items-center gap-3">
                <a href="./" class="text-xl font-bold">Safex</a>
                <span class="text-sm opacity-70" data-i18n="hero.kicker"></span>
            </div>
            <div class="navbar-end items-center gap-3">
                <nav class="hidden md:flex items-center gap-3 text-sm">
                    <a href="https://github.com/gonfff/safex" class="link link-hover" data-i18n="nav.source" target="_blank" rel="noreferrer noopener"></a>
                    <a href="./faq.html" class="link link-hover" data-i18n="nav.faq"></a>
                </nav>
                <div class="dropdown dropdown-end">
                    <button
                        class="btn btn-ghost btn-square btn-sm rounded-lg border border-transparent hover:border-base-content/60"
                        tabindex="0" role="button" aria-label="Switch language">
                        <svg viewBox="0 0 25 25" class="h-8 w-8" fill="none" xmlns="http://www.w3.org/2000/svg"
                            stroke="currentColor" stroke-width="1.2">
                            <path stroke-linecap="round" stroke-linejoin="round"
                                d="M5.5 16.5H19.5M5.5 8.5H19.5M4.5 12.5H20.5M12.5 20.5C12.5 20.5 8 18.5 8 12.5C8 6.5 12.5 4.5 12.5 4.5M12.5 4.5C12.5 4.5 17 6.5 17 12.5C17 18.5 12.5 20.5 12.5 20.5M12.5 4.5V20.5M20.5 12.5C20.5 16.9183 16.9183 20.5 12.5 20.5C8.08172 20.5 4.5 16.9183 4.5 12.5C4.5 8.08172 8.08172 4.5 12.5 4.5C16.9183 4.5 20.5 8.08172 20.5 12.5Z" />
                        </svg>
                    </button>
                    <ul tabindex="0"
                        class="dropdown-content menu menu-sm bg-base-200 rounded-box z-[1] mt-3 w-28 p-2 shadow">
                        <li><button class="btn btn-ghost btn-sm justify-start" type="button" data-lang="ru">RU</button></li>
                        <li><button class="btn btn-ghost btn-sm justify-start" type="button" data-lang="en">EN</button></li>
                    </ul>
                </div>
                <label
                    class="swap swap-rotate btn btn-ghost btn-square btn-sm rounded-lg border border-transparent hover:border-base-content/60"
                    aria-label="Toggle theme">
                    <input id="theme-toggle" type="checkbox" class="theme-controller" />
                    <svg class="swap-off h-6 w-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M12 4.5v-3M12 22.5v-3M4.5 12h-3M22.5 12h-3" />
                        <path d="M5.636 5.636l-2.121-2.12M20.485 20.485l-2.121-2.121" />
                        <path d="M5.636 18.364l-2.121 2.121M20.485 3.515l-2.121 2.121" />
                        <path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                    <svg class="swap-on h-6 w-6" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 25 25" fill="none"
                        stroke="currentColor" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M21 12.79A9 9 0 1111.21 3 7 7 0 0021 12.79z" />
                    </svg>
                </label>
            </div>
        </header>
    `;
})();
