CREATE TABLE IF NOT EXISTS cards
(
    email TEXT NOT NULL UNIQUE,
    balance INTEGER NOT NULL DEFAULT 0,
    phone_numberHash BLOB NOT NULL, 
    card_numberHash BLOB NOT NULL,
    cvcHash BLOB NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_email ON cards (email);
