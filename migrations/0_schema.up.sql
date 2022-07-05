-- Creating feed table + indexes
CREATE TABLE IF NOT EXISTS feed (
    id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    link VARCHAR(255) NOT NULL,
    feed_link VARCHAR(255) NOT NULL,
    category VARCHAR(255) NOT NULL,
    language VARCHAR (255) NOT NULL,
    provider VARCHAR(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz,
);

-- Creating article table + indexes
CREATE TABLE IF NOT EXISTS article (
    id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    feed_id UUID NOT NULL REFERENCES feed(id),
    title VARCHAR(255) NOT NULL,
    description VARCHAR(255) NOT NULL,
    link VARCHAR(255) NOT NULL,
    thumbnail_url VARCHAR(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz,
    published_at timestamptz,
);

CREATE INDEX article_feed_id_idx ON article(feed_id);