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
        console.log('Conectado ao servidor WebSocket!');
    };

    ws.onmessage = (event) => {
        const response = JSON.parse(event.data);
        console.log('Mensagem recebida:', response);
    };

    ws.onclose = () => {
        console.log('Desconectado do servidor WebSocket');
    };

    ws.onerror = (error) => {
        console.error('Erro na conexão WebSocket:', error);
    };
}

// Handlers de eventos
$(document).ready(() => {
    console.log('Documento carregado');
    
    // Verifica se já existe um token
    const storedToken = localStorage.getItem('token');
    if (storedToken) {
        console.log('Token encontrado, redirecionando para o jogo...');
        window.location.href = '/game.html';
        return;
    }
    
    // Gerenciamento de abas
    $('#tabLogin').on('click', function() {
        console.log('Aba de login clicada');
        $('#tabLogin').addClass('active');
        $('#tabRegister').removeClass('active');
        $('#loginForm').removeClass('hidden');
        $('#registerForm').addClass('hidden');
    });

    $('#tabRegister').on('click', function() {
        console.log('Aba de registro clicada');
        $('#tabRegister').addClass('active');
        $('#tabLogin').removeClass('active');
        $('#registerForm').removeClass('hidden');
        $('#loginForm').addClass('hidden');
    });
    
    $('#btnLogin').on('click', function() {
        console.log('Botão de login clicado');
        const username = $('#usernameInput').val();
        const password = $('#passwordInput').val();
        
        console.log('Username:', username);
        console.log('Password:', password);
        
        if (!username || !password) {
            showError('Por favor, preencha todos os campos');
            return;
        }

        $.ajax({
            url: '/api/login',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ username, password }),
            success: (response) => {
                console.log('Resposta do servidor:', response);
                if (response.data && response.data.token) {
                    token = response.data.token;
                    localStorage.setItem('token', token);
                    localStorage.setItem('username', username);
                    connectWebSocket();
                    window.location.href = '/game.html';
                } else {
                    showError('Erro ao fazer login: Token não recebido');
                }
            },
            error: (xhr) => {
                console.error('Erro na requisição:', xhr);
                showError(xhr.responseJSON?.error || 'Erro ao fazer login');
            }
        });
    });

    $('#btnRegister').on('click', function() {
        console.log('Botão de registro clicado');
        const username = $('#registerUsername').val();
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

        $.ajax({
            url: '/api/register',
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ username, password }),
            success: (response) => {
                console.log('Resposta do servidor:', response);
                if (response.success) {
                    showError('Registro realizado com sucesso! Faça login para continuar.');
                    $('#tabLogin').click();
                } else {
                    showError('Erro ao realizar registro');
                }
            },
            error: (xhr) => {
                console.error('Erro na requisição:', xhr);
                showError(xhr.responseJSON?.error || 'Erro ao realizar registro');
            }
        });
    });

    if (localStorage.getItem('token')) {
        connectWebSocket();
    }
});

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