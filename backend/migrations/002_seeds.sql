-- ============================================================
-- 002_seeds.sql
-- 開發 / 測試環境種子資料
-- 冪等：重複執行不報錯
-- 密碼均為 a12345678
-- ============================================================

-- ============================================================
-- 1. 後台管理員
-- ============================================================

INSERT INTO backend_users (username, password_hash, email, role)
VALUES ('admin001', '$2a$10$9D/Tw7aaSWESQuHvN2xk6uLnUcLq.3YrAuK7tOfh.egmY9MOFSov.', 'admin001@example.com', 'admin')
ON CONFLICT (username) DO NOTHING;

-- ============================================================
-- 2. App 測試使用者
-- ============================================================

INSERT INTO app_users (username, password_hash, email)
VALUES
    ('testdemo001', '$2a$10$ud.WrVNEaq2pgpNc11cWuOCujWPxOojWwrqgt2A4U/qselpu5mafS', 'testdemo001@example.com'),
    ('testdemo002', '$2a$10$ud.WrVNEaq2pgpNc11cWuOCujWPxOojWwrqgt2A4U/qselpu5mafS', 'testdemo002@example.com'),
    ('testdemo003', '$2a$10$ud.WrVNEaq2pgpNc11cWuOCujWPxOojWwrqgt2A4U/qselpu5mafS', 'testdemo003@example.com'),
    ('seller001',   '$2a$10$ud.WrVNEaq2pgpNc11cWuOCujWPxOojWwrqgt2A4U/qselpu5mafS', 'seller001@example.com'),
    ('buyer001',    '$2a$10$ud.WrVNEaq2pgpNc11cWuOCujWPxOojWwrqgt2A4U/qselpu5mafS', 'buyer001@example.com')
ON CONFLICT (username) DO NOTHING;

-- ============================================================
-- 3. USDT 錢包
-- ============================================================

INSERT INTO wallets (user_id, currency, available_balance, frozen_balance)
SELECT id, 'USDT', 0, 0
FROM app_users
WHERE username IN ('testdemo001', 'testdemo002', 'testdemo003')
ON CONFLICT (user_id, currency) DO NOTHING;

INSERT INTO wallets (user_id, currency, available_balance, frozen_balance)
SELECT id, 'USDT', 1000, 0
FROM app_users WHERE username = 'seller001'
ON CONFLICT (user_id, currency) DO NOTHING;

INSERT INTO wallets (user_id, currency, available_balance, frozen_balance)
SELECT id, 'USDT', 500, 0
FROM app_users WHERE username = 'buyer001'
ON CONFLICT (user_id, currency) DO NOTHING;

-- ============================================================
-- 4. 帳本初始化記錄（type=deposit）
-- 僅在該錢包尚無任何流水時才寫入，避免重複
-- ============================================================

INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after)
SELECT w.id, 'deposit', 1000, 1000
FROM wallets w
JOIN app_users u ON u.id = w.user_id
WHERE u.username = 'seller001' AND w.currency = 'USDT'
  AND NOT EXISTS (
      SELECT 1 FROM wallet_ledgers wl WHERE wl.wallet_id = w.id
  );

INSERT INTO wallet_ledgers (wallet_id, type, amount, balance_after)
SELECT w.id, 'deposit', 500, 500
FROM wallets w
JOIN app_users u ON u.id = w.user_id
WHERE u.username = 'buyer001' AND w.currency = 'USDT'
  AND NOT EXISTS (
      SELECT 1 FROM wallet_ledgers wl WHERE wl.wallet_id = w.id
  );

-- ============================================================
-- 5. seller001 收款方式
-- ============================================================

INSERT INTO payment_methods (user_id, type, bank_name, account_name, account_number, is_active)
SELECT u.id, 'bank_transfer', '台灣銀行', 'Seller 001', '123-456-789012', true
FROM app_users u
WHERE u.username = 'seller001'
  AND NOT EXISTS (
      SELECT 1 FROM payment_methods pm WHERE pm.user_id = u.id
  );

-- ============================================================
-- 6. seller001 sell listing（100 USDT，33.0 TWD/USDT）
-- ============================================================

INSERT INTO listings (
    user_id, type,
    crypto_currency, fiat_currency,
    total_amount, remaining_amount,
    price,
    min_order_fiat, max_order_fiat,
    payment_method_id,
    payment_method_label,
    status
)
SELECT
    u.id, 'sell',
    'USDT', 'TWD',
    100, 100,
    33.0,
    100, 3300,
    pm.id,
    'bank_transfer',
    'active'
FROM app_users u
JOIN payment_methods pm ON pm.user_id = u.id
WHERE u.username = 'seller001'
  AND NOT EXISTS (
      SELECT 1 FROM listings l
      WHERE l.user_id = u.id AND l.status = 'active'
  )
LIMIT 1;
