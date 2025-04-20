CREATE TABLE "public"."clients" (
    "guid" UUID PRIMARY KEY,
    "username" VARCHAR(60) UNIQUE NOT NULL,
    "password" VARCHAR(255) NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    "deleted_at" TIMESTAMP WITH TIME ZONE
);
CREATE TRIGGER set_client_updated_at
BEFORE UPDATE ON "public"."clients"
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();