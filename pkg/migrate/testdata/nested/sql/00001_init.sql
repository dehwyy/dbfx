-- +goose Up
CREATE TABLE orders (id bigserial PRIMARY KEY);

-- +goose Down
DROP TABLE orders;
