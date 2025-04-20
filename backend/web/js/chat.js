let ws = null;
let token = null;
let username = null;

// Funções de utilidade
function showError(message) {
    alert(message);
}

function addMessage(message, type = 'system') {
    const chatMessages = $('#chatMessages');
    const messageDiv = $('<div>').addClass(`message ${type}`).text(message);
    chatMessages.append(messageDiv);
    chatMessages.scrollTop(chatMessages[0].scrollHeight);
}

// Conexão WebSocket
function connectWebSocket() {
    if (ws) {
        ws.close();
    }

    ws = new WebSocket(`ws://localhost/ws?token=${token}`);

    ws.onopen = () => {
        addMessage('Conectado ao servidor!', 'system');
    };

    ws.onmessage = (event) => {
        const response = JSON.parse(event.data);
        
        switch(response.type) {
            case 'welcome':
                addMessage(response.message, 'system');
                if (response.data && response.data.client_id) {
                    addMessage(`Seu ID: ${response.data.client_id}`, 'system');
                }
                break;
                
            case 'user_joined':
                addMessage(response.message, 'system');
                break;
                
            case 'response':
                if (response.data && response.data.error) {
                    addMessage(response.data.error, 'system');
                } else if (response.message) {
                    addMessage(response.message, 'system');
                }
                break;
                
            case 'error':
                addMessage(response.error || response.message, 'system');
                break;
                
            default:
                addMessage(response.message || 'Mensagem recebida', 'other');
        }
    };

    ws.onclose = () => {
        addMessage('Desconectado do servidor!', 'system');
    };

    ws.onerror = (error) => {
        addMessage('Erro na conexão: ' + error.message, 'system');
    };
}

// Handlers de eventos
$('#btnSend').click(() => {
    const message = $('#messageInput').val();
    if (!message) return;

    if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            action: 'new_player',
            body: {
                message: message
            }
        }));
        addMessage(message, 'self');
        $('#messageInput').val('');
    } else {
        showError('Não conectado ao servidor');
    }
});

$('#btnLogout').click(() => {
    localStorage.removeItem('token');
    localStorage.removeItem('username');
    if (ws) {
        ws.close();
    }
    window.location.href = '/';
});

// Inicialização
$(document).ready(() => {
    token = localStorage.getItem('token');
    username = localStorage.getItem('username');
    
    if (!token || !username) {
        window.location.href = '/';
        return;
    }

    $('#usernameDisplay').text(username);
    connectWebSocket();
}); 