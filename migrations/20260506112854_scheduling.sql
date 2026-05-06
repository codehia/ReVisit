-- +goose Up
ALTER TABLE cards ADD ease_factor REAL DEFAULT 2.5;
ALTER TABLE cards ADD interval_days INTEGER DEFAULT 1;
ALTER TABLE cards ADD due_date TIMESTAMP;

CREATE TABLE card_attempts (
    id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
    card_id TEXT NOT NULL REFERENCES cards(id),
    quality_score INTEGER NOT NULL CHECK (quality_score BETWEEN 0 AND 5),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_timestamp_cards
AFTER UPDATE ON card_attempts
FOR EACH ROW
WHEN OLD.id = NEW.id
BEGIN
    UPDATE cards
    SET updated_at=CURRENT_TIMESTAMP
    WHERE id=NEW.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE card_attempts;
ALTER TABLE cards DROP COLUMN ease_factor;
ALTER TABLE cards DROP COLUMN interval_days;
ALTER TABLE cards DROP COLUMN due_date;
-- +goose StatementEnd
