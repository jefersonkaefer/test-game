\c game

CREATE SCHEMA IF NOT EXISTS "public";

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TABLE IF NOT EXISTS "public"."clients" (
    "guid" UUID PRIMARY KEY,
    "username" VARCHAR(60) UNIQUE NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ
);

CREATE TRIGGER set_client_updated_at
BEFORE UPDATE ON "public"."clients"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

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

CREATE TRIGGER set_wallet_updated_at
BEFORE UPDATE ON "public"."wallets"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column(); 