function showError(message) {
    $('#errorMessage').text(message);
    $('#errorModal').removeClass('hidden');
}

function showLoading(show = true) {
    $('#loadingOverlay').toggleClass('hidden', !show);
}

$(document).ready(function() {
    // Se já estiver autenticado, redireciona para o jogo
    if (window.StateManager && window.StateManager.isAuthenticated()) {
        window.location.href = '/game.html';
        return;
    }

    // Gerenciamento de abas
    $('#tabLogin').click(function() {
        $('#tabLogin').addClass('active');
        $('#tabRegister').removeClass('active');
        $('#loginForm').removeClass('hidden');
        $('#registerForm').addClass('hidden');
    });

    $('#tabRegister').click(function() {
        $('#tabRegister').addClass('active');
        $('#tabLogin').removeClass('active');
        $('#registerForm').removeClass('hidden');
        $('#loginForm').addClass('hidden');
    });

    // Handler do login
    $('#btnLogin').click(function() {
        const username = $('#usernameInput').val().trim();
        const password = $('#passwordInput').val().trim();

        if (!username || !password) {
            showError('Por favor, preencha todos os campos');
            return;
        }

        showLoading(true);
        $.ajax({
            url: CONFIG.API_BASE_URL + CONFIG.ENDPOINTS.login,
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ username, password }),
            success: function(response) {
                if (response && response.token) {
                    window.StateManager.setUserData(response.token, username);
                    window.location.href = '/game.html';
                } else {
                    showError('Erro ao fazer login: Token não recebido');
                }
            },
            error: function(xhr) {
                showLoading(false);
                const errorMessage = xhr.responseJSON?.message || xhr.responseJSON?.error || 'Erro ao fazer login';
                showError(errorMessage);
            }
        });
    });

    // Handler do registro
    $('#btnRegister').click(function() {
        const username = $('#registerUsername').val().trim();
        const password = $('#registerPassword').val().trim();
        const confirmPassword = $('#confirmPassword').val().trim();

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

        showLoading(true);
        $.ajax({
            url: CONFIG.API_BASE_URL + CONFIG.ENDPOINTS.register,
            method: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({ username, password }),
            success: function(response) {
                showLoading(false);
                if (response && response.id) {
                    $('#registerUsername, #registerPassword, #confirmPassword').val('');
                    $('#tabLogin').click();
                    showSuccess('Registro realizado com sucesso! Faça login para continuar.');
                } else {
                    showError('Erro ao realizar registro: ' + (response.message || 'Erro desconhecido'));
                }
            },
            error: function(xhr) {
                showLoading(false);
                const errorMessage = xhr.responseJSON?.message || xhr.responseJSON?.error || 'Erro ao realizar registro';
                showError(errorMessage);
            }
        });
    });

    // Handler do modal de erro
    $('.modal-close').click(function() {
        $(this).closest('.modal').addClass('hidden');
    });
}); 