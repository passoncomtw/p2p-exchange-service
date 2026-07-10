# 產品規格書 — P2P Exchange Service

本文件描述截至目前已實作的產品功能，並包含 v0.2.0 規劃內容，涵蓋資料庫、後端 API、Web 後台、App 前端。

---

## 一、系統架構

| 層級 | 技術選型 |
|------|---------|
| 後端框架 | go-zero (REST) |
| 資料庫 | PostgreSQL (Neon, pgx driver) |
| Web 後台 | React 19 + Redux Toolkit + Redux-Saga + MUI + Vite 8 |
| App 前端 | React Native (Expo SDK 56) + Redux Toolkit + Redux-Saga + React Navigation |
| API 文件 | Swagger UI (OAS 2.0, 內嵌於後端 /swagger) |

---

## 二、資料庫結構

所有表定義於 `backend/migrations/001_init.sql`。

### 2.1 app_users（App 使用者）

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| username | VARCHAR(50) UNIQUE NOT NULL | 帳號 |
| password_hash | VARCHAR(255) NOT NULL | 密碼雜湊（bcrypt） |
| email | VARCHAR(255) | 電子郵件（可為 NULL） |
| created_at | TIMESTAMPTZ | 建立時間 |
| updated_at | TIMESTAMPTZ | 更新時間 |

### 2.2 backend_users（後台管理員）

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | 自增主鍵 |
| username | VARCHAR(50) UNIQUE NOT NULL | 帳號 |
| password_hash | VARCHAR(255) NOT NULL | 密碼雜湊（bcrypt） |
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

基底路徑：`/`，Port：`8888`，統一回傳格式：`{ code, message, data }`。

### 3.1 App 端

**公開路由：**

| Method | Path | 說明 |
|--------|------|------|
| POST | /app/auth/login | App 使用者登入（回傳 JWT + 使用者資訊） |

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
| PUT | /app/orders/:id/confirm | 確認收款（賣家放行） |
| PUT | /app/orders/:id/cancel | 取消訂單 |
| PUT | /app/orders/:id/dispute | 發起申訴 |

### 3.2 Backend 後台端

**公開路由：**

| Method | Path | 說明 |
|--------|------|------|
| POST | /backend/auth/login | 後台管理員登入（回傳 JWT） |

**需登入路由（JWT: Backend.AccessSecret）：**

| Method | Path | 說明 |
|--------|------|------|
| GET | /backend/dashboard | 儀表板（回傳管理員資訊） |
| GET | /backend/members | 會員列表（支援 keyword 搜尋、分頁） |
| GET | /backend/listings | 掛單列表 |
| GET | /backend/orders | 訂單列表（支援 keyword/status 搜尋、分頁、total） |
| PUT | /backend/orders/:id/resolve | 訂單爭議處理（complete / refund） |

### 3.3 v1 掛單（免登入，沿用 listings 表）

| Method | Path | 說明 |
|--------|------|------|
| POST | /v1/orders | 建立 v1 掛單（createdBy 固定 demo_user） |
| GET | /v1/orders/mine | 我的掛單列表 |
| POST | /v1/orders/:id/cancel | 取消掛單 |
| GET | /v1/admin/orders | 管理員查看所有掛單 |
| GET | /v1/admin/orders/:id | 管理員查看單筆掛單 |
| POST | /v1/admin/orders/:id/complete | 管理員完成掛單 |

---

## 四、Web 後台（管理端）

技術棧：React 19 + Redux Toolkit + Redux-Saga + MUI + Vite 8 + react-router + react-i18next。

### 4.1 頁面結構

| 路由 | 頁面 | 說明 |
|------|------|------|
| /login | LoginScreen | 後台登入（帳號 + 密碼） |
| / | DashboardScreen | 首頁儀表板 |
| /members | MemberListScreen | 會員列表 |
| /orders | OrderListScreen | 訂單列表 |

登入後套用 MainLayout（側邊欄 + Header），受 ProtectedRoute 保護。

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

