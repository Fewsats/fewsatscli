-- api_keys is a table that stores the API Keys for accessing the fewsats 
-- platform API by the users.
CREATE TABLE IF NOT EXISTS api_keys (
    -- id is the primary key of the table.
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- key is the API Key that is used to authenticate the user.
    key TEXT NOT NULL,
    -- expires_at is the date and time when the API Key expires.
    expires_at DATETIME,
    -- user_id is the ID of the user that the API Key belongs to.
    user_id INTEGER NOT NULL,
    -- enabled is a flag that indicates whether the API Key is enabled or not.
    enabled BOOLEAN NOT NULL DEFAULT 1
);
