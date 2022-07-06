DROP EXTENSION IF EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Creating feed table + indexes
CREATE TABLE IF NOT EXISTS feed (
    id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4(),
    title varchar(255) NOT NULL,
    description varchar(255) NOT NULL,
    link varchar(255) NOT NULL,
    feed_link varchar(255) NOT NULL UNIQUE,
    category varchar(255) NOT NULL,
    language VARCHAR(255) NOT NULL,
    provider varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz
);

-- Creating article table + indexes
CREATE TABLE IF NOT EXISTS article (
    id uuid NOT NULL PRIMARY KEY DEFAULT uuid_generate_v4 (),
    feed_id uuid NOT NULL REFERENCES feed (id),
    title varchar(255) NOT NULL,
    description varchar(255) NOT NULL,
    link varchar(255) NOT NULL,
    thumbnail_url varchar(255) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz,
    published_at timestamptz
);

CREATE INDEX article_feed_id_idx ON article (feed_id);