技術棧：React Native (Expo SDK 56) + Redux Toolkit + Redux-Saga + React Navigation + redux-persist。

State 管理：auth / orders / bankCards 三個 slice，auth 持久化至 AsyncStorage。

### 5.1 導航結構

**公開頁面（未登入）：**

| 畫面 | 說明 |
|------|------|
| LoginScreen | 登入頁（帳號 + 密碼） |
| RegisterScreen | 註冊頁 |

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
| CreateOrderScreen | 建立買/賣掛單（透過 route name CreateOrderBuy / CreateOrderSell 區分） |
| ConfirmOrderScreen | 確認接單（傳入 listingId + cryptoAmount，呼叫 POST /app/orders） |
| OrderDetailScreen | 訂單詳情（顯示完整訂單資訊，依角色顯示操作按鈕） |

### 5.2 API 整合（Saga 流程）

App 已完成從 v1 API 遷移至 P2P API，所有訂單/掛單操作透過 Redux-Saga 處理。

| Saga | API 模組 | 後端路徑 | 說明 |
|------|----------|---------|------|
| createListingSaga | listingsApi.create | POST /app/listings | 建立掛單 |
| createOrderSaga | p2pOrdersApi.create | POST /app/orders | 建立訂單（接單） |
| fetchOrderListSaga | p2pOrdersApi.list | GET /app/orders | 取得訂單列表 |
| markOrderAsPaidSaga | p2pOrdersApi.markPaid | PUT /app/orders/:id/pay | 標記已付款 |
| applyOrderSaga | p2pOrdersApi.confirm | PUT /app/orders/:id/confirm | 確認收款放行 |

### 5.3 前端 API 模組

| 模組 | 對接路徑 | 說明 |
|------|---------|------|
| authApi | /app/auth/login | 登入、登出 |
| userApi | /app/profile | 個人資料 |
| listingsApi | /app/listings, /app/listings/mine, /app/listings/:id, /app/listings/:id/cancel | 掛單 CRUD |
| p2pOrdersApi | /app/orders, /app/orders/:id/pay, /app/orders/:id/confirm, /app/orders/:id/cancel, /app/orders/:id/dispute | 訂單 CRUD 與狀態操作 |
| paymentMethodsApi | /app/payment-methods | 收款方式 CRUD |
| bankCardsApi | — | 銀行卡（前端 store 定義，資料來自登入回傳） |

### 5.4 前端資料型別

**Order（訂單）：**

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | number | 訂單 ID |
| orderNo | string | 訂單編號 |
| listingId | number | 來源掛單 ID |
| listingType | 'buy' \| 'sell' | 掛單類型 |
| sellerId | number | 賣家 ID |
| buyerId | number | 買家 ID |
| cryptoCurrency | string | 加密貨幣 |
| fiatCurrency | string | 法幣 |
| cryptoAmount | number | 交易數量 |
| price | number | 成交單價 |
| fiatAmount | number | 法幣金額 |
| totalFee | number | 手續費合計 |
| totalAmount | number | 買家實付金額 |
| status | string | matched / paid / releasing / completed / cancelled / timeout / disputed |
| paymentDeadline | string | 付款截止時間 |

**ListingItem（掛單）：**

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | number | 掛單 ID |
| userId | number | 掛單者 ID |
| type | 'buy' \| 'sell' | 類型 |
| totalAmount | number | 總量 |
| remainingAmount | number | 剩餘量 |
| price | number | 單價 |
| minOrderFiat | number | 最低金額 |
| maxOrderFiat | number | 最高金額 |
| status | string | active / paused / completed / cancelled |

**CreateListingParams（建立掛單請求）：**

| 欄位 | 型別 | 說明 |
|------|------|------|
| type | 'buy' \| 'sell' | 類型 |
| totalAmount | number | 總量 |
| price | number | 單價 |
| minOrderFiat | number | 最低金額 |
| maxOrderFiat | number | 最高金額 |
| paymentTimeLimit | number（可選） | 付款時限 |
| paymentMethodId | number \| null（可選） | 收款方式（僅 sell） |

