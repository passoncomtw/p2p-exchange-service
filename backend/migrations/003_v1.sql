-- ============================================================
-- 003_v1.sql
-- v1 最小掛單：沿用 listings 表，補上 v1 所需的付款方式字串欄位。
--
-- 設計原則（沿用 listings 最小改造）：
--   - v1「掛單」直接對應既有 listings 表。
--   - v1 付款方式為簡單字串（bank_transfer / convenience_store），
--     不綁定 payment_methods 表，故新增 payment_method_label 欄位。
--   - 狀態沿用既有 CHECK：v1 的 open 對應 DB 的 active，
--     completed / cancelled 兩端一致，故毋須變更 status 約束。
-- ============================================================

ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS payment_method_label VARCHAR(50);

COMMENT ON COLUMN listings.payment_method_label IS
    'v1 付款方式字串：bank_transfer（銀行轉帳）/ convenience_store（超商代碼）';
