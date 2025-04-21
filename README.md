# Jogo Ímpar ou Par

Este é um jogo de apostas online onde os jogadores podem apostar se um número sorteado será ímpar ou par. O projeto é composto por um backend em Go com uma arquitetura baseada em domínios e um frontend em HTML, CSS e JavaScript.

## Tecnologias Utilizadas

### Backend
- Go (Golang)
- PostgreSQL
- Redis
- JWT para autenticação
- Arquitetura em camadas (Domain-Driven Design)
- WebSockets para comunicação em tempo real

### Frontend
- HTML5
- CSS3
- JavaScript
- jQuery
- WebSockets para comunicação em tempo real

## Requisitos

- Docker
- Docker Compose

## Como Iniciar o Projeto

1. Clone o repositório
2. Crie um arquivo `.env` baseado no arquivo `sample.env`:
   ```
   cp sample.env .env
   ```
3. Inicie os contêineres usando Docker Compose:
   ```
   docker-compose up -d
   ```
4. Acesse o jogo em seu navegador através do endereço: http://localhost

## Arquitetura do Projeto

O projeto segue uma arquitetura em camadas, inspirada em Domain-Driven Design (DDD):

- **domain**: Contém as regras de negócio e entidades principais
- **application**: Contém os casos de uso da aplicação
- **infra**: Implementação técnica (banco de dados, rede, etc.)

## Serviços

- **API**: Serviço backend em Go (porta 8000)
- **NGINX**: Servidor web para servir os arquivos estáticos do frontend (porta 80)
- **Postgres**: Banco de dados relacional (porta 5432)
- **Redis**: Armazenamento em memória usado para sessões (porta 6379)
- **Adminer**: Interface web para gerenciar o banco de dados (porta 8080)
- **RedisInsight**: Interface web para gerenciar o Redis (porta 5540)

## Endpoints da API

### REST API

- **POST /register**: Registra um novo usuário
  - Body: `{ "username": "string", "password": "string" }`
  - Response: `{ "id": "uuid" }`

- **POST /login**: Autentica um usuário
  - Body: `{ "username": "string", "password": "string" }`
  - Response: `{ "token": "jwt-token" }`

- **POST /logout**: Encerra a sessão do usuário (requer autenticação)
  - Headers: `Authorization: Bearer <token>`

- **GET /wallet**: Obtém o saldo do usuário (requer autenticação)
  - Headers: `Authorization: Bearer <token>`
  - Response: `{ "balance": float }`

### WebSocket API (requer autenticação)

- **GET /ws**: Endpoint WebSocket para comunicação em tempo real
  - Headers: `Authorization: Bearer <token>`
  - Parameters: `/ws?authorization=Bearer <token>`

#### Ações WebSocket:

1. **new_match**: Inicia uma nova partida
   - Request: `{ "action": "new_match" }`
   - Response: `{ "action": "new_match", "data": null }`

2. **place_bet**: Realiza uma aposta
   - Request: `{ "action": "place_bet", "data": { "amount": float, "choice": "odd|even" } }`
   - Response: `{ "action": "place_bet", "data": { "result": "win|lose", "number": int } }`

3. **wallet**: Consulta o saldo
   - Request: `{ "action": "wallet" }`
   - Response: `{ "action": "wallet", "data": { "balance": float } }`

4. **end_match**: Finaliza a partida atual
   - Request: `{ "action": "end_match" }`
   - Response: `{ "action": "end_match", "data": null }`

## Fluxo do Jogo

1. Usuário se registra ou faz login
2. Usuário inicia uma nova partida
3. Usuário escolhe entre ímpar ou par e faz sua aposta
4. O sistema sorteia um número aleatório
5. Se o número for compatível com a escolha do usuário (ímpar ou par), ele ganha o dobro do valor apostado
6. O usuário pode fazer novas apostas ou encerrar a partida

## Ferramentas de Administração

- **Adminer**: Acesse http://localhost:8080 para gerenciar o banco de dados
  - Sistema: PostgreSQL
  - Servidor: db
  - Usuário: postgres
  - Senha: t3st (ou o valor definido em .env)
  - Banco de dados: game

- **RedisInsight**: Acesse http://localhost:5540 para gerenciar o Redis
  - Host: redis
  - Porta: 6379
  - Senha: t3st (ou o valor definido em .env)

## Notas de Desenvolvimento

- O projeto utiliza WebSockets para comunicação em tempo real entre o cliente e o servidor
- A autenticação é baseada em tokens JWT com expiração
- Cada usuário começa com um saldo padrão em sua carteira
- As sessões são armazenadas no Redis com um tempo de vida configurável 
