-- ============================================================
-- 004_deposits_withdrawals.sql
-- v0.3.0 充值提領資料表 + 帳本類型擴充
-- ============================================================

-- ============================================================
-- 1. 擴充 wallet_ledgers.type CHECK 約束
-- ============================================================
-- 新增 v0.3.0 充提帳本類型：
--   crypto_deposit  — USDT 鏈上充值入帳
--   crypto_withdraw — USDT 鏈上提領出金
--   fiat_deposit    — TWD ECPay 入金
--   fiat_withdraw   — TWD 提領出金（核可後）

ALTER TABLE wallet_ledgers
    DROP CONSTRAINT IF EXISTS wallet_ledgers_type_check;

ALTER TABLE wallet_ledgers
    ADD CONSTRAINT wallet_ledgers_type_check CHECK (type IN (
        'deposit', 'withdraw',
        'freeze', 'unfreeze',
        'transfer_in', 'transfer_out',
        'fee_deduct',
        'crypto_deposit', 'crypto_withdraw',
        'fiat_deposit', 'fiat_withdraw'
    ));

-- ============================================================
-- 2. 鏈上 USDT 充值記錄（crypto_deposits）
-- ============================================================
-- status:
--   pending   — 已偵測到交易，等待區塊確認
--   confirmed — 區塊確認完成，已入帳
--   failed    — memo 無法識別或確認超時

CREATE TABLE IF NOT EXISTS crypto_deposits (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       NOT NULL REFERENCES app_users(id),
    currency     VARCHAR(20)  NOT NULL DEFAULT 'USDT',
    amount       NUMERIC(38, 18) NOT NULL,
    tx_hash      VARCHAR(100) NOT NULL,
    from_address VARCHAR(100) NOT NULL,
    memo         VARCHAR(50),
    status       VARCHAR(20)  NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'confirmed', 'failed')),
    confirmed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (tx_hash)
);

CREATE INDEX IF NOT EXISTS idx_crypto_deposits_user_id   ON crypto_deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_crypto_deposits_status    ON crypto_deposits(status);
CREATE INDEX IF NOT EXISTS idx_crypto_deposits_created_at ON crypto_deposits(created_at DESC);

-- ============================================================
-- 3. 鏈上 USDT 提領記錄（crypto_withdrawals）
-- ============================================================
-- status:
--   pending      — 申請建立，等待 Job 廣播
--   broadcasting — 已廣播，等待區塊確認
--   confirmed    — 已確認到帳，帳本已扣款
--   failed       — 廣播失敗，餘額自動解凍

CREATE TABLE IF NOT EXISTS crypto_withdrawals (
    id           BIGSERIAL    PRIMARY KEY,
    user_id      BIGINT       NOT NULL REFERENCES app_users(id),
    currency     VARCHAR(20)  NOT NULL DEFAULT 'USDT',
    amount       NUMERIC(38, 18) NOT NULL,
    to_address   VARCHAR(100) NOT NULL,
    tx_hash      VARCHAR(100),
    status       VARCHAR(20)  NOT NULL DEFAULT 'pending'
                 CHECK (status IN ('pending', 'broadcasting', 'confirmed', 'failed')),
    broadcast_at TIMESTAMPTZ,
    confirmed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (tx_hash) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX IF NOT EXISTS idx_crypto_withdrawals_user_id    ON crypto_withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_crypto_withdrawals_status     ON crypto_withdrawals(status);
CREATE INDEX IF NOT EXISTS idx_crypto_withdrawals_created_at ON crypto_withdrawals(created_at DESC);

-- ============================================================
-- 4. ECPay TWD 入金記錄（fiat_deposits）
-- ============================================================
-- ecpay_order_no     — ECPay 回傳的訂單編號（RtnMsg 中取得）
-- merchant_trade_no  — 平台產生的唯一交易號（送給 ECPay）
-- status:
--   pending — 已建立，等待使用者付款
--   paid    — ECPay Webhook 確認付款成功
--   failed  — 付款失敗或 Webhook 通知 RtnCode != 1

CREATE TABLE IF NOT EXISTS fiat_deposits (
    id                  BIGSERIAL    PRIMARY KEY,
    user_id             BIGINT       NOT NULL REFERENCES app_users(id),
    currency            VARCHAR(10)  NOT NULL DEFAULT 'TWD',
    amount              NUMERIC(18, 2) NOT NULL,
    ecpay_order_no      VARCHAR(50),
    merchant_trade_no   VARCHAR(50)  NOT NULL,
    status              VARCHAR(20)  NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'paid', 'failed')),
    payment_type        VARCHAR(30),
    paid_at             TIMESTAMPTZ,
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (merchant_trade_no)
);

CREATE INDEX IF NOT EXISTS idx_fiat_deposits_user_id          ON fiat_deposits(user_id);
CREATE INDEX IF NOT EXISTS idx_fiat_deposits_merchant_trade_no ON fiat_deposits(merchant_trade_no);
CREATE INDEX IF NOT EXISTS idx_fiat_deposits_status           ON fiat_deposits(status);

-- ============================================================
-- 5. TWD 提領申請記錄（fiat_withdrawals）
-- ============================================================
-- status:
--   pending  — 申請建立，等待後台審核
--   approved — 後台核可，帳本已扣款（管理員手動匯款）
--   rejected — 後台拒絕，餘額自動解凍

CREATE TABLE IF NOT EXISTS fiat_withdrawals (
    id             BIGSERIAL    PRIMARY KEY,
    user_id        BIGINT       NOT NULL REFERENCES app_users(id),
    currency       VARCHAR(10)  NOT NULL DEFAULT 'TWD',
    amount         NUMERIC(18, 2) NOT NULL,
    bank_code      VARCHAR(10)  NOT NULL,
    bank_account   VARCHAR(30)  NOT NULL,
    account_name   VARCHAR(50)  NOT NULL,
    status         VARCHAR(20)  NOT NULL DEFAULT 'pending'
                   CHECK (status IN ('pending', 'approved', 'rejected')),
    reviewed_by    BIGINT       REFERENCES backend_users(id),
    reviewed_at    TIMESTAMPTZ,
    reject_reason  TEXT,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fiat_withdrawals_user_id    ON fiat_withdrawals(user_id);
CREATE INDEX IF NOT EXISTS idx_fiat_withdrawals_status     ON fiat_withdrawals(status);
CREATE INDEX IF NOT EXISTS idx_fiat_withdrawals_created_at ON fiat_withdrawals(created_at DESC);
