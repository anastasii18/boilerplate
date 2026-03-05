-- +goose Up
create type order_payment_method as enum (
    'UNKNOWN',
	'CARD',
	'SBP',
	'CREDIT_CARD',
	'INVESTOR_MONEY'
);

create type order_status as enum(
    'PENDING_PAYMENT',
	'PAID',
	'CANCELLED'
);

create table  "order" (
   order_uuid  UUID                     PRIMARY KEY,
   user_uuid       VARCHAR(36)          NOT NULL,
   part_uuids      TEXT[]               NOT NULL DEFAULT '{}',
   total_price     NUMERIC(12,2)        NOT NULL CHECK (total_price >= 0),
   transaction_uuid VARCHAR(36),
   payment_method  order_payment_method NOT NULL,
   status          order_status         NOT NULL,
   created_at      TIMESTAMPTZ          NOT NULL DEFAULT NOW(),
   updated_at      TIMESTAMPTZ
);
-- +goose Down
drop table "order";
drop type order_payment_method;
drop type order_status;