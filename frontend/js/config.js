const CONFIG = {
    API_BASE_URL: 'http://localhost/api',
    WS_URL: 'ws://localhost/api/ws',
    ENDPOINTS: {
        login: '/login',
        register: '/register',
        logout: '/logout',
        wallet: '/wallet'
    }
};

// Log das configurações para depuração
console.log("Config inicializada:", {
    "API_BASE_URL": CONFIG.API_BASE_URL,
    "WS_URL": CONFIG.WS_URL
});

// Gerenciamento de estado global
window.StateManager = {
    setUserData(token, username) {
        localStorage.setItem('token', token);
        localStorage.setItem('username', username);
    },

    clearUserData() {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        if (window.wsManager) {
            window.wsManager.disconnect();
        }
    },

    getToken() {
        return localStorage.getItem('token');
    },

    getUsername() {
        return localStorage.getItem('username');
    },

    isAuthenticated() {
        return !!this.getToken();
    }
}; 