// Frontend/js/session.js — единое определение "кто я" (зарегистрированный
// пользователь / анонимный гость), используется вместо разрозненных
// checkAuth/loadMe, которые раньше были в каждом файле по-своему.
// Зависит от Api (js/api.js) — подключать после него.

const Session = (function () {
    // Возвращает { id, isAnon } или null, если не удалось определить
    // (например, бэкенд недоступен). У анонимов в ответе /profile нет
    // поля email — так отличаем гостя от зарегистрированного пользователя.
    async function loadMe() {
        try {
            const result = await Api.api('/profile', 'GET');
            const data = result.data || result;
            const isAnon = !('email' in data);
            return { id: data.id, isAnon: isAnon, raw: data };
        } catch (e) {
            return null;
        }
    }

    return { loadMe };
})();
