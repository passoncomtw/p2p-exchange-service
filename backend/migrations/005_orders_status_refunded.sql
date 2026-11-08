-- ============================================================
-- 005_orders_status_refunded.sql
-- orders.status 的 CHECK 約束加入 'refunded'（後台仲裁判賣方勝的終態）。
-- 規格：p2p-dispute「refunded 為終態，orders.status 的資料庫 CHECK 約束 MUST 包含 refunded」。
-- 以 NOT VALID 加上約束後再 VALIDATE，縮短持有 ACCESS EXCLUSIVE 鎖的時間。
-- ============================================================

ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_status_check;

ALTER TABLE orders ADD CONSTRAINT orders_status_check CHECK (status IN (
    'matched', 'paid', 'releasing', 'completed',
    'cancelled', 'timeout', 'disputed', 'refunded'
)) NOT VALID;

ALTER TABLE orders VALIDATE CONSTRAINT orders_status_check;
