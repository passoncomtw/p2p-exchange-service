# 產品規格書 — P2P Exchange Service

本文件描述截至目前已實作的產品功能，涵蓋資料庫、後端 API、Web 後台、App 前端。

---

## 一、系統架構

| 層級 | 技術選型 |
|------|---------|
| 後端框架 | go-zero (REST) |
| 資料庫 | PostgreSQL (pgx driver) |
| Web 後台 | React 19 + Redux Toolkit + Redux-Saga + MUI + Vite 8 |
| App 前端 | React Native (Expo SDK 56) + Redux Toolkit + React Navigation |
| 建置發版 | EAS Build → TestFlight (iOS) / Google Play Internal (Android) |

---

## 二、資料庫結構

### 2.1 app_users（App 使用者）

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| username | VARCHAR(50) UNIQUE NOT NULL | 帳號 |
| password_hash | VARCHAR(255) NOT NULL | 密碼雜湊 |
| email | VARCHAR(255) | 電子郵件（可為 NULL） |
| created_at | TIMESTAMPTZ | 建立時間 |
| updated_at | TIMESTAMPTZ | 更新時間 |

### 2.2 backend_users（後台管理員）

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| username | VARCHAR(50) UNIQUE NOT NULL | 帳號 |
| password_hash | VARCHAR(255) NOT NULL | 密碼雜湊 |
| email | VARCHAR(255) | 電子郵件（可為 NULL） |
| role | VARCHAR(50) DEFAULT 'admin' | 角色 |
| created_at | TIMESTAMPTZ | 建立時間 |
| updated_at | TIMESTAMPTZ | 更新時間 |

### 2.3 payment_methods（收款方式）

歸屬 app_users（賣家綁定），一人可多筆。

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| user_id | BIGINT FK → app_users | 所屬使用者 |
| type | VARCHAR(50) DEFAULT 'bank_transfer' | 收款類型（目前僅 bank_transfer） |
| bank_name | VARCHAR(100) NOT NULL | 銀行名稱 |
| account_name | VARCHAR(100) NOT NULL | 戶名 |
| account_number | VARCHAR(50) NOT NULL | 帳號 |
| is_active | BOOLEAN DEFAULT TRUE | 是否啟用（軟刪除用） |
| created_at | TIMESTAMPTZ | 建立時間 |
| updated_at | TIMESTAMPTZ | 更新時間 |

### 2.4 listings（掛單）

使用者發布的買/賣掛單。

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| user_id | BIGINT FK → app_users | 掛單者 |
| type | VARCHAR(10) | buy / sell |
| crypto_currency | VARCHAR(20) DEFAULT 'USDT' | 加密貨幣幣種 |
| fiat_currency | VARCHAR(10) DEFAULT 'TWD' | 法幣幣種 |
| total_amount | NUMERIC(20,8) | 掛單總量（USDT） |
| remaining_amount | NUMERIC(20,8) | 剩餘可交易量 |
| price | NUMERIC(20,4) | 單價（TWD/USDT） |
| min_order_fiat | NUMERIC(20,4) | 最低交易金額（TWD） |
| max_order_fiat | NUMERIC(20,4) | 最高交易金額（TWD） |
| platform_fee_base | NUMERIC(20,4) DEFAULT 0 | 平台固定基礎費 |
| platform_fee_rate | NUMERIC(8,6) DEFAULT 0 | 平台比例費率 |
| payment_fee_base | NUMERIC(20,4) DEFAULT 0 | 管道固定基礎費 |
| payment_fee_rate | NUMERIC(8,6) DEFAULT 0 | 管道比例費率 |
| payment_time_limit | INT DEFAULT 30 | 付款時限（分鐘） |
| payment_method_id | BIGINT FK → payment_methods | 掛賣單的收款方式（buy 時為 NULL） |
| payment_method_label | VARCHAR(50) | v1 付款方式字串（bank_transfer / convenience_store） |
| status | VARCHAR(20) DEFAULT 'active' | active / paused / completed / cancelled |
| created_at | TIMESTAMPTZ | 建立時間 |
| updated_at | TIMESTAMPTZ | 更新時間 |

