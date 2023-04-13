CREATE TABLE "user"
(
    "id"         SERIAL PRIMARY KEY,
    "login"      VARCHAR(255) NOT NULL UNIQUE,
    "password"   VARCHAR(255) NOT NULL,
    "balance"    DECIMAL   DEFAULT 0,
    "created_at" TIMESTAMP DEFAULT now(),
    "updated_at" TIMESTAMP DEFAULT now()
);