**CreateOrderParams（建立訂單請求）：**

| 欄位 | 型別 | 說明 |
|------|------|------|
| listingId | number | 掛單 ID |
| cryptoAmount | number | 交易數量 |

### 5.5 訂單詳情頁邏輯

OrderDetailScreen 的 OrderContent 元件依據當前使用者角色（buyerId / sellerId vs user.id）顯示不同操作按鈕：

| 訂單狀態 | 買家操作 | 賣家操作 |
|---------|---------|---------|
| matched | 標記已付款 | — |
| paid | — | 確認收款放行 |
| 其他狀態 | 僅檢視 | 僅檢視 |

### 5.6 v1 模式（保留但非主流程）

App 保留 v1 導航（v1.tsx），包含簡化版交易流程，對接 /v1/* API：

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

種子資料定義於 `backend/migrations/002_seeds.sql`。
執行方式：`go run backend/cmd/seed/main.go -f backend/etc/config.yaml`（seed 工具額外建立 demo_user、trader_alice/bob/carol 與樣本掛單）。

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
| platform_fee_amount | 平台比例費金額 = fiat_amount x platform_fee_rate |
| payment_fee_base | 收款管道固定基礎費（TWD） |
| payment_fee_amount | 收款管道比例費金額 = fiat_amount x payment_fee_rate |

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

---

## 十、部署架構

| 項目 | 說明 |
|------|------|
| K8s Cluster | DigitalOcean SGP1 (do-sgp1-k8s-1-35-1-do-5-passontw) |
| 後端 Image | ghcr.io/passoncomtw/p2p-exchange-service/backend |
| Namespace | p2p-exchange |
| Pods | p2p-backend (2 replicas) + p2p-web (1 replica) |
| 資料庫 | Neon PostgreSQL (ap-southeast-1) |
| 域名 | p2p-exchange-api.passon.tw（後端 API）|
| Reverse Proxy | Cloudflare → nginx-ingress |
| Swagger | https://p2p-exchange-api.passon.tw/swagger（非 pro mode） |

---

## 十一、v0.2.0 規劃 — 補齊交易閉環

### 11.1 目標

補齊掛單 → 接單 → 付款 → 放行的完整閉環，建立內部帳本（Virtual Balance）支撐餘額流轉，並補上缺失的註冊 API 與付款超時機制。充值/提領留待後續版本處理。

### 11.2 內部帳本架構（Virtual Balance）

採用內部 DB 帳本，不上鏈。支援多國、多幣種、多錢包擴展。

**設計原則：**
- 每位使用者可擁有多個錢包，每個錢包對應一種幣種
- 餘額分為 `available`（可用）與 `frozen`（凍結，交易中鎖定）
- 所有餘額變動透過 `wallet_ledgers` append-only 記錄，不可修改刪除
- `total = available + frozen`，不另存欄位

**餘額流轉：**

```
賣家建立 sell listing
  └→ freeze: 賣家 USDT available → frozen

買家接單（create order）
  └→ 無餘額變動（等待法幣線下付款）

買家標記已付款（pay）
  └→ 無鏈上動作

賣家確認收款（confirm → completed）
  └→ unfreeze + deduct: 賣家 frozen 扣減
  └→ transfer_in: 買家 available 增加

取消 / 超時（cancel / timeout）
  └→ unfreeze: 賣家 frozen → available（退還）
```

### 11.3 新增資料表

**wallets（錢包）**

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | |
| user_id | BIGINT FK → app_users | 所屬使用者 |
| currency | VARCHAR(20) NOT NULL | 幣種（USDT / BTC / TWD / USD ...） |
| currency_type | VARCHAR(10) NOT NULL | crypto / fiat |
| available | NUMERIC(20,8) DEFAULT 0 | 可用餘額 |
| frozen | NUMERIC(20,8) DEFAULT 0 | 凍結餘額 |
| created_at | TIMESTAMPTZ | 建立時間 |
| updated_at | TIMESTAMPTZ | 更新時間 |
| UNIQUE(user_id, currency) | | 一人一幣種一錢包 |

**wallet_ledgers（帳本）**

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | BIGSERIAL PK | |
| wallet_id | BIGINT FK → wallets | 對應錢包 |
| type | VARCHAR(20) NOT NULL | deposit / withdraw / freeze / unfreeze / transfer_in / transfer_out |
| amount | NUMERIC(20,8) NOT NULL | 變動金額（正數） |
| direction | VARCHAR(5) NOT NULL | in / out |
| balance_after | NUMERIC(20,8) NOT NULL | 變動後可用餘額快照 |
| ref_type | VARCHAR(30) | order / listing / admin / system |
| ref_id | BIGINT | 關聯 ID |
| remark | TEXT | 備註 |
| created_at | TIMESTAMPTZ | 建立時間（Append-only，無 updated_at） |

### 11.4 後端新增 API

**App 端（公開）：**

| Method | Path | 說明 |
|--------|------|------|
| POST | /app/auth/register | 使用者註冊（建立帳號並自動建立 USDT 錢包） |

**App 端（需登入）：**

| Method | Path | 說明 |
|--------|------|------|
| GET | /app/wallets | 列出使用者所有錢包與餘額 |
| GET | /app/wallets/:currency/ledgers | 指定幣種的帳本記錄（分頁） |

**Backend 端（需登入）：**

| Method | Path | 說明 |
|--------|------|------|
| POST | /backend/members/:id/deposit | 管理員對指定使用者充值（測試用） |

### 11.5 修改現有流程

| 流程 | 修改點 |
|------|--------|
| 建立 sell listing | 驗證賣家 USDT available 足夠，freeze 對應金額 |
| 接單（create order） | 驗證賣家 frozen 仍足夠 |
| 確認收款（confirm） | unfreeze 賣家 frozen，transfer_in 買家 available |
| 取消訂單（cancel） | unfreeze 退還賣家 available |
| 付款超時（timeout job） | 定時掃描 payment_deadline 過期，自動 cancel 並 unfreeze |

### 11.6 付款超時機制

後端新增定時 Job（go-zero cron 或獨立 goroutine），每分鐘掃描：

```sql
SELECT id FROM orders
WHERE status = 'matched'
  AND payment_deadline < NOW()
```

對每筆超時訂單執行：
1. 更新 status → `timeout`
2. unfreeze 賣家餘額
3. 寫入 wallet_ledger + order_status_log
4. 更新 listing.remaining_amount（退還接單扣減的量）

### 11.7 App 新增畫面

| 畫面 | 說明 |
|------|------|
| WalletScreen（更新） | 顯示各幣種錢包餘額（available / frozen）、帳本記錄列表 |
| RegisterScreen（串接） | 呼叫 POST /app/auth/register，成功後自動登入 |

### 11.8 後台新增功能

| 功能 | 說明 |
|------|------|
| 管理員充值 | 後台會員詳情頁可對指定使用者 USDT 錢包充值（開發測試用） |

### 11.9 與 escrow_records 的關係

| | escrow_records | wallet_ledgers |
|--|----------------|----------------|
| 職責 | 記錄托管事件（lock / release / refund） | 記錄餘額變動明細 |
| 粒度 | 每筆訂單一筆 | 每次餘額變動一筆 |
| 關聯 | ref_type=order + ref_id | ref_type=order + ref_id |
| 未來 | 串接鏈上 tx_ref | 純內部帳務，可退化為快取層 |

### 11.10 v0.2.0 不包含

- 真實充值/提領（鏈上出入金）
- 手續費計算啟用（目前費率仍為 0）
- 法幣錢包（TWD 等）
- KYC / 實名驗證