### 2.5 orders（訂單）

接單後由系統產生，一筆 listing 可對應多筆 order（部分成交）。

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| order_no | VARCHAR(30) UNIQUE | 訂單編號（P2P + YYYYMMDD + 8位序號） |
| listing_id | BIGINT FK → listings | 來源掛單 |
| listing_type | VARCHAR(10) | buy / sell |
| seller_id | BIGINT FK → app_users | 賣家 |
| buyer_id | BIGINT FK → app_users | 買家 |
| crypto_currency | VARCHAR(20) DEFAULT 'USDT' | 加密貨幣 |
| fiat_currency | VARCHAR(10) DEFAULT 'TWD' | 法幣 |
| crypto_amount | NUMERIC(20,8) | 交易 USDT 數量 |
| price | NUMERIC(20,4) | 成交單價 |
| fiat_amount | NUMERIC(20,4) | 純交易法幣金額 |
| platform_fee_base | NUMERIC(20,4) DEFAULT 0 | 平台固定基礎費 |
| platform_fee_amount | NUMERIC(20,4) DEFAULT 0 | 平台比例費金額 |
| payment_fee_base | NUMERIC(20,4) DEFAULT 0 | 管道固定基礎費 |
| payment_fee_amount | NUMERIC(20,4) DEFAULT 0 | 管道比例費金額 |
| total_fee | NUMERIC(20,4) DEFAULT 0 | 手續費合計 |
| total_amount | NUMERIC(20,4) DEFAULT 0 | 買家實際應付金額 |
| payment_method_id | BIGINT FK → payment_methods | 收款方式 |
| status | VARCHAR(20) DEFAULT 'matched' | matched / paid / releasing / completed / cancelled / timeout / disputed |
| payment_deadline | TIMESTAMPTZ | 付款截止時間 |
| paid_at | TIMESTAMPTZ | 買家點「已付款」時間 |
| confirmed_at | TIMESTAMPTZ | 賣家確認收款時間 |
| completed_at | TIMESTAMPTZ | 訂單完成時間 |
| cancelled_at | TIMESTAMPTZ | 取消時間 |
| cancel_reason | TEXT | 取消原因 |
| created_at | TIMESTAMPTZ | 建立時間 |
| updated_at | TIMESTAMPTZ | 更新時間 |

### 2.6 escrow_records（托管記錄）

記錄每次鎖幣 / 放行 / 退還。

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| order_id | BIGINT FK → orders | 關聯訂單 |
| crypto_currency | VARCHAR(20) DEFAULT 'USDT' | 幣種 |
| amount | NUMERIC(20,8) | 數量 |
| action | VARCHAR(20) | lock / release / refund |
| status | VARCHAR(20) DEFAULT 'pending' | pending / completed / failed |
| tx_ref | VARCHAR(100) | 預留區塊鏈 tx hash（目前為 NULL） |
| remark | TEXT | 備註 |
| created_at | TIMESTAMPTZ | 建立時間 |
| updated_at | TIMESTAMPTZ | 更新時間 |

### 2.7 order_status_logs（訂單狀態日誌）

Append-only，記錄每次狀態轉換。

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| order_id | BIGINT FK → orders | 關聯訂單 |
| from_status | VARCHAR(20) | 原狀態（初始建立時為 NULL） |
| to_status | VARCHAR(20) NOT NULL | 新狀態 |
| operator_type | VARCHAR(20) | buyer / seller / admin / system |
| operator_id | BIGINT | 操作者 ID |
| remark | TEXT | 備註 |
| created_at | TIMESTAMPTZ | 建立時間 |

---

## 三、後端 API

### 3.1 v1 掛單（免登入，沿用 listings 表）

