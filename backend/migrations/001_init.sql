-- ============================================================
-- 001_init.sql
-- P2P Exchange Service — 完整資料庫初始化
-- ============================================================

-- ============================================================
-- 1. 使用者表
-- ============================================================

-- app_users：一般使用者（App 平台）
CREATE TABLE IF NOT EXISTS app_users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(50)  UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email         VARCHAR(255),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- backend_users：管理員（Backend 平台）
CREATE TABLE IF NOT EXISTS backend_users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(50)  UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email         VARCHAR(255),
    role          VARCHAR(50)  NOT NULL DEFAULT 'admin',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- 2. 收款方式（payment_methods）
-- ============================================================
CREATE TABLE IF NOT EXISTS payment_methods (
    id             BIGSERIAL    PRIMARY KEY,
    user_id        BIGINT       NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,

    type           VARCHAR(50)  NOT NULL DEFAULT 'bank_transfer'
                   CHECK (type IN ('bank_transfer')),

    bank_name      VARCHAR(100) NOT NULL,
    account_name   VARCHAR(100) NOT NULL,
    account_number VARCHAR(50)  NOT NULL,

    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_methods_user_id ON payment_methods(user_id);

-- ============================================================
-- 3. 掛單（listings）
-- ============================================================
CREATE TABLE IF NOT EXISTS listings (
    id                BIGSERIAL    PRIMARY KEY,
    user_id           BIGINT       NOT NULL REFERENCES app_users(id),

    type              VARCHAR(10)  NOT NULL CHECK (type IN ('buy', 'sell')),

    crypto_currency   VARCHAR(20)  NOT NULL DEFAULT 'USDT',
    fiat_currency     VARCHAR(10)  NOT NULL DEFAULT 'TWD',

    total_amount      NUMERIC(20, 8) NOT NULL,
    remaining_amount  NUMERIC(20, 8) NOT NULL,
    price             NUMERIC(20, 4) NOT NULL,

    min_order_fiat    NUMERIC(20, 4) NOT NULL,
    max_order_fiat    NUMERIC(20, 4) NOT NULL,

    platform_fee_base   NUMERIC(20, 4) NOT NULL DEFAULT 0,
    platform_fee_rate   NUMERIC(8, 6)  NOT NULL DEFAULT 0,
    payment_fee_base    NUMERIC(20, 4) NOT NULL DEFAULT 0,
    payment_fee_rate    NUMERIC(8, 6)  NOT NULL DEFAULT 0,

    payment_time_limit INT          NOT NULL DEFAULT 30,

    payment_method_id BIGINT        REFERENCES payment_methods(id),

    -- v1 付款方式字串（bank_transfer / convenience_store）
    payment_method_label VARCHAR(50),

    status            VARCHAR(20)  NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'paused', 'completed', 'cancelled')),

    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listings_user_id ON listings(user_id);
CREATE INDEX IF NOT EXISTS idx_listings_type_status ON listings(type, status);
CREATE INDEX IF NOT EXISTS idx_listings_crypto ON listings(crypto_currency, fiat_currency);

-- ============================================================
-- 4. 訂單（orders）
-- ============================================================
CREATE TABLE IF NOT EXISTS orders (
    id                 BIGSERIAL    PRIMARY KEY,

    order_no           VARCHAR(30)  UNIQUE NOT NULL,

    listing_id         BIGINT       NOT NULL REFERENCES listings(id),
    listing_type       VARCHAR(10)  NOT NULL CHECK (listing_type IN ('buy', 'sell')),

    seller_id          BIGINT       NOT NULL REFERENCES app_users(id),
    buyer_id           BIGINT       NOT NULL REFERENCES app_users(id),

    crypto_currency    VARCHAR(20)  NOT NULL DEFAULT 'USDT',
    fiat_currency      VARCHAR(10)  NOT NULL DEFAULT 'TWD',

    crypto_amount         NUMERIC(20, 8) NOT NULL,
    price                 NUMERIC(20, 4) NOT NULL,
    fiat_amount           NUMERIC(20, 4) NOT NULL,

    platform_fee_base     NUMERIC(20, 4) NOT NULL DEFAULT 0,
    platform_fee_amount   NUMERIC(20, 4) NOT NULL DEFAULT 0,
    payment_fee_base      NUMERIC(20, 4) NOT NULL DEFAULT 0,
    payment_fee_amount    NUMERIC(20, 4) NOT NULL DEFAULT 0,

    total_fee             NUMERIC(20, 4) NOT NULL DEFAULT 0,
    total_amount          NUMERIC(20, 4) NOT NULL DEFAULT 0,

    payment_method_id  BIGINT       NOT NULL REFERENCES payment_methods(id),

    status             VARCHAR(20)  NOT NULL DEFAULT 'matched'
                       CHECK (status IN (
                           'matched', 'paid', 'releasing', 'completed',
                           'cancelled', 'timeout', 'disputed'
                       )),

    payment_deadline   TIMESTAMPTZ  NOT NULL,
    paid_at            TIMESTAMPTZ,
    confirmed_at       TIMESTAMPTZ,
    completed_at       TIMESTAMPTZ,
    cancelled_at       TIMESTAMPTZ,
    cancel_reason      TEXT,

    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_order_no ON orders(order_no);
CREATE INDEX IF NOT EXISTS idx_orders_listing_id ON orders(listing_id);
CREATE INDEX IF NOT EXISTS idx_orders_seller_id ON orders(seller_id);
CREATE INDEX IF NOT EXISTS idx_orders_buyer_id ON orders(buyer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at DESC);

-- ============================================================
-- 5. 托管記錄（escrow_records）
-- ============================================================
CREATE TABLE IF NOT EXISTS escrow_records (
    id              BIGSERIAL    PRIMARY KEY,
    order_id        BIGINT       NOT NULL REFERENCES orders(id),

    crypto_currency VARCHAR(20)  NOT NULL DEFAULT 'USDT',
    amount          NUMERIC(20, 8) NOT NULL,

    action          VARCHAR(20)  NOT NULL CHECK (action IN ('lock', 'release', 'refund')),

    status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'completed', 'failed')),

    tx_ref          VARCHAR(100),

    remark          TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_escrow_order_id ON escrow_records(order_id);

-- ============================================================
-- 6. 訂單狀態日誌（order_status_logs）
-- ============================================================
CREATE TABLE IF NOT EXISTS order_status_logs (
    id            BIGSERIAL   PRIMARY KEY,
    order_id      BIGINT      NOT NULL REFERENCES orders(id),

    from_status   VARCHAR(20),
    to_status     VARCHAR(20) NOT NULL,

    operator_type VARCHAR(20) NOT NULL
                  CHECK (operator_type IN ('buyer', 'seller', 'admin', 'system')),
    operator_id   BIGINT,

    remark        TEXT,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_order_status_logs_order_id ON order_status_logs(order_id);
