CREATE TABLE IF NOT EXISTS github_links (
    id BIGSERIAL PRIMARY KEY,
    url TEXT NOT NULL,
    periodic_interval TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS github_links_url_unique_idx ON github_links (url);
