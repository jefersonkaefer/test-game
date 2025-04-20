# Jogo de Dados - Backend

Este é um backend para um jogo de dados simples onde os jogadores podem apostar em números pares ou ímpares.

## Requisitos

- Go 1.21 ou superior
- Dependências listadas no `go.mod`

## Instalação

1. Clone o repositório
2. Execute `go mod download` para baixar as dependências
3. Execute `go run main.go` para iniciar o servidor

## API REST

### Autenticação

1. Registrar novo usuário
```http
POST /register
Content-Type: application/json

{
    "username": "seu_usuario",
    "password": "sua_senha"
}
```

2. Login
```http
POST /login
Content-Type: application/json

{
    "username": "seu_usuario",
    "password": "sua_senha"
}
```

Resposta:
```json
{
    "token": "seu_token_jwt"
}
```

## API WebSocket

O servidor WebSocket está disponível em `ws://localhost:8080/ws/{token}`

Onde `{token}` é o token JWT obtido no login.

### Mensagens suportadas

1. Consultar Saldo (Wallet)
```json
{
    "type": "wallet"
}
```

2. Fazer Aposta (Play)
```json
{
    "type": "play",
    "payload": {
        "amount": 100.0,
        "betType": "even" // ou "odd"
    }
}
```

3. Finalizar Jogo (EndPlay)
```json
{
    "type": "endPlay"
}
```

### Respostas

Todas as respostas seguem o formato:
```json
{
    "type": "tipo-da-resposta",
    "success": true/false,
    "data": { ... },
    "error": "mensagem de erro, se houver"
}
```

## Testes

Para testar a aplicação:

1. Registre um novo usuário usando o endpoint `/register`
2. Faça login usando o endpoint `/login` para obter o token JWT
3. Conecte ao WebSocket em `ws://localhost:8080/ws/seu_token_jwt`
4. Envie uma mensagem para consultar o saldo
5. Faça uma aposta
6. Finalize o jogo para receber os ganhos (se houver)

## Observações

- Cada jogador começa com 1000.0 de saldo
- As apostas devem ser menores ou iguais ao saldo disponível
- Não é possível iniciar um novo jogo sem finalizar o anterior
- Em caso de vitória, o jogador recebe o dobro da aposta
- O token JWT expira em 24 horas
- Todas as rotas da API (exceto registro e login) requerem autenticação 