| Method | Path | 說明 |
|--------|------|------|
| POST | /v1/orders | 建立 v1 掛單 |
| GET | /v1/orders/mine | 我的掛單列表 |
| POST | /v1/orders/:id/cancel | 取消掛單 |
| GET | /v1/admin/orders | 管理員查看所有掛單 |
| GET | /v1/admin/orders/:id | 管理員查看單筆掛單 |
| POST | /v1/admin/orders/:id/complete | 管理員完成掛單 |

### 3.2 App 端（JWT 驗證）

**公開路由：**

| Method | Path | 說明 |
|--------|------|------|
| POST | /app/auth/login | App 登入 |

**需登入路由（JWT: App.AccessSecret）：**

| Method | Path | 說明 |
|--------|------|------|
| GET | /app/profile | 取得個人資料 |
| POST | /app/payment-methods | 新增收款方式 |
| GET | /app/payment-methods | 列出收款方式 |
| DELETE | /app/payment-methods/:id | 刪除收款方式（軟刪除） |
| POST | /app/listings | 建立掛單 |
| GET | /app/listings | 列出掛單（市場） |
| GET | /app/listings/mine | 我的掛單 |
| GET | /app/listings/:id | 掛單詳情 |
| PUT | /app/listings/:id/cancel | 取消掛單 |
| POST | /app/orders | 建立訂單（接單） |
| GET | /app/orders | 列出訂單 |
| GET | /app/orders/:id | 訂單詳情 |
| PUT | /app/orders/:id/pay | 標記已付款 |
| PUT | /app/orders/:id/confirm | 確認收款 |
| PUT | /app/orders/:id/cancel | 取消訂單 |
| PUT | /app/orders/:id/dispute | 發起申訴 |

### 3.3 Backend 後台端（JWT 驗證）

**公開路由：**

| Method | Path | 說明 |
|--------|------|------|
| POST | /backend/auth/login | 後台登入 |

**需登入路由（JWT: Backend.AccessSecret）：**

| Method | Path | 說明 |
|--------|------|------|
| GET | /backend/members | 會員列表（支援 keyword 搜尋、分頁） |
| GET | /backend/dashboard | 儀表板（回傳管理員資訊） |
| GET | /backend/listings | 掛單列表 |
| GET | /backend/orders | 訂單列表（支援 keyword/status 搜尋、分頁、total） |
| PUT | /backend/orders/:id/resolve | 訂單爭議處理（complete / refund） |

---

## 四、Web 後台（管理端）

### 4.1 頁面結構

| 路由 | 頁面 | 說明 |
|------|------|------|
| /login | LoginScreen | 後台登入（帳號 + 密碼） |
| / | DashboardScreen | 首頁儀表板 |
| /members | MemberListScreen | 會員列表 |
| /orders | OrderListScreen | 訂單列表 |

### 4.2 側邊欄

- 會員管理（展開子項）
  - 會員列表
- 訂單管理

### 4.3 會員列表功能

- 關鍵字搜尋（帳號、信箱 ILIKE 模糊搜尋）
- 表格欄位：帳號、電子郵件、建立時間、更新時間
- 分頁（每頁筆數可選 10/20/50，顯示總數與頁數）

### 4.4 訂單列表功能

- 關鍵字搜尋（訂單編號 ILIKE 模糊搜尋）
- 狀態篩選下拉（全部 / 待付款 / 已付款 / 已完成 / 已取消 / 申訴中）
- 表格欄位：訂單編號、類型（買入/賣出）、數位貨幣、法幣金額、單價、狀態（彩色 Chip）、建立時間
- 分頁（每頁筆數可選 10/20/50，顯示總數與頁數）

### 4.5 多語系

支援繁體中文（zh-TW）與簡體中文（zh-CN）。

---

## 五、App 前端（使用者端）

### 5.1 導航結構

**公開頁面（未登入）：**

| 畫面 | 說明 |
|------|------|
| Login | 登入頁 |
| Register | 註冊頁 |

