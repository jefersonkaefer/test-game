CREATE TABLE "public"."wallets" (
    guid UUID PRIMARY KEY,
    balance DOUBLE PRECISION NOT NULL DEFAULT 0,
    client_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_client FOREIGN KEY (client_id) REFERENCES clients(guid)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);
CREATE TRIGGER set_wallet_updated_at
BEFORE UPDATE ON  "public"."wallets"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();