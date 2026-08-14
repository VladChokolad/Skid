// Frontend/js/config.js

const AppConfig = (function () {
    // Определяем окружение автоматически
    const hostname = window.location.hostname;
    const isLocalhost = hostname === 'localhost' ||
        hostname === '127.0.0.1' ||
        hostname === '0.0.0.0' ||
        hostname.includes('192.168.') ||
        hostname.includes('10.');

    // URL бэкенда в зависимости от окружения
    const API_URL = isLocalhost
        ? 'http://localhost:8080'
        : 'https://skid-app.ru';

    // URL фронтенда
    const FRONTEND_URL = isLocalhost
        ? 'http://localhost:5500'
        : 'https://skid-app.ru';

    // Публичный API
    return {
        API_URL,
        FRONTEND_URL,
        isLocalhost,

        // Вспомогательные методы
        getEndpoint: (path) => `${API_URL}${path}`,

        // Для отладки
        debug: () => {
            console.log('Environment:', isLocalhost ? 'LOCAL' : 'PRODUCTION');
            console.log('API URL:', API_URL);
            console.log('Frontend URL:', FRONTEND_URL);
        }
    };
})();

// Автоматически логируем окружение при загрузке
if (AppConfig.isLocalhost) {
    console.log('%c🚀 Local Development Mode', 'color: #4CAF50; font-weight: bold;');
    AppConfig.debug();
} else {
    console.log('%c🌍 Production Mode', 'color: #2196F3; font-weight: bold;');
}