CREATE TABLE IF NOT EXISTS links (
    id         SERIAL PRIMARY KEY,
    code       TEXT NOT NULL UNIQUE,
    url        TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS clicks (
    id         SERIAL PRIMARY KEY,
    link_id    INT NOT NULL REFERENCES links (id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_clicks_link_id ON clicks (link_id);
