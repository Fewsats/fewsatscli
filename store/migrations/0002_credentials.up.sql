--  credentials stores the L402 credentials that the user can use to access
-- L402 paywalled services.
CREATE TABLE IF NOT EXISTS credentials (
    -- id is the primary key of the table.
    id INTEGER PRIMARY KEY AUTOINCREMENT,

    -- external_id is the external ID for the resource.
    external_id TEXT NOT NULL,
    -- macaroon is the base64 encoded macaroon needed in the 
    -- L402 request header.
    macaroon TEXT NOT NULL,
    -- preimage is the preimage linked to the macaroon payment
    -- hash. Also needed in the L402 request header.
    preimage TEXT NOT NULL,
    -- invoice is the LN invoice that was paid to complete the
    -- credentials.
    invoice TEXT NOT NULL,
    
    -- created_at is the date and time when the credentials were
    -- created.
    created_at DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS credentials_external_id_index ON credentials (external_id);
