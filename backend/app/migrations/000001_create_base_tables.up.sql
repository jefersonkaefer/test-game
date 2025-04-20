-- Função para atualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Tabela de clientes
CREATE TABLE IF NOT EXISTS "public"."clients" (
    "guid" UUID PRIMARY KEY,
    "username" VARCHAR(60) UNIQUE NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ
);

CREATE TRIGGER IF NOT EXISTS set_client_updated_at
BEFORE UPDATE ON "public"."clients"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Tabela de carteiras
CREATE TABLE IF NOT EXISTS "public"."wallets" (
    "guid" UUID PRIMARY KEY,
    "client_id" UUID NOT NULL,
    "balance" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ,
    CONSTRAINT fk_client FOREIGN KEY (client_id) REFERENCES clients(guid)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE TRIGGER IF NOT EXISTS set_wallet_updated_at
BEFORE UPDATE ON "public"."wallets"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Tabela de partidas
CREATE TABLE IF NOT EXISTS "public"."matches" (
    "guid" UUID PRIMARY KEY,
    "status" VARCHAR(20) NOT NULL,
    "min_players" INTEGER NOT NULL,
    "max_players" INTEGER NOT NULL,
    "game_mode" VARCHAR(20) NOT NULL,
    "bets" JSONB DEFAULT '{}',
    "choices" JSONB DEFAULT '{}',
    "result" VARCHAR(20),
    "current_turn" VARCHAR(36),
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ
);

CREATE TRIGGER IF NOT EXISTS set_match_updated_at
BEFORE UPDATE ON "public"."matches"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Tabela de jogadores da partida
CREATE TABLE IF NOT EXISTS "public"."match_players" (
    "guid" UUID PRIMARY KEY,
    "match_id" UUID NOT NULL,
    "player_id" UUID NOT NULL,
    "role" VARCHAR(20) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ,
    CONSTRAINT fk_match FOREIGN KEY (match_id) REFERENCES matches(guid)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT fk_client FOREIGN KEY (player_id) REFERENCES clients(guid)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE TRIGGER IF NOT EXISTS set_match_player_updated_at
BEFORE UPDATE ON "public"."match_players"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column(); 