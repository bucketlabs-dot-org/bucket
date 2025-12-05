--
-- USERS TABLE
--
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    tier TEXT NOT NULL DEFAULT 'free',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

--
-- API KEYS TABLE
--
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    api_key TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);

--
-- FILES TABLE
--
CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    size_bytes BIGINT NOT NULL,
    storage_path TEXT NOT NULL,
    tiny_code TEXT UNIQUE NOT NULL,
    download_secret_hash TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_files_user_id ON files(user_id);
CREATE INDEX IF NOT EXISTS idx_files_tiny_code ON files(tiny_code);
CREATE INDEX IF NOT EXISTS idx_files_expires_at ON files(expires_at);

--
-- CLEANUP EXPIRED FILES (optional manual job)
-- DELETE FROM files WHERE expires_at < NOW();
--

