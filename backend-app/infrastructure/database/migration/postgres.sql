CREATE TABLE IF NOT EXISTS url_shortener (
    id VARCHAR(36) PRIMARY KEY,
    short VARCHAR(255) UNIQUE NOT NULL,
    long VARCHAR(2048) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index on short column for faster lookups
CREATE INDEX IF NOT EXISTS idx_url_shortener_short ON url_shortener(short);

-- Index on long column for faster duplicate checks
CREATE INDEX IF NOT EXISTS idx_url_shortener_long ON url_shortener(long);
