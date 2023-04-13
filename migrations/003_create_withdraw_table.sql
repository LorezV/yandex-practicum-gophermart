CREATE TABLE "withdraw"
(
    "id"         SERIAL PRIMARY KEY,
    "user_id"    INT          NOT NULL REFERENCES "public"."user" (id),
    "number"     VARCHAR(255) NOT NULL UNIQUE,
    "sum"        DECIMAL      NULL,
    "created_at" TIMESTAMP DEFAULT now()
);
