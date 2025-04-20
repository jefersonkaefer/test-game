let ws = null;

// Funções de utilidade
function showError(message) {
    alert(message);
}

function checkAuth() {
    const storedToken = localStorage.getItem('token');
    if (!storedToken) {
        console.log('Nenhum token encontrado, redirecionando para login...');
        window.location.href = '/';
        return false;
    }
    return true;
}

$(document).ready(() => {
    const API_URL = 'http://localhost:3000';

    $('#loginForm').on('submit', function(e) {
        e.preventDefault();
        
        const username = $('#username').val().trim();
        
        if (!username) {
            alert('Por favor, insira um nome de usuário');
            return;
        }

        $.ajax({
            url: `${API_URL}/login`,
            method: 'POST',
            data: JSON.stringify({ username }),
            contentType: 'application/json',
            success: function(response) {
                localStorage.setItem('username', username);
                window.location.href = 'game.html';
            },
            error: function(xhr) {
                alert('Erro ao fazer login. Por favor, tente novamente.');
                console.error('Erro:', xhr);
            }
        });
    });

    // Verifica se o usuário está logado ao carregar a página
    const username = localStorage.getItem('username');
    if (!username && window.location.pathname.includes('game.html')) {
        window.location.href = 'index.html';
    }

    // Verifica se existe um token
    if (!checkAuth()) {
        return;
    }

    const token = localStorage.getItem('token');
    
    // Inicializa o WebSocket
    ws = new WebSocket(`ws://localhost/ws?token=${token}`);

    ws.onopen = () => {
        console.log('Conectado ao servidor WebSocket!');
    };

    ws.onmessage = (event) => {
        const response = JSON.parse(event.data);
        console.log('Mensagem recebida:', response);
    };

    ws.onclose = (event) => {
        console.log('Desconectado do servidor WebSocket');
        if (event.code === 1006 || event.code === 4001) { // 1006 é o código para conexão fechada anormalmente
            console.log('Conexão fechada devido a erro de autenticação');
            localStorage.removeItem('token');
            localStorage.removeItem('username');
            window.location.href = '/';
        }
    };

    ws.onerror = (error) => {
        console.error('Erro na conexão WebSocket:', error);
        if (error.target.readyState === WebSocket.CLOSED) {
            localStorage.removeItem('token');
            localStorage.removeItem('username');
            window.location.href = '/';
        }
    };

    // Handler do botão de logout
    $('#logout').on('click', function() {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        window.location.href = '/';
    });

    // Handler do botão de iniciar jogo
    $('#startGame').on('click', function() {
        if (!checkAuth()) {
            return;
        }

        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({
                action: 'start_game'
            }));
        } else {
            showError('Não conectado ao servidor');
        }
    });
});

// Função para iniciar o jogo
$('#startGame').click(function() {
    const username = localStorage.getItem('username');
    if (!username) {
        alert('Por favor, faça login primeiro.');
        window.location.href = 'index.html';
        return;
    }

    // Aqui você pode adicionar a lógica para iniciar o jogo
    alert('Jogo iniciado!');
});

// Função para fazer logout
$('#logout').click(function() {
    localStorage.removeItem('username');
    window.location.href = 'index.html';
}); 