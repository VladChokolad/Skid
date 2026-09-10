// Frontend/js/nav.js — единая шапка/навигация для всех страниц.
//
// Использование: <div id="site-header" data-page="parties"></div>
// data-page — id текущей страницы (подсвечивается как активная ссылка).
// data-prefix — префикс для ссылок со страниц во вложенных папках
// (например data-prefix="../" на страницах внутри auth/).
//
// Зависит от Api и Session (js/api.js, js/session.js) — подключать после них.

const SiteNav = (function () {
    const links = [
        { id: 'main', href: 'main.html', label: 'Главная' },
        { id: 'parties', href: 'parties.html', label: 'Мои тусовки' },
        { id: 'profile', href: 'profile.html', label: 'Профиль' },
        { id: 'create_party', href: 'create_party.html', label: 'Создать тусовку' }
    ];

    const guestLinks = [
        { id: 'login', href: 'auth/login.html', label: 'Войти' },
        { id: 'register', href: 'auth/register.html', label: 'Зарегистрироваться' }
    ];

    function linksHtml(list, page, prefix) {
        return list.map(function (link) {
            const isActive = link.id === page;
            return '<a href="' + prefix + link.href + '"' +
                (isActive ? ' class="active" aria-current="page"' : '') + '>' + link.label + '</a>';
        }).join(' | ');
    }

    // Текущая страница как путь относительно Frontend/ (для ?redirect=) —
    // в проекте всего два уровня вложенности: корень и auth/
    function currentPageRelativePath() {
        const path = window.location.pathname;
        const inAuth = path.indexOf('/auth/') !== -1;
        const filename = path.substring(path.lastIndexOf('/') + 1);
        return (inAuth ? 'auth/' : '') + filename + window.location.search;
    }

    // Ссылки "Войти"/"Зарегистрироваться" несут ?redirect=, чтобы после входа
    // вернуть пользователя туда, откуда он пришёл (а не всегда на main.html)
    function guestLinksHtml(page, prefix) {
        const redirect = encodeURIComponent(currentPageRelativePath());
        return guestLinks.map(function (link) {
            const isActive = link.id === page;
            return '<a href="' + prefix + link.href + '?redirect=' + redirect + '"' +
                (isActive ? ' class="active" aria-current="page"' : '') + '>' + link.label + '</a>';
        }).join(' | ');
    }

    async function init() {
        const container = document.getElementById('site-header');
        if (!container) return;

        const page = container.dataset.page || '';
        const prefix = container.dataset.prefix || '';

        container.innerHTML = linksHtml(links, page, prefix);

        // Гостю (аноним/не удалось определить) — предлагаем войти/зарегистрироваться
        const me = await Session.loadMe();
        if (!me || me.isAnon) {
            container.innerHTML += ' | ' + guestLinksHtml(page, prefix);
        }

        // Другие скрипты на странице (main.html) могут отреагировать на "кто я"
        // без повторного запроса /profile
        document.dispatchEvent(new CustomEvent('sitenav:me', { detail: me }));
    }

    document.addEventListener('DOMContentLoaded', init);
    return { init };
})();
