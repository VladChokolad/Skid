// Frontend/js/api.js — общая обёртка над fetch и мелкие хелперы,
// используемые всеми страницами (раньше копипастились в каждый файл).

const Api = (function () {
    // Запрос к бэкенду. Бросает Error при не-2xx ответе; на брошенной ошибке
    // выставлен .status — код ответа, если он был получен (undefined при
    // сетевой ошибке, когда ответа не было вовсе — так вызывающий код может
    // отличить "сервер отказал" от "не достучались до сервера").
    async function api(path, method, body) {
        const options = {
            method: method,
            credentials: 'include'
        };
        if (body !== undefined) {
            options.headers = { 'Content-Type': 'application/json' };
            options.body = JSON.stringify(body);
        }
        const response = await fetch(AppConfig.getEndpoint(path), options);
        let result = null;
        try { result = await response.json(); } catch (e) { result = {}; }
        if (!response.ok) {
            const err = new Error((result && result.error) || JSON.stringify(result));
            err.status = response.status;
            throw err;
        }
        return result;
    }

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text == null ? '' : String(text);
        return div.innerHTML;
    }

    function fmtMoney(value) {
        const num = Number(value) || 0;
        return num.toFixed(2) + ' ₽';
    }

    return { api, escapeHtml, fmtMoney };
})();
