-- ============================================================
-- 002_orders.sql
-- P2P Exchange — 掛單、訂單、收款方式、托管、狀態日誌
--
-- 設計原則：
--   - 幣種欄位保留 VARCHAR，預設 USDT / TWD，未來多幣種直接擴展
--   - 手續費拆為四個欄位，第一階段均為 0：
--       platform_fee_base   平台固定基礎費
--       platform_fee_amount 平台比例費（fiat_amount × rate）
--       payment_fee_base    收款管道固定基礎費
--       payment_fee_amount  收款管道比例費（fiat_amount × rate）
--     total_fee = 四者相加；total_amount = fiat_amount + total_fee
--   - payment_methods 以 type 區分收款方式，第一階段僅 bank_transfer
--   - 所有狀態用 VARCHAR + CHECK，避免過早引入 ENUM 難以 ALTER
-- ============================================================

-- ============================================================
-- 1. 收款方式（payment_methods）
--    歸屬 app_users（賣家綁定），一人可多筆
-- ============================================================
CREATE TABLE IF NOT EXISTS payment_methods (
    id             BIGSERIAL    PRIMARY KEY,
    user_id        BIGINT       NOT NULL REFERENCES app_users(id) ON DELETE CASCADE,

    -- 第一階段：bank_transfer；未來可擴充 linepay / jkopay 等
    type           VARCHAR(50)  NOT NULL DEFAULT 'bank_transfer'
                   CHECK (type IN ('bank_transfer')),

    bank_name      VARCHAR(100) NOT NULL,
    account_name   VARCHAR(100) NOT NULL,   -- 戶名
    account_number VARCHAR(50)  NOT NULL,   -- 帳號

    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,

    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payment_methods_user_id ON payment_methods(user_id);


-- ============================================================
-- 2. 掛單（listings）
--    掛買（type='buy'）或掛賣（type='sell'）由使用者發布
-- ============================================================
CREATE TABLE IF NOT EXISTS listings (
    id                BIGSERIAL    PRIMARY KEY,
    user_id           BIGINT       NOT NULL REFERENCES app_users(id),

    type              VARCHAR(10)  NOT NULL CHECK (type IN ('buy', 'sell')),

    -- 幣種（預留多幣種）
    crypto_currency   VARCHAR(20)  NOT NULL DEFAULT 'USDT',
    fiat_currency     VARCHAR(10)  NOT NULL DEFAULT 'TWD',

    -- 數量與價格
    total_amount      NUMERIC(20, 8) NOT NULL,          -- 掛單總量（USDT）
    remaining_amount  NUMERIC(20, 8) NOT NULL,          -- 剩餘可交易量
    price             NUMERIC(20, 4) NOT NULL,          -- 單價（TWD/USDT）

    -- 每筆訂單限制（法幣金額）
    min_order_fiat    NUMERIC(20, 4) NOT NULL,          -- 最低交易金額 TWD
    max_order_fiat    NUMERIC(20, 4) NOT NULL,          -- 最高交易金額 TWD

    -- 手續費設定（第一階段均為 0）
    -- 每筆訂單實際費用 = 基礎固定費 + (交易法幣金額 × 比例費率)
    --
    -- 平台手續費
    platform_fee_base   NUMERIC(20, 4) NOT NULL DEFAULT 0,  -- 平台固定基礎費（TWD）
    platform_fee_rate   NUMERIC(8, 6)  NOT NULL DEFAULT 0,  -- 平台比例費率（例如 0.001 = 0.1%）
    -- 收款管道手續費（依 payment_method.type 不同費率）
    payment_fee_base    NUMERIC(20, 4) NOT NULL DEFAULT 0,  -- 管道固定基礎費（TWD）
    payment_fee_rate    NUMERIC(8, 6)  NOT NULL DEFAULT 0,  -- 管道比例費率

    -- 付款時限（分鐘，買家接單後需在此時間內完成付款）
    payment_time_limit INT          NOT NULL DEFAULT 30,

    -- 掛賣單的收款方式（buy 時為 NULL）
    payment_method_id BIGINT        REFERENCES payment_methods(id),

    -- active: 可接單 | paused: 暫停 | completed: 額度用盡 | cancelled: 主動撤單
    status            VARCHAR(20)  NOT NULL DEFAULT 'active'
                      CHECK (status IN ('active', 'paused', 'completed', 'cancelled')),

    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_listings_user_id ON listings(user_id);
CREATE INDEX IF NOT EXISTS idx_listings_type_status ON listings(type, status);
CREATE INDEX IF NOT EXISTS idx_listings_crypto ON listings(crypto_currency, fiat_currency);


-- ============================================================
-- 3. 訂單（orders）
--    接單後由系統產生，一筆 listing 可對應多筆 order（部分成交）
-- ============================================================
CREATE TABLE IF NOT EXISTS orders (
    id                 BIGSERIAL    PRIMARY KEY,

    -- 對外顯示的訂單編號，格式：P2P + YYYYMMDD + 8位序號
    order_no           VARCHAR(30)  UNIQUE NOT NULL,

    listing_id         BIGINT       NOT NULL REFERENCES listings(id),
    listing_type       VARCHAR(10)  NOT NULL CHECK (listing_type IN ('buy', 'sell')),

    -- 掛買：seller 接單；掛賣：buyer 接單
    seller_id          BIGINT       NOT NULL REFERENCES app_users(id),
    buyer_id           BIGINT       NOT NULL REFERENCES app_users(id),

    -- 幣種（冗餘自 listing，方便查詢）
    crypto_currency    VARCHAR(20)  NOT NULL DEFAULT 'USDT',
    fiat_currency      VARCHAR(10)  NOT NULL DEFAULT 'TWD',

    -- 金額
    crypto_amount         NUMERIC(20, 8) NOT NULL,          -- 交易 USDT 數量
    price                 NUMERIC(20, 4) NOT NULL,          -- 成交單價
    fiat_amount           NUMERIC(20, 4) NOT NULL,          -- 純交易法幣金額（不含手續費）

    -- 手續費明細（第一階段均為 0，建立訂單時由 listing 費率快照計算後寫入）
    --
    -- 平台手續費
    platform_fee_base     NUMERIC(20, 4) NOT NULL DEFAULT 0, -- 平台固定基礎費
    platform_fee_amount   NUMERIC(20, 4) NOT NULL DEFAULT 0, -- 平台比例費金額 = fiat_amount × platform_fee_rate
    -- 收款管道手續費
    payment_fee_base      NUMERIC(20, 4) NOT NULL DEFAULT 0, -- 管道固定基礎費
    payment_fee_amount    NUMERIC(20, 4) NOT NULL DEFAULT 0, -- 管道比例費金額 = fiat_amount × payment_fee_rate
    --
    -- 加總
    total_fee             NUMERIC(20, 4) NOT NULL DEFAULT 0,
    -- total_fee = platform_fee_base + platform_fee_amount + payment_fee_base + payment_fee_amount
    --
    total_amount          NUMERIC(20, 4) NOT NULL DEFAULT 0,
    -- total_amount = fiat_amount + total_fee（買家實際應付金額）

    -- 此筆訂單使用的收款方式（買家付款目標）
    payment_method_id  BIGINT       NOT NULL REFERENCES payment_methods(id),

    -- 狀態
    -- matched    : 已配對，等待買家付款
    -- paid       : 買家已點擊「已付款」
    -- releasing  : 賣家確認收款，系統放行中
    -- completed  : 虛擬貨幣已到買家帳戶
    -- cancelled  : 主動取消（matched 前）
    -- timeout    : 付款超時自動取消
    -- disputed   : 申訴中
    status             VARCHAR(20)  NOT NULL DEFAULT 'matched'
                       CHECK (status IN (
                           'matched', 'paid', 'releasing', 'completed',
                           'cancelled', 'timeout', 'disputed'
                       )),

    -- 時間戳記
    payment_deadline   TIMESTAMPTZ  NOT NULL,            -- 付款截止時間
    paid_at            TIMESTAMPTZ,                      -- 買家點「已付款」時間
    confirmed_at       TIMESTAMPTZ,                      -- 賣家確認收款時間
    completed_at       TIMESTAMPTZ,                      -- 訂單完成時間
    cancelled_at       TIMESTAMPTZ,                      -- 取消時間
    cancel_reason      TEXT,                             -- 取消原因

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
-- 4. 托管記錄（escrow_records）
--    記錄每一次鎖幣 / 放行 / 退還的動作
-- ============================================================
CREATE TABLE IF NOT EXISTS escrow_records (
    id              BIGSERIAL    PRIMARY KEY,
    order_id        BIGINT       NOT NULL REFERENCES orders(id),

    crypto_currency VARCHAR(20)  NOT NULL DEFAULT 'USDT',
    amount          NUMERIC(20, 8) NOT NULL,

    -- lock    : 鎖定（掛賣發布時 or 掛買接單時）
    -- release : 放行給買家（訂單完成）
    -- refund  : 退還給賣家（取消 / 超時）
    action          VARCHAR(20)  NOT NULL CHECK (action IN ('lock', 'release', 'refund')),

    -- pending / completed / failed
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'completed', 'failed')),

    -- 預留區塊鏈 tx hash 欄位（第一階段為 NULL，內部帳務用）
    tx_ref          VARCHAR(100),

    remark          TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_escrow_order_id ON escrow_records(order_id);


-- ============================================================
-- 5. 訂單狀態日誌（order_status_logs）
--    記錄每一次狀態轉換，包含操作者與備註，不可刪改
-- ============================================================
CREATE TABLE IF NOT EXISTS order_status_logs (
    id            BIGSERIAL   PRIMARY KEY,
    order_id      BIGINT      NOT NULL REFERENCES orders(id),

    from_status   VARCHAR(20),                    -- 初始建立時為 NULL
    to_status     VARCHAR(20) NOT NULL,

    -- 操作者類型：buyer / seller / admin / system
    operator_type VARCHAR(20) NOT NULL
                  CHECK (operator_type IN ('buyer', 'seller', 'admin', 'system')),
    operator_id   BIGINT,                         -- 對應 app_users.id 或 backend_users.id

    remark        TEXT,

    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
    -- 此表為 append-only，不設 updated_at
);

CREATE INDEX IF NOT EXISTS idx_order_status_logs_order_id ON order_status_logs(order_id);
