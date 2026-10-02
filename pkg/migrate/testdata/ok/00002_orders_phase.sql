-- +goose Up
ALTER TABLE orders ADD COLUMN balance_phase text;

-- +goose Down
ALTER TABLE orders DROP COLUMN balance_phase;
