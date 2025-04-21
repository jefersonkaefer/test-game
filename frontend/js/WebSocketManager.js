class WebSocketManager {
    constructor(config) {
        this.token = config.token;
        this.onOpen = config.onOpen || (() => {});
        this.onClose = config.onClose || (() => {});
        this.onError = config.onError || (() => {});
        this.onMessage = config.onMessage || (() => {});
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000; // 1 segundo
        
        if (config.autoConnect !== false) {
            this.connect();
        }
    }

    connect() {
        if (this.ws) {
            this.ws.close();
        }

        try {
            // Formato correto: authorization=Bearer token (sem aspas extras)
            const wsUrl = `${CONFIG.WS_URL}?authorization=Bearer ${encodeURIComponent(this.token)}`;
            console.log('Conectando ao WebSocket:', wsUrl);
            
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                console.log('WebSocket conectado com sucesso!');
                this.reconnectAttempts = 0;
                this.onOpen();
            };

            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.onMessage(data);
                } catch (error) {
                    console.error('Erro ao processar mensagem:', error);
                }
            };

            this.ws.onclose = (event) => {
                console.log('WebSocket desconectado', event.code);
                this.onClose();
                this.tryReconnect();
            };

            this.ws.onerror = (error) => {
                console.error('Erro no WebSocket:', error);
                this.onError(error);
            };
        } catch (error) {
            console.error('Erro ao criar conexão WebSocket:', error);
            this.onError(error);
        }
    }

    tryReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.log('Número máximo de tentativas de reconexão atingido');
            return;
        }

        this.reconnectAttempts++;
        console.log(`Tentativa de reconexão ${this.reconnectAttempts} de ${this.maxReconnectAttempts}`);
        
        setTimeout(() => {
            this.connect();
        }, this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1));
    }

    send(data) {
        if (!this.isConnected()) {
            console.error('WebSocket não está conectado');
            return false;
        }

        try {
            this.ws.send(JSON.stringify(data));
            return true;
        } catch (error) {
            console.error('Erro ao enviar mensagem:', error);
            return false;
        }
    }

    isConnected() {
        return this.ws && this.ws.readyState === WebSocket.OPEN;
    }

    disconnect() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
    }
} 