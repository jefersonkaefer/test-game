let currentChoice = null;
let wsManager = null;
let currentBalance = 0;

function showError(message) {
    console.error('Erro:', message);
    const errorModal = $('#errorModal');
    const errorMessage = $('#errorMessage');
    const modalFooter = $('.modal-footer');
    
    errorMessage.text(message);
    errorMessage.addClass('show');
    errorModal.removeClass('hidden');
    
    // Limpa o footer do modal
    modalFooter.empty();
    
    // Se o erro for "player already in match", adiciona o botão de finalizar partida
    if (message === 'player already in match') {
        modalFooter.append(`
            <button class="btn-primary" id="endMatchFromModal">
                <i class="fas fa-stop-circle"></i>
                Finalizar Partida Atual
            </button>
        `);
        
        // Adiciona o handler para o botão
        $('#endMatchFromModal').click(() => {
            showLoading();
            wsManager.send({ action: 'end_match' });
            errorModal.addClass('hidden');
        });
    } else {
        // Para outros erros, adiciona apenas o botão OK
        modalFooter.append(`
            <button class="btn-primary modal-close">OK</button>
        `);
    }
    
    // Foca no modal para garantir que ele seja exibido
    errorModal.focus();
}

function showLoading(show = true) {
    $('#loadingOverlay').toggleClass('hidden', !show);
}

function updateBalance(balance) {
    currentBalance = parseFloat(balance);
    $('#balance').text(`R$ ${currentBalance.toFixed(2)}`);
    
    const betInput = $('#betAmount');
    const currentBet = parseFloat(betInput.val()) || 0;
    if (currentBet > currentBalance) {
        betInput.val(currentBalance);
    }
}

function updateResultIcon(result) {
    const icon = $('.result-icon i');
    icon.removeClass('fa-question fa-check fa-times');
    
    if (result === 'win') {
        icon.addClass('fa-check text-success');
    } else {
        icon.addClass('fa-times text-danger');
    }
}

function handleGameMessage(data) {
    console.log('Mensagem recebida:', data);
    showLoading(false);
    
    if (data.error) {
        showError(data.error);
        wsManager.send({ action: 'wallet' });
        return;
    }
    
    switch (data.action) {
        case 'wallet':
            updateBalance(data.data.balance);
            break;
            
        case 'new_match':
            $('#waitingArea').addClass('hidden');
            $('#gamePlayArea').removeClass('hidden');
            $('#gameResult').addClass('hidden');
            currentChoice = null;
            $('.btn-choice').removeClass('active');
            wsManager.send({ action: 'wallet' });
            break;
            
        case 'place_bet':
            $('#gamePlayArea').addClass('hidden');
            $('#gameResult').removeClass('hidden');
            
            $('#resultNumber').text(data.data.number);
            
            const isEven = data.data.number % 2 === 0;
            const won = (isEven && currentChoice === 'even') || (!isEven && currentChoice === 'odd');
            
            updateResultIcon(data.data.result);
            const resultMessage = data.data.result === 'win' ? 
                'Parabéns! Você ganhou!' :
                'Que pena, você perdeu. Tente novamente!';
            
            $('#resultMessage').text(resultMessage);
            wsManager.send({ action: 'wallet' });
            break;
            
        case 'end_match':
            $('#gamePlayArea').addClass('hidden');
            $('#gameResult').addClass('hidden');
            $('#waitingArea').removeClass('hidden');
            
            currentChoice = null;
            $('.btn-choice').removeClass('active');
            $('#resultNumber').text('?');
            $('.result-icon i').removeClass('fa-check fa-times').addClass('fa-question');
            
            wsManager.send({ action: 'wallet' });
            break;
    }
}

function initializeWebSocket() {
    if (!window.StateManager.isAuthenticated()) return;
    
    console.log('Inicializando WebSocket...');
    console.log('Token:', window.StateManager.getToken().substring(0, 10) + '...');
    
    wsManager = new WebSocketManager({
        token: window.StateManager.getToken(),
        onOpen: () => {
            console.log('Conectado ao servidor WebSocket');
            wsManager.send({ action: 'wallet' });
        },
        onMessage: (data) => {
            console.log('Mensagem recebida do servidor:', data);
            handleGameMessage(data);
        },
        onClose: () => {
            console.log('Desconectado do servidor WebSocket');
            showError('Conexão perdida. Por favor, recarregue a página.');
        },
        onError: (error) => {
            console.error('Erro no WebSocket:', error);
            showError('Erro na conexão com o servidor. Verifique o console para mais detalhes.');
        }
    });
    
    // Adiciona uma verificação para alertar se não conectar em 5 segundos
    setTimeout(() => {
        if (!wsManager.isConnected()) {
            console.error('Não foi possível estabelecer conexão WebSocket após 5 segundos');
            showError('Falha na conexão WebSocket. Verifique se o servidor está acessível.');
        }
    }, 5000);
}

$(document).ready(function() {
    if (!window.StateManager || !window.StateManager.isAuthenticated()) {
        window.location.href = '/index.html';
        return;
    }

    $('#username').text(window.StateManager.getUsername());
    initializeWebSocket();

    $('#newMatch').click(() => {
        showLoading();
        wsManager.send({ action: 'new_match' });
    });

    $('.btn-choice').click(function() {
        $('.btn-choice').removeClass('active');
        $(this).addClass('active');
        currentChoice = $(this).data('choice');
    });

    $('.btn-amount').click(function() {
        const input = $('#betAmount');
        const currentValue = parseInt(input.val()) || 0;
        const action = $(this).data('action');
        
        if (action === 'increase') {
            const newValue = Math.min(currentValue + 10, currentBalance);
            input.val(newValue);
        } else {
            input.val(Math.max(10, currentValue - 10));
        }
    });

    $('#betAmount').on('input', function() {
        const input = $(this);
        let value = parseFloat(input.val()) || 0;
        value = Math.max(10, Math.min(value, currentBalance));
        input.val(value);
    });

    $('#placeBet').click(() => {
        if (!currentChoice) {
            showError('Por favor, escolha ímpar ou par');
            return;
        }

        const amount = parseFloat($('#betAmount').val());
        if (isNaN(amount) || amount < 10) {
            showError('Por favor, insira um valor válido para a aposta (mínimo R$ 10)');
            return;
        }

        if (amount > currentBalance) {
            showError('Valor da aposta maior que o saldo disponível');
            return;
        }

        showLoading();
        wsManager.send({
            action: 'place_bet',
            data: {
                amount: amount,
                choice: currentChoice
            }
        });
    });

    $('#playAgain').click(() => {
        $('#gameResult').addClass('hidden');
        $('#gamePlayArea').removeClass('hidden');
        $('.btn-choice').removeClass('active');
        currentChoice = null;
        wsManager.send({ action: 'wallet' });
    });

    $('#endMatch, #endMatchGame').click(() => {
        showLoading();
        wsManager.send({ action: 'end_match' });
    });

    $('#logout').click(() => {
        showLoading();
        $.ajax({
            url: CONFIG.API_BASE_URL + CONFIG.ENDPOINTS.logout,
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${window.StateManager.getToken()}`
            },
            success: () => {
                window.StateManager.clearUserData();
                window.location.href = '/index.html';
            },
            error: () => {
                window.StateManager.clearUserData();
                window.location.href = '/index.html';
            }
        });
    });

    $('.modal-close').click(function() {
        $(this).closest('.modal').addClass('hidden');
    });
}); 