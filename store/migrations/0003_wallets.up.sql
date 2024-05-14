-- wallets is a table that stores the connected wallets of the user.
CREATE TABLE IF NOT EXISTS wallets (
    -- id is the primary key of the table.
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- wallet_type is the type of the wallet.
    wallet_type TEXT NOT NULL,
    -- expires_at is the date and time when the wallet expires.
    expires_at DATETIME,
    -- created_at is the date and time when the wallet was created.
    created_at DATETIME NOT NULL
);

-- default_wallet is a table that stores the current wallet of the user.
CREATE TABLE IF NOT EXISTS default_wallet (
    -- wallet_id is the ID of the current wallet.
    wallet_id INTEGER NOT NULL
);

-- wallet_tokens is a table that stores the tokens for the wallets
-- connections that are token based.
CREATE TABLE IF NOT EXISTS wallet_tokens (
    -- wallet_id is the ID of the wallet that the token belongs to.
    wallet_id INTEGER NOT NULL,
    -- token is the token that is used to authenticate the wallet.
    token TEXT NOT NULL
);