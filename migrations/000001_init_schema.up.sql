CREATE TABLE IF NOT EXISTS short_urls (
    short_url    VARCHAR(16) PRIMARY KEY,
    original_url TEXT NOT NULL
);
