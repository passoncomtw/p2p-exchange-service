-- ============================================================
-- 003_currencies.sql
-- 幣種註冊表 + 現有資料欄位 FK 約束
-- ============================================================

-- ============================================================
-- 1. 幣種註冊表
-- ============================================================
-- code：平台內部唯一識別碼，格式 {TICKER}-{NETWORK} 或 {ISO4217}
--   加密：USDT-TRC20 / USDT-ERC20 / BTC / ETH
--   法幣：TWD / USD / JPY（ISO 4217）
-- network：鏈名稱，法幣為 NULL
-- decimals：顯示精度（儲存仍用 NUMERIC(30,8)）

CREATE TABLE IF NOT EXISTS currencies (
    code         VARCHAR(30)  PRIMARY KEY,
    display_name VARCHAR(100) NOT NULL,
    type         VARCHAR(10)  NOT NULL CHECK (type IN ('crypto', 'fiat')),
    network      VARCHAR(30),
    decimals     SMALLINT     NOT NULL DEFAULT 8,
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- 2. 初始支援幣種
-- ============================================================
INSERT INTO currencies (code, display_name, type, network, decimals) VALUES
    ('USDT-TRC20', 'Tether (TRC-20)',  'crypto', 'tron',     6),
    ('USDT-ERC20', 'Tether (ERC-20)',  'crypto', 'ethereum', 6),
    ('USDT-BEP20', 'Tether (BEP-20)',  'crypto', 'bsc',      18),
    ('BTC',        'Bitcoin',          'crypto', 'bitcoin',  8),
    ('ETH',        'Ethereum',         'crypto', 'ethereum', 18),
    ('TWD',        '新台幣',            'fiat',   NULL,       0),
    ('USD',        'US Dollar',        'fiat',   NULL,       2),
    ('JPY',        '日本円',            'fiat',   NULL,       0)
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- 3. 現有資料遷移：USDT → USDT-TRC20（平台預設鏈）
-- ============================================================
UPDATE listings       SET crypto_currency = 'USDT-TRC20' WHERE crypto_currency = 'USDT';
UPDATE orders         SET crypto_currency = 'USDT-TRC20' WHERE crypto_currency = 'USDT';
UPDATE escrow_records SET crypto_currency = 'USDT-TRC20' WHERE crypto_currency = 'USDT';

-- ============================================================
-- 4. 欄位預設值更新
-- ============================================================
ALTER TABLE listings       ALTER COLUMN crypto_currency SET DEFAULT 'USDT-TRC20';
ALTER TABLE orders         ALTER COLUMN crypto_currency SET DEFAULT 'USDT-TRC20';
ALTER TABLE escrow_records ALTER COLUMN crypto_currency SET DEFAULT 'USDT-TRC20';

-- ============================================================
-- 5. FK 約束
-- ============================================================
ALTER TABLE listings
    ADD CONSTRAINT fk_listings_crypto_currency
        FOREIGN KEY (crypto_currency) REFERENCES currencies(code),
    ADD CONSTRAINT fk_listings_fiat_currency
        FOREIGN KEY (fiat_currency)   REFERENCES currencies(code);

ALTER TABLE orders
    ADD CONSTRAINT fk_orders_crypto_currency
        FOREIGN KEY (crypto_currency) REFERENCES currencies(code),
    ADD CONSTRAINT fk_orders_fiat_currency
        FOREIGN KEY (fiat_currency)   REFERENCES currencies(code);

ALTER TABLE escrow_records
    ADD CONSTRAINT fk_escrow_crypto_currency
        FOREIGN KEY (crypto_currency) REFERENCES currencies(code);
