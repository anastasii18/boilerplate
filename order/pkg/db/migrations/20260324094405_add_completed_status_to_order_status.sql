-- +goose Up
-- +goose StatementBegin
ALTER TYPE order_status ADD VALUE 'COMPLETED';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Full recreation (more dangerous, use only if necessary)
ALTER TYPE order_status RENAME TO order_status_old;
CREATE TYPE order_status AS ENUM('PENDING_PAYMENT', 'PAID', 'CANCELLED');
ALTER TABLE "order" ALTER COLUMN status TYPE order_status USING status::text::order_status;
DROP TYPE order_status_old;
-- +goose StatementEnd
