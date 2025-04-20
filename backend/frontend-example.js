// Configuração da API
const API_BASE_URL = 'http://localhost/api';
const WS_BASE_URL = 'ws://localhost/ws';

// Configuração do Axios para requisições HTTP
import axios from 'axios';

const api = axios.create({
    baseURL: API_BASE_URL,
    headers: {
        'Content-Type': 'application/json',
    },
    withCredentials: true
});

// Interceptor para adicionar token JWT
api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// Configuração do WebSocket
const ws = new WebSocket(WS_BASE_URL);

ws.onopen = () => {
    console.log('WebSocket conectado');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    console.log('Mensagem recebida:', data);
};

ws.onerror = (error) => {
    console.error('Erro no WebSocket:', error);
};

ws.onclose = () => {
    console.log('WebSocket desconectado');
};

// Função para enviar mensagem via WebSocket
function sendWebSocketMessage(action, body) {
    if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            action,
            body
        }));
    } else {
        console.error('WebSocket não está conectado');
    }
}

// Exemplo de uso
export async function login(username, password) {
    try {
        const response = await api.post('/login', {
            username,
            password
        });
        localStorage.setItem('token', response.data.token);
        return response.data;
    } catch (error) {
        console.error('Erro no login:', error);
        throw error;
    }
}

export async function createNewClient(username, password) {
    try {
        const response = await api.post('/client', {
            username,
            password
        });
        return response.data;
    } catch (error) {
        console.error('Erro ao criar cliente:', error);
        throw error;
    }
}

export function createNewGame() {
    sendWebSocketMessage('new_game', {});
}

export function createNewMatch() {
    sendWebSocketMessage('new_match', {});
}

export default {
    login,
    createNewClient,
    createNewGame,
    createNewMatch
}; 