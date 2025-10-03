-- +goose Up

CREATE
EXTENSION IF NOT EXISTS pgcrypto;


CREATE TABLE politicians
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT pk_politicians PRIMARY KEY (id),

    last_name  TEXT        NOT NULL,
    first_name TEXT        NOT NULL,
    image_url  TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);


CREATE TABLE governments
(
    id                UUID        NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT pk_governments PRIMARY KEY (id),

    prime_minister_id UUID        NOT NULL,
    CONSTRAINT fk_governments_prime_minister FOREIGN KEY (prime_minister_id) REFERENCES politicians (id),

    reference         SMALLINT    NOT NULL,
    CONSTRAINT uq_governments_reference UNIQUE (reference),

    start_date        TIMESTAMPTZ NOT NULL,
    end_date          TIMESTAMPTZ,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);


CREATE TABLE occupations
(
    id                     UUID        NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT pk_occupations PRIMARY KEY (id),

    politician_id          UUID        NOT NULL,
    CONSTRAINT fk_occupations_politician FOREIGN KEY (politician_id) REFERENCES politicians (id),

    government_id          UUID,
    CONSTRAINT fk_occupations_government FOREIGN KEY (government_id) REFERENCES governments (id),

    presidential_reference SMALLINT,
    CONSTRAINT uq_occupations_presidential UNIQUE (presidential_reference),

    CONSTRAINT ck_occupations CHECK ((presidential_reference IS NOT NULL AND government_id IS NULL)
        OR (presidential_reference IS NULL AND government_id IS NOT NULL)),

    code                   TEXT        NOT NULL,
    title                  TEXT        NOT NULL,
    start_date             TIMESTAMPTZ NOT NULL,
    end_date               TIMESTAMPTZ,

    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at             TIMESTAMPTZ
);


CREATE TABLE roles
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT pk_roles PRIMARY KEY (id),

    name       TEXT        NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_roles_name ON roles (name);


