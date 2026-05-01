-- +goose Up
CREATE TABLE cards (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    question    TEXT NOT NULL,
    answer      TEXT NOT NULL,
    examples    TEXT NOT NULL,
    tradeoffs   TEXT NOT NULL,
    card_type   TEXT NOT NULL CHECK(card_type IN ('definition', 'mechanism', 'tradeoff', 'application')),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_timestamp_cards
AFTER UPDATE ON cards
FOR EACH ROW
WHEN OLD.id = NEW.id
BEGIN
    UPDATE cards
    SET updated_at=CURRENT_TIMESTAMP
    WHERE id=NEW.id;
END;
-- +goose StatementEnd


CREATE TABLE tags (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    name        TEXT NOT NULL UNIQUE,
    parent_id   TEXT REFERENCES tags(id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_timestamp_tags
AFTER UPDATE ON tags
FOR EACH ROW
BEGIN
    UPDATE tags
    SET updated_at = CURRENT_TIMESTAMP
    WHERE id = NEW.id;
END;
-- +goose StatementEnd

CREATE TABLE card_tags (
    card_id     TEXT NOT NULL REFERENCES cards(id),
    tag_id      TEXT NOT NULL REFERENCES tags(id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (card_id, tag_id)
);


-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_timestamp_card_tags
AFTER UPDATE ON card_tags
FOR EACH ROW
BEGIN
    UPDATE card_tags
    SET updated_at = CURRENT_TIMESTAMP
    WHERE card_id = NEW.card_id AND tag_id = NEW.tag_id;
END;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS card_tags;
DROP TRIGGER IF EXISTS update_timestamp_card_tags;
DROP TABLE IF EXISTS tags;
DROP TRIGGER IF EXISTS update_timestamp_tags;
DROP TABLE IF EXISTS cards;
DROP TRIGGER IF EXISTS update_timestamp_cards;