**已登入 — 底部 Tab：**

| Tab | 畫面 | 說明 |
|-----|------|------|
| 錢包 | WalletScreen | 錢包首頁 |
| 交易 | TradeScreen | 交易市場（買/賣掛單列表） |
| 掛單 | OrdersScreen | 我的掛單管理 |
| 訂單 | OrderListScreen | 我的訂單列表 |
| 我 | ProfileScreen | 個人資料 |

**已登入 — Stack 頁面：**

| 畫面 | 說明 |
|------|------|
| CreateOrderBuy / CreateOrderSell | 建立買/賣單 |
| ConfirmOrder | 確認訂單 |
| OrderDetail | 訂單詳情 |

### 5.2 API 整合

| API 模組 | 對接後端路徑 | 說明 |
|----------|-------------|------|
| authApi | /app/auth/login | 登入 |
| userApi | /app/profile | 個人資料 |
| listingsApi | /app/listings, /app/listings/mine | 掛單 CRUD |
| p2pOrdersApi | /app/orders | 訂單 CRUD 與狀態操作 |
| paymentMethodsApi | /app/payment-methods | 收款方式 CRUD |
| bankCardsApi | — | 銀行卡（前端定義，待實作） |
| ordersApi / v1Orders | /v1/orders | v1 舊版掛單 |

### 5.3 v1 模式

App 另有 v1 導航（v1.tsx），包含簡化版的交易流程：

| 畫面 | 說明 |
|------|------|
| V1LoginScreen | v1 登入 |
| V1TradeMarketScreen | v1 交易市場 |
| V1CreateOrderScreen | v1 建立掛單 |
| V1OrdersScreen | v1 掛單列表 |
| V1MyOrdersScreen | v1 我的掛單 |
| V1OrderDetailScreen | v1 掛單詳情 |
| V1ListingDetailScreen | v1 掛單詳情 |
| V1AddPaymentMethodScreen | v1 新增付款方式 |

---

## 六、預設帳號

| 平台 | 帳號 | 密碼 | 角色 |
|------|------|------|------|
| 後台 | admin001 | admin@1234 | admin |
| App | testdemo001 | a12345678 | 使用者 |
| App | testdemo002 | a12345678 | 使用者 |
| App | testdemo003 | a12345678 | 使用者 |

建立方式：`go run backend/cmd/seed/main.go -f backend/etc/config.yaml`

---

## 七、手續費設計（第一階段均為 0）

每筆訂單的手續費由四個部分組成：

```
total_fee = platform_fee_base + platform_fee_amount + payment_fee_base + payment_fee_amount
total_amount = fiat_amount + total_fee（買家實際應付金額）
```

| 項目 | 說明 |
|------|------|
| platform_fee_base | 平台固定基礎費（TWD） |
| platform_fee_amount | 平台比例費金額 = fiat_amount × platform_fee_rate |
| payment_fee_base | 收款管道固定基礎費（TWD） |
| payment_fee_amount | 收款管道比例費金額 = fiat_amount × payment_fee_rate |

費率設定在 listing 上，建立訂單時快照計算後寫入 order。

---

## 八、訂單狀態流轉

```
matched → paid → releasing → completed
    ↓        ↓
cancelled  disputed → (admin resolve: complete / refund)
    ↑
  timeout
```

| 狀態 | 說明 |
|------|------|
| matched | 已配對，等待買家付款 |
| paid | 買家已點擊「已付款」 |
| releasing | 賣家確認收款，系統放行中 |
| completed | 虛擬貨幣已到買家帳戶 |
| cancelled | 主動取消 |
| timeout | 付款超時自動取消 |
| disputed | 申訴中 |

---

## 九、掛單狀態

| 狀態 | 說明 |
|------|------|
| active | 可接單 |
| paused | 暫停 |
| completed | 額度用盡 |
| cancelled | 主動撤單 |
