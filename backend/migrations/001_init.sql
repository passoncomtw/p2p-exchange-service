-- ============================================================
-- 001_init.sql
-- P2P Exchange Service — 完整資料庫初始化
-- ============================================================

-- ============================================================
-- 1. 使用者表
-- ============================================================

CREATE TABLE IF NOT EXISTS app_users (
    id            BIGSERIAL    PRIMARY KEY,
    username      VARCHAR(50)  UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email         VARCHAR(255),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS backend_users (
    id            BIGSERIAL    PRIMARY KEY,
    username      VARCHAR(50)  UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email         VARCHAR(255),
    role          VARCHAR(50)  NOT NULL DEFAULT 'admin',
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- 2. 幣種註冊表
-- ============================================================
-- code：平台幣種唯一識別碼
--   加密貨幣：代幣 Ticker（USDT / BTC / ETH）
--   法幣：ISO 4217（TWD / USD / JPY）
-- type：crypto / fiat
-- is_active：控制平台是否開放此幣種交易

CREATE TABLE IF NOT EXISTS currencies (
    code         VARCHAR(20)  PRIMARY KEY,
    display_name VARCHAR(100) NOT NULL,
    type         VARCHAR(10)  NOT NULL CHECK (type IN ('crypto', 'fiat')),
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO currencies (code, display_name, type) VALUES
    ('USDT', 'Tether',      'crypto'),
    ('BTC',  'Bitcoin',     'crypto'),
    ('ETH',  'Ethereum',    'crypto'),
    ('TWD',  '新台幣',       'fiat'),
    ('USD',  'US Dollar',   'fiat'),
    ('JPY',  '日本円',       'fiat')
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- 3. 收款方式（payment_methods）
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
-- 4. 掛單（listings）
-- ============================================================
-- crypto 金額：NUMERIC(38, 18)，對應加密貨幣最高 18 位小數
-- fiat 金額：NUMERIC(20, 4)，法幣不需要小數精度

CREATE TABLE IF NOT EXISTS listings (
    id                BIGSERIAL    PRIMARY KEY,
    user_id           BIGINT       NOT NULL REFERENCES app_users(id),

    type              VARCHAR(10)  NOT NULL CHECK (type IN ('buy', 'sell')),

    crypto_currency   VARCHAR(20)  NOT NULL DEFAULT 'USDT' REFERENCES currencies(code),
    fiat_currency     VARCHAR(20)  NOT NULL DEFAULT 'TWD'  REFERENCES currencies(code),

    total_amount      NUMERIC(38, 18) NOT NULL,
    remaining_amount  NUMERIC(38, 18) NOT NULL,
    price             NUMERIC(20, 8)  NOT NULL,

    min_order_fiat    NUMERIC(20, 4) NOT NULL,
    max_order_fiat    NUMERIC(20, 4) NOT NULL,

    platform_fee_base   NUMERIC(20, 4) NOT NULL DEFAULT 0,
    platform_fee_rate   NUMERIC(10, 8) NOT NULL DEFAULT 0,
    payment_fee_base    NUMERIC(20, 4) NOT NULL DEFAULT 0,
    payment_fee_rate    NUMERIC(10, 8) NOT NULL DEFAULT 0,

    payment_time_limit INT          NOT NULL DEFAULT 30,

    payment_method_id BIGINT       REFERENCES payment_methods(id),

    payment_method_label VARCHAR(50),

    status            VARCHAR(20)  NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'paused', 'completed', 'cancelled')),

    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listings_user_id       ON listings(user_id);
CREATE INDEX IF NOT EXISTS idx_listings_type_status   ON listings(type, status);
CREATE INDEX IF NOT EXISTS idx_listings_crypto        ON listings(crypto_currency, fiat_currency);

-- ============================================================
-- 5. 訂單（orders）
-- ============================================================

CREATE TABLE IF NOT EXISTS orders (
    id                 BIGSERIAL    PRIMARY KEY,

    order_no           VARCHAR(30)  UNIQUE NOT NULL,

    listing_id         BIGINT       NOT NULL REFERENCES listings(id),
    listing_type       VARCHAR(10)  NOT NULL CHECK (listing_type IN ('buy', 'sell')),

    seller_id          BIGINT       NOT NULL REFERENCES app_users(id),
    buyer_id           BIGINT       NOT NULL REFERENCES app_users(id),

    crypto_currency    VARCHAR(20)  NOT NULL DEFAULT 'USDT' REFERENCES currencies(code),
    fiat_currency      VARCHAR(20)  NOT NULL DEFAULT 'TWD'  REFERENCES currencies(code),

    crypto_amount         NUMERIC(38, 18) NOT NULL,
    price                 NUMERIC(20, 8)  NOT NULL,
    fiat_amount           NUMERIC(20, 4)  NOT NULL,

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

CREATE INDEX IF NOT EXISTS idx_orders_order_no    ON orders(order_no);
CREATE INDEX IF NOT EXISTS idx_orders_listing_id  ON orders(listing_id);
CREATE INDEX IF NOT EXISTS idx_orders_seller_id   ON orders(seller_id);
CREATE INDEX IF NOT EXISTS idx_orders_buyer_id    ON orders(buyer_id);
CREATE INDEX IF NOT EXISTS idx_orders_status      ON orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_created_at  ON orders(created_at DESC);

-- ============================================================
-- 6. 托管記錄（escrow_records）
-- ============================================================

CREATE TABLE IF NOT EXISTS escrow_records (
    id              BIGSERIAL    PRIMARY KEY,
    order_id        BIGINT       NOT NULL REFERENCES orders(id),

    crypto_currency VARCHAR(20)  NOT NULL DEFAULT 'USDT' REFERENCES currencies(code),
    amount          NUMERIC(38, 18) NOT NULL,

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
-- 7. 錢包（wallets）
-- ============================================================
-- 一位使用者 × 一種幣種 = 一個錢包
-- available_balance：可用餘額（可交易、可提領）
-- frozen_balance：凍結中（訂單鎖倉期間不可動用）

CREATE TABLE IF NOT EXISTS wallets (
    id                BIGSERIAL      PRIMARY KEY,
    user_id           BIGINT         NOT NULL REFERENCES app_users(id),
    currency          VARCHAR(20)    NOT NULL REFERENCES currencies(code),

    available_balance NUMERIC(38, 18) NOT NULL DEFAULT 0,
    frozen_balance    NUMERIC(38, 18) NOT NULL DEFAULT 0,

    created_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ    NOT NULL DEFAULT NOW(),

    UNIQUE (user_id, currency)
);

CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

-- ============================================================
-- 8. 帳本流水（wallet_ledgers）
-- ============================================================
-- append-only：只新增，不修改不刪除
-- amount：正數為增加，負數為減少
-- balance_after：異動後的 available_balance 快照（稽核用）
-- type：
--   freeze          接單時凍結賣方加密貨幣
--   unfreeze        取消訂單時解凍退還
--   transfer_in     買方收到加密貨幣（訂單完成）
--   transfer_out    賣方放行加密貨幣（訂單完成）
--   fee_deduct      平台手續費扣除（預留，v0.2.0 費率為 0）

CREATE TABLE IF NOT EXISTS wallet_ledgers (
    id            BIGSERIAL      PRIMARY KEY,
    wallet_id     BIGINT         NOT NULL REFERENCES wallets(id),

    type          VARCHAR(20)    NOT NULL
                  CHECK (type IN (
                      'freeze', 'unfreeze',
                      'transfer_in', 'transfer_out',
                      'fee_deduct'
                  )),

    amount        NUMERIC(38, 18) NOT NULL,
    balance_after NUMERIC(38, 18) NOT NULL,

    ref_order_no  VARCHAR(30),

    created_at    TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallet_ledgers_wallet_id     ON wallet_ledgers(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_ledgers_ref_order_no  ON wallet_ledgers(ref_order_no);
CREATE INDEX IF NOT EXISTS idx_wallet_ledgers_created_at    ON wallet_ledgers(created_at DESC);

-- ============================================================
-- 9. 訂單狀態日誌（order_status_logs）
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
