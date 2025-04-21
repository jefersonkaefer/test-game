// Variáveis globais

// Funções de utilidade
function showError(message) {
    const errorDiv = $('#errorMessage');
    errorDiv.text(message).show();
    setTimeout(() => errorDiv.hide(), 3000);
}

function addMessage(message, type = 'system') {
    const chatMessages = $('#chatMessages');
    const messageDiv = $('<div>').addClass(`message ${type}`).text(message);
    chatMessages.append(messageDiv);
    chatMessages.scrollTop(chatMessages[0].scrollHeight);
}

// Conexão WebSocket
function initializeChatWebSocket() {
    if (!window.StateManager.isAuthenticated()) {
        window.location.href = '/';
        return;
    }

    window.wsManager = new WebSocketManager({
        url: `ws://localhost/api/ws?token=${window.StateManager.getToken()}`,
        onOpen: () => {
            addMessage('Conectado ao chat!', 'system');
        },
        onMessage: handleChatMessage,
        onClose: () => {
            addMessage('Desconectado do chat!', 'system');
        },
        onError: (error) => {
            addMessage('Erro na conexão do chat: ' + error.message, 'system');
        }
    });

    window.wsManager.connect();
}

function handleChatMessage(event) {
    const response = JSON.parse(event.data);
    
    switch(response.type) {
        case 'chat_message':
            const isOwnMessage = response.data.username === window.StateManager.getUsername();
            addMessage(`${response.data.username}: ${response.data.message}`, isOwnMessage ? 'self' : 'other');
            break;
            
        case 'user_joined':
            addMessage(`${response.data.username} entrou no chat`, 'system');
            break;
            
        case 'user_left':
            addMessage(`${response.data.username} saiu do chat`, 'system');
            break;
            
        case 'error':
            showError(response.message || 'Erro no chat');
            break;
            
        default:
            if (response.message) {
                addMessage(response.message, 'system');
            }
    }
}

// Handlers de eventos
$('#btnSend').click(() => {
    const message = $('#messageInput').val().trim();
    if (!message) return;

    if (window.wsManager && window.wsManager.isConnected()) {
        window.wsManager.send({
            type: 'chat_message',
            data: {
                message: message,
                username: window.StateManager.getUsername()
            }
        });
        $('#messageInput').val('');
    } else {
        showError('Não conectado ao chat');
    }
});

$('#messageInput').keypress((e) => {
    if (e.which === 13) {
        $('#btnSend').click();
    }
});

$('#btnLogout').click(() => {
    window.StateManager.clearUserData();
    if (window.wsManager) {
        window.wsManager.disconnect();
    }
    window.location.href = '/';
});

// Inicialização
$(document).ready(() => {
    if (!window.StateManager.isAuthenticated()) {
        window.location.href = '/';
        return;
    }

    $('#usernameDisplay').text(window.StateManager.getUsername());
    initializeChatWebSocket();
}); 