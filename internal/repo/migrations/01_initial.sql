--------------------------------------------------------------------------------
-- +goose Up
--------------------------------------------------------------------------------
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Типы операций
CREATE TYPE operation_type AS ENUM (
    'order_accrual',
    'order_withdrawal',
    'promo_accrual'
    );

-- Статусы заказа
CREATE TYPE operation_status AS ENUM (
    'NEW',
    'PROCESSING',
    'PROCESSED',
    'INVALID',
    'CANCELED'
    );


-- Пользователи
CREATE TABLE users
(
    id         INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username   VARCHAR(256)   NOT NULL,
    pass_hash  VARCHAR(256)   NOT NULL,
    balance    DECIMAL(16, 4) NOT NULL DEFAULT 0,
    withdrawn  DECIMAL(16, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMP      NOT NULL DEFAULT now(),
    updated_at TIMESTAMP      NOT NULL DEFAULT now(),
    CONSTRAINT username_unique UNIQUE (username),
    CONSTRAINT balance_not_negative CHECK ( balance >= 0 ),
    CONSTRAINT withdrawn_not_negative CHECK ( withdrawn >= 0 )
);

-- Промо-кампании
CREATE TABLE promos
(
    id          INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    code        VARCHAR(64)    NOT NULL,
    description VARCHAR(256)   NOT NULL,
    reward      DECIMAL(16, 4) NOT NULL,
    valid_until TIMESTAMP      NOT NULL,
    created_at  TIMESTAMP      NOT NULL DEFAULT now(),
    CONSTRAINT promo_code_unique UNIQUE (code),
    CONSTRAINT promo_reward_positive CHECK ( reward > 0 )
);

INSERT INTO promos (code, description, reward, valid_until)
VALUES ('WELCOME-GOPHER', 'Приветственный бонус', 20, '2025-01-01'),
       ('GOLANG-2021', 'В честь дня рождения Go', 10, '2021-10-11');

-- Операции
CREATE TABLE operations
(
    id           INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    user_id      INTEGER          NOT NULL,
    op_type      operation_type   NOT NULL,
    status       operation_status NOT NULL,
    amount       DECIMAL(16, 4)   NOT NULL,
    description  VARCHAR(255)     NOT NULL,
    created_at   TIMESTAMP        NOT NULL DEFAULT now(),
    updated_at   TIMESTAMP        NOT NULL DEFAULT now(),
    order_number VARCHAR(64)               DEFAULT NULL,
    promo_id     INTEGER                   DEFAULT NULL,
    CONSTRAINT must_refs_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT order_belongs_to_user EXCLUDE USING gist (order_number WITH =, user_id WITH <>),
    CONSTRAINT order_unique_for_op_type UNIQUE (order_number, op_type),
    CONSTRAINT must_refs_promo FOREIGN KEY (promo_id) REFERENCES promos (id),
    CONSTRAINT promo_unique_for_user UNIQUE (promo_id, user_id),
    CONSTRAINT amount_valid_sign CHECK (
            (amount >= 0 AND op_type IN ('order_accrual', 'promo_accrual'))
            OR
            (amount <= 0 AND op_type IN ('order_withdrawal'))
        ),
    CONSTRAINT operation_valid_attrs CHECK (
            (op_type = 'order_accrual' AND order_number IS NOT NULL and promo_id IS NULL)
            OR
            (op_type = 'order_withdrawal' AND order_number IS NOT NULL AND promo_id IS NULL)
            OR
            (op_type = 'promo_accrual' AND order_number IS NULL AND promo_id IS NOT NULL)
        )
);

CREATE INDEX total_accrued_idx ON operations (user_id)
    INCLUDE (amount)
    WHERE status = 'PROCESSED' AND amount >= 0;

CREATE INDEX total_withdrawn_idx ON operations (user_id)
    INCLUDE (amount)
    WHERE status NOT IN ('INVALID', 'CANCELED') AND amount < 0;

CREATE INDEX update_further_idx on operations (op_type, updated_at ASC)
    WHERE status IN ('NEW', 'PROCESSING');

--------------------------------------------------------------------------------
-- +goose Down
--------------------------------------------------------------------------------
DROP TABLE IF EXISTS operations;
DROP TABLE IF EXISTS promos;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS operation_status;
DROP TYPE IF EXISTS operation_type;
