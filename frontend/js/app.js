// Funções de utilidade
function showError(message) {
    $('#errorModal .modal-body').text(message);
    $('#errorModal').modal('show');
}

function showSuccess(message) {
    $('#successModal .modal-body').text(message);
    $('#successModal').modal('show');
}

function addMessage(message, type = 'system') {
    const chatMessages = $('#chatMessages');
    const messageDiv = $('<div>').addClass(`message ${type}`).text(message);
    chatMessages.append(messageDiv);
    chatMessages.scrollTop(chatMessages[0].scrollHeight);
}

// Handlers de eventos
$(document).ready(() => {
    console.log('Documento carregado');
    
    // Verifica autenticação e redireciona se necessário
    const isLoginPage = window.location.pathname === '/' || window.location.pathname === '/index.html';
    const isGamePage = window.location.pathname === '/game.html';
    
    if (isLoginPage && window.StateManager.isAuthenticated()) {
        window.location.href = '/game.html';
        return;
    }
    
    if (isGamePage && !window.StateManager.isAuthenticated()) {
        window.location.href = '/';
        return;
    }
    
    // Setup da página de login
    if (isLoginPage) {
        setupLoginPage();
    }
    
    // Setup da página do jogo
    if (isGamePage) {
        setupGamePage();
    }
});

function setupLoginPage() {
    // Gerenciamento de abas
    $('#tabLogin').on('click', function() {
        $('#tabLogin').addClass('active');
        $('#tabRegister').removeClass('active');
        $('#loginForm').removeClass('hidden');
        $('#registerForm').addClass('hidden');
    });

    $('#tabRegister').on('click', function() {
        $('#tabRegister').addClass('active');
        $('#tabLogin').removeClass('active');
        $('#registerForm').removeClass('hidden');
        $('#loginForm').addClass('hidden');
    });
    
    // Handler do formulário de login
    $('#btnLogin').on('click', handleLogin);
    
    // Handler do formulário de registro
    $('#btnRegister').on('click', handleRegister);
}

function setupGamePage() {
    // Inicializa o WebSocket
    initializeWebSocket();
    
    // Setup dos handlers do jogo
    $('#btnLogout').on('click', handleLogout);
    $('#btnStartGame').on('click', handleStartGame);
    $('#btnSend').on('click', handleSendMessage);
    
    // Atualiza a interface
    $('#welcomeMessage').text(`Bem-vindo, ${window.StateManager.getUsername()}!`);
}

function handleLogin() {
    const username = $('#usernameInput').val().trim();
    const password = $('#passwordInput').val().trim();
    
    if (!username || !password) {
        showError('Por favor, preencha todos os campos');
        return;
    }

    $.ajax({
        url: CONFIG.API_BASE_URL + CONFIG.ENDPOINTS.login,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ username, password }),
        success: (response) => {
            if (response && response.token) {
                window.StateManager.setUserData(response.token, username);
                window.location.href = '/game.html';
            } else {
                showError('Erro ao fazer login: Token não recebido');
            }
        },
        error: (xhr) => {
            const errorMessage = xhr.responseJSON?.message || xhr.responseJSON?.error || 'Erro ao fazer login';
            showError(errorMessage);
        }
    });
}

function handleRegister() {
    const username = $('#registerUsername').val().trim();
    const password = $('#registerPassword').val();
    const confirmPassword = $('#confirmPassword').val();
    
    if (!username || !password || !confirmPassword) {
        showError('Por favor, preencha todos os campos');
        return;
    }

    if (password !== confirmPassword) {
        showError('As senhas não coincidem');
        return;
    }

    if (username.length < 3) {
        showError('O nome de usuário deve ter pelo menos 3 caracteres');
        return;
    }

    if (password.length < 6) {
        showError('A senha deve ter pelo menos 6 caracteres');
        return;
    }

    $.ajax({
        url: CONFIG.API_BASE_URL + CONFIG.ENDPOINTS.register,
        method: 'POST',
        contentType: 'application/json',
        data: JSON.stringify({ username, password }),
        success: (response) => {
            if (response && response.success) {
                showSuccess('Registro realizado com sucesso! Faça login para continuar.');
                $('#registerUsername, #registerPassword, #confirmPassword').val('');
                $('#tabLogin').click();
            } else {
                showError('Erro ao realizar registro: ' + (response.message || 'Erro desconhecido'));
            }
        },
        error: (xhr) => {
            const errorMessage = xhr.responseJSON?.message || xhr.responseJSON?.error || 'Erro ao realizar registro';
            showError(errorMessage);
        }
    });
}

function initializeWebSocket() {
    if (!window.StateManager.isAuthenticated()) return;
    
    window.wsManager = new WebSocketManager({
        token: window.StateManager.getToken(),
        onOpen: () => {
            console.log('Conectado ao servidor WebSocket!');
            $('#connectionStatus').text('Conectado').removeClass('text-danger').addClass('text-success');
        },
        onClose: () => {
            console.log('Desconectado do servidor WebSocket');
            $('#connectionStatus').text('Desconectado').removeClass('text-success').addClass('text-danger');
        },
        onError: (error) => {
            console.error('Erro na conexão WebSocket:', error);
            showError('Erro na conexão com o servidor');
        },
        onMessage: handleWebSocketMessage
    });
}

function handleWebSocketMessage(data) {
    if (!data) return;
    
    switch (data.type) {
        case 'chat_message':
            addMessage(data.message, data.username === window.StateManager.getUsername() ? 'self' : 'other');
            break;
        case 'player_joined':
            addMessage(`${data.username} entrou no jogo`, 'system');
            updatePlayerList(data.players);
            break;
        case 'player_left':
            addMessage(`${data.username} saiu do jogo`, 'system');
            updatePlayerList(data.players);
            break;
        case 'game_status':
            updateGameStatus(data);
            break;
        case 'game_error':
            showError(data.message);
            break;
        default:
            console.log('Mensagem não tratada:', data);
    }
}

function updatePlayerList(players) {
    if (!players) return;
    
    const playerList = $('#playerList');
    playerList.empty();
    players.forEach(player => {
        const playerItem = $('<li>')
            .addClass('player-item')
            .text(player.username);
        if (player.ready) {
            playerItem.addClass('ready');
        }
        playerList.append(playerItem);
    });
}

function updateGameStatus(data) {
    if (data.gameState) {
        $('#gameState').text(data.gameState);
        $('#gameContainer').attr('data-state', data.gameState.toLowerCase());
    }
    
    if (data.players) {
        updatePlayerList(data.players);
    }
    
    if (data.currentTurn) {
        $('#currentTurn').text(`Turno: ${data.currentTurn}`);
    }
}

function handleStartGame() {
    if (!window.wsManager || !window.wsManager.isConnected()) {
        showError('Não foi possível iniciar o jogo: Sem conexão com o servidor');
        return;
    }
    
    window.wsManager.send({
        type: 'start_game',
        username: window.StateManager.getUsername()
    });
}

function handleSendMessage() {
    const messageInput = $('#messageInput');
    const message = messageInput.val().trim();
    
    if (!message) return;
    
    if (!window.wsManager || !window.wsManager.isConnected()) {
        showError('Não foi possível enviar a mensagem: Sem conexão com o servidor');
        return;
    }
    
    window.wsManager.send({
        type: 'chat_message',
        message: message,
        username: window.StateManager.getUsername()
    });
    
    messageInput.val('');
}

function handleLogout() {
    window.StateManager.clearUserData();
    window.location.href = '/';
} 