CREATE TABLE users
(
    id         UUID        NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT pk_users PRIMARY KEY (id),

    tag        TEXT        NOT NULL,
    email      TEXT        NOT NULL,
    password   TEXT        NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uq_users_tag ON users (tag);
CREATE UNIQUE INDEX uq_users_email ON users (email);


CREATE TABLE user_roles
(
    user_id UUID NOT NULL,
    role_id UUID NOT NULL,
    CONSTRAINT fk_user_roles_user FOREIGN KEY (user_id) REFERENCES users (id),
    CONSTRAINT fk_user_roles_role FOREIGN KEY (role_id) REFERENCES roles (id),
    CONSTRAINT pk_user_roles PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user ON user_roles (user_id);


CREATE TABLE user_tokens
(
    id       UUID        NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT pk_user_tokens PRIMARY KEY (id),

    user_id  UUID        NOT NULL,
    CONSTRAINT fk_user_tokens_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,

    category TEXT        NOT NULL,
    CONSTRAINT ck_user_tokens_category CHECK (category IN ('REFRESH', 'PASSWORD')),

    hash     TEXT        NOT NULL,
    CONSTRAINT uq_user_tokens_hash UNIQUE (hash),

    expiry   TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_tokens_hash_category_expiry ON user_tokens (hash, category, expiry);
CREATE INDEX IF NOT EXISTS idx_user_tokens_user_category ON user_tokens (user_id, category);


CREATE TABLE articles
(
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT pk_articles PRIMARY KEY (id),

    redactor_id  UUID        NOT NULL,
    CONSTRAINT fk_articles_redactor FOREIGN KEY (redactor_id) REFERENCES users (id),

    moderator_id UUID,
    CONSTRAINT fk_articles_moderator FOREIGN KEY (moderator_id) REFERENCES users (id),
    CONSTRAINT ck_articles_moderator CHECK (moderator_id IS NULL OR moderator_id <> redactor_id),

    status       TEXT        NOT NULL,
    CONSTRAINT ck_articles_status CHECK (status IN
                                         ('DRAFT', 'PUBLISHED', 'ARCHIVED', 'UNDER_REVIEW', 'CHANGE_REQUESTED')),

    category     TEXT        NOT NULL,
    CONSTRAINT ck_articles_category CHECK (category IN ('LIE', 'FALSEHOOD')),

    title        TEXT        NOT NULL,
    body         TEXT        NOT NULL,
    event_date   TIMESTAMPTZ NOT NULL,

    reference    UUID        NOT NULL,
    major        SMALLINT    NOT NULL,
    minor        SMALLINT    NOT NULL,
    CONSTRAINT uq_articles_version UNIQUE (reference, major, minor),

    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_articles_reference_not_archived ON articles (reference) WHERE status <> 'ARCHIVED';
CREATE INDEX idx_articles_redactor_status_created_at_desc ON articles (redactor_id, status, created_at DESC);
CREATE INDEX idx_articles_moderator_status_created_at_desc ON articles (moderator_id, status, created_at DESC);
CREATE INDEX idx_articles_redactor_reference_created_at_desc ON articles (redactor_id, reference, created_at DESC);

CREATE TABLE article_sources
(
    article_id UUID NOT NULL,
    CONSTRAINT fk_article_sources_article FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE,

    url        TEXT NOT NULL,

    CONSTRAINT pk_article_sources PRIMARY KEY (article_id, url)
);

CREATE INDEX idx_article_sources_article ON article_sources (article_id);


CREATE TABLE article_tags
(
    article_id UUID NOT NULL,
    CONSTRAINT fk_article_tags_article FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE,

    tag        TEXT NOT NULL,
    CONSTRAINT pk_article_tags PRIMARY KEY (article_id, tag)
);

CREATE INDEX idx_article_tags_article ON article_tags (article_id);


CREATE TABLE article_reviews
(
    id           UUID        NOT NULL DEFAULT gen_random_uuid(),
    CONSTRAINT pk_article_reviews PRIMARY KEY (id),

    article_id   UUID        NOT NULL,
    CONSTRAINT fk_article_reviews_article FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE,
    CONSTRAINT uq_article_reviews_article UNIQUE (article_id),

    moderator_id UUID        NOT NULL,
    CONSTRAINT fk_article_reviews_moderator FOREIGN KEY (moderator_id) REFERENCES users (id),

    decision     TEXT        NOT NULL,
    CONSTRAINT ck_article_reviews_decision CHECK (decision IN ('PUBLISHED', 'ARCHIVED', 'CHANGE_REQUESTED')),

    notes        TEXT        NOT NULL,

    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);

CREATE INDEX idx_article_reviews_article ON article_reviews (article_id);


CREATE TABLE article_politicians
(
    article_id    UUID NOT NULL,
    CONSTRAINT fk_article_politicians_article FOREIGN KEY (article_id) REFERENCES articles (id) ON DELETE CASCADE,

    politician_id UUID NOT NULL,
    CONSTRAINT fk_article_politicians_politician FOREIGN KEY (politician_id) REFERENCES politicians (id),

    CONSTRAINT pk_article_politicians PRIMARY KEY (article_id, politician_id)
);

CREATE INDEX idx_article_politicians_article ON article_politicians (article_id);
CREATE INDEX idx_article_politicians_politician ON article_politicians (politician_id);

-- TEST DATA

INSERT INTO roles (id, name)
VALUES ('de966b93-a885-45e9-8ed9-48074912de55', 'ADMIN'),
       ('02f0eccd-b0b2-42c0-aef1-6b306ca23005', 'MODERATOR'),
       ('86080136-8365-4541-a918-9ad8f1fd27ac', 'REDACTOR');

INSERT INTO users (id, tag, email, password)
VALUES ('2d7c2090-179e-4084-9489-85b6a70934bc', 'admin0123', 'admin@vdm.com',
        '$2a$12$oAvivQopMo3ZjlebA2BwgO4zENkTuf.4M5y4tDnDM.9YxrGmc.h42'), -- password: Test123!
       ('6df0b4ab-9ad4-408c-9eca-f337136a6ede', 'moderator0123', 'moderator@vdm.com',
        '$2a$12$oAvivQopMo3ZjlebA2BwgO4zENkTuf.4M5y4tDnDM.9YxrGmc.h42'), -- password: Test123!
       ('15611c31-a7b6-4cf7-affc-db2a49072d63', 'redactor0123', 'redactor@vdm.com',
        '$2a$12$oAvivQopMo3ZjlebA2BwgO4zENkTuf.4M5y4tDnDM.9YxrGmc.h42'); -- password: Test123!

INSERT INTO user_roles (user_id, role_id)
VALUES ('2d7c2090-179e-4084-9489-85b6a70934bc', 'de966b93-a885-45e9-8ed9-48074912de55'), -- admin0123 role ADMIN
       ('2d7c2090-179e-4084-9489-85b6a70934bc', '02f0eccd-b0b2-42c0-aef1-6b306ca23005'), -- admin0123 role MODERATOR
       ('2d7c2090-179e-4084-9489-85b6a70934bc', '86080136-8365-4541-a918-9ad8f1fd27ac'), -- admin0123 role REDACTOR
       ('6df0b4ab-9ad4-408c-9eca-f337136a6ede', '02f0eccd-b0b2-42c0-aef1-6b306ca23005'), -- moderator0123 role MODERATOR
       ('6df0b4ab-9ad4-408c-9eca-f337136a6ede', '86080136-8365-4541-a918-9ad8f1fd27ac'), -- moderator0123 role REDACTOR
       ('15611c31-a7b6-4cf7-affc-db2a49072d63', '86080136-8365-4541-a918-9ad8f1fd27ac'); -- redactor0123 role REDACTOR
