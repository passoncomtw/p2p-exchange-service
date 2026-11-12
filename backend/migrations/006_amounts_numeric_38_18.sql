-- ============================================================
-- 006_amounts_numeric_38_18.sql
-- 所有金額、價格、費用與費率欄位統一為 NUMERIC(38, 18)。
-- 皆為擴張精度（小數位與整數位只增不減），既有資料不會被捨入或截斷。
-- ============================================================

ALTER TABLE listings
    ALTER COLUMN price             TYPE NUMERIC(38, 18),
    ALTER COLUMN min_order_fiat    TYPE NUMERIC(38, 18),
    ALTER COLUMN max_order_fiat    TYPE NUMERIC(38, 18),
    ALTER COLUMN platform_fee_base TYPE NUMERIC(38, 18),
    ALTER COLUMN platform_fee_rate TYPE NUMERIC(38, 18),
    ALTER COLUMN payment_fee_base  TYPE NUMERIC(38, 18),
    ALTER COLUMN payment_fee_rate  TYPE NUMERIC(38, 18);

ALTER TABLE orders
    ALTER COLUMN price               TYPE NUMERIC(38, 18),
    ALTER COLUMN fiat_amount         TYPE NUMERIC(38, 18),
    ALTER COLUMN platform_fee_base   TYPE NUMERIC(38, 18),
    ALTER COLUMN platform_fee_amount TYPE NUMERIC(38, 18),
    ALTER COLUMN payment_fee_base    TYPE NUMERIC(38, 18),
    ALTER COLUMN payment_fee_amount  TYPE NUMERIC(38, 18),
    ALTER COLUMN total_fee           TYPE NUMERIC(38, 18),
    ALTER COLUMN total_amount        TYPE NUMERIC(38, 18);

ALTER TABLE fiat_deposits
    ALTER COLUMN amount TYPE NUMERIC(38, 18);

ALTER TABLE fiat_withdrawals
    ALTER COLUMN amount TYPE NUMERIC(38, 18);
