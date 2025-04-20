CREATE TABLE "public"."games" (
    "guid" UUID PRIMARY KEY,
    "name" VARCHAR(100) NOT NULL,
    "description" TEXT,
    "min_players" INTEGER NOT NULL,
    "max_players" INTEGER NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ
);

CREATE TRIGGER set_game_updated_at
BEFORE UPDATE ON "public"."games"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE "public"."matches" (
    "guid" UUID PRIMARY KEY,
    "game_id" UUID NOT NULL,
    "status" VARCHAR(20) NOT NULL,
    "min_players" INTEGER NOT NULL,
    "max_players" INTEGER NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ,
    CONSTRAINT fk_game FOREIGN KEY (game_id) REFERENCES games(guid)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE TRIGGER set_match_updated_at
BEFORE UPDATE ON "public"."matches"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE "public"."match_players" (
    "guid" UUID PRIMARY KEY,
    "match_id" UUID NOT NULL,
    "client_id" UUID NOT NULL,
    "role" VARCHAR(20) NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "updated_at" TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    "deleted_at" TIMESTAMPTZ,
    CONSTRAINT fk_match FOREIGN KEY (match_id) REFERENCES matches(guid)
        ON DELETE CASCADE
        ON UPDATE CASCADE,
    CONSTRAINT fk_client FOREIGN KEY (client_id) REFERENCES clients(guid)
        ON DELETE CASCADE
        ON UPDATE CASCADE
);

CREATE TRIGGER set_match_player_updated_at
BEFORE UPDATE ON "public"."match_players"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column(); 