# P2P Exchange 前端規格文件 v1

本文件涵蓋 Web(使用者端 + 後台)與 App(使用者端)的 v1 規格,並與既有後端服務對齊。

v1 只做一件事的完整閉環:使用者可建立掛單(買幣/賣幣),且後台能透過後端取得並顯示該筆訂單的完整資訊與正確狀態。

---

## 1. 範圍與決策

### 已確定方向(沿用既有、避免過度工程化)

- 後端:沿用既有 Go(go-zero)服務與 PostgreSQL;v1 掛單直接對應既有 `listings` 表,僅補最小端點與一個欄位。
- Web:沿用既有 Vite + React + MUI v9 + react-i18next 技術棧(非 Tailwind)。
- App:沿用既有 Expo + React Native + RN StyleSheet(非 NativeWind),補上 react-i18next。
- 共用核心 `/shared`:純 TypeScript,供 Web 與 App 匯入。
- Node 版本:統一 Node 22 LTS。
- 訂單 id:沿用既有自增主鍵(BIGSERIAL),API 以字串型別輸出。

### 不包含(v1 明確排除)

真實登入/權限、真實金流、撮合/媒合引擎、托管(escrow)、聊天室、KYC、訂單限額。
i18n 架構保留,v1 僅提供 zh-TW。

### 預設值

- 幣別/法幣:USDT / TWD(下拉結構保留)。
- 付款方式:銀行轉帳(bank_transfer)/ 超商代碼(convenience_store)。
- 目前使用者固定 `demo_user`(無登入)。

---

## 2. 資料模型

### 領域模型 Order(API 契約)

| 欄位 | 型別 | 說明 |
|------|------|------|
| id | string | 後端自增主鍵,序列化為字串 |
| type | "buy" \| "sell" | 買幣 / 賣幣 |
| asset | string | v1: USDT |
| fiat | string | v1: TWD |
| price | number | 單價(法幣/每單位幣) |
| quantity | number | 數量 |
| totalAmount | number | = price × quantity(自動計算) |
| paymentMethod | string | bank_transfer / convenience_store |
| status | "open" \| "completed" \| "cancelled" | 訂單狀態 |
| createdBy | string | 目前固定 demo_user;seed 為其他使用者 |
| createdAt / updatedAt | string | ISO 字串 |

### 與既有 DB 的對應(沿用 listings 最小改造)

v1 掛單沿用既有 `listings` 表。對應關係:

| API(Order) | DB(listings) | 備註 |
|-------------|---------------|------|
| id | id | BIGSERIAL |
| type | type | buy / sell |
| asset | crypto_currency | 預設 USDT |
| fiat | fiat_currency | 預設 TWD |
| price | price | |
| quantity | total_amount | listings.total_amount 為掛單數量 |
| totalAmount | (計算) | price × quantity,不另存 |
| paymentMethod | payment_method_label | v1 新增欄位(migration 003) |
| status | status | active↔open;completed/cancelled 兩端一致 |
| createdBy | app_users.username | 透過 user_id 關聯 |

`min_order_fiat`、`max_order_fiat`、手續費、`remaining_amount`、`payment_method_id` 等既有欄位於 v1 以預設值填入(remaining = quantity、min = 0、max = price × quantity、費率 = 0、payment_method_id = NULL),不影響 v1 行為。

### 狀態機

```
open ──(使用者端取消)──▶ cancelled
open ──(後台標記完成)──▶ completed
```

- open(待成交):建立後初始狀態(DB = active)。
- cancelled(已取消):使用者於「我的掛單」取消 open 訂單。
- completed(已完成):Web 後台於詳情標記完成。
- 其餘轉換不開放(伺服器端與 shared 狀態機雙重把關)。

### 種子資料

seed 寫入 `demo_user` 與三位 `trader_*` 帳號,並植入 4 筆他人掛單(createdBy ≠ demo_user),涵蓋 open / completed / cancelled,使後台呈現匯總樣貌。

---

## 3. API 契約(v1,免登入)

統一回傳格式:`{ code, message, data }`,成功 code = 0。

| 方法 | 路徑 | 說明 | 回傳 data |
|------|------|------|-----------|
| POST | /v1/orders | 建立掛單(createdBy = demo_user) | Order |
| GET | /v1/orders/mine | demo_user 的掛單 | { list: Order[] } |
| POST | /v1/orders/:id/cancel | 取消 open 掛單 | { ok: true } |
| GET | /v1/admin/orders?status= | 全部掛單,可依狀態篩選 | { list: Order[] } |
| GET | /v1/admin/orders/:id | 單筆詳情 | Order |
| POST | /v1/admin/orders/:id/complete | open → completed | { ok: true } |

建立請求本體:`{ type, asset, fiat, price, quantity, paymentMethod }`。

伺服器端驗證:type ∈ {buy, sell}、asset = USDT、fiat = TWD、price > 0、quantity > 0、paymentMethod ∈ {bank_transfer, convenience_store}。

---

## 4. shared 共用核心(/shared)

純 TypeScript,無平台相依。Web 透過 Vite alias `@shared`、App 透過 Metro alias + watchFolders 匯入。

- `domain/order.ts`:Order 型別、列舉常數、calcTotalAmount。
- `domain/status.ts`:狀態機(canTransition / canCancel / canComplete)。
- `validation/order.ts`:validateCreateOrder(回傳 i18n 訊息鍵)。
- `dto/order.ts`:API DTO 與回傳信封型別。
- `api/client.ts`:fetch 實作的 API client(Web 與 App 共用,僅 base URL 不同)。
- `i18n/`:訊息鍵與 zh-TW 語系。

設計 token 不放在 shared,改由各平台各自擁有:Web 定義於 `frontend/web/src/theme`,App 定義於 `frontend/app/src/theme`(值維持一致以保留視覺語言)。

---

## 5. 設計 token(萃取自既有頁面 / DESIGN.md)

| 類別 | 重點值 |
|------|--------|
| 主色 | primary #FFC107、hover #FFB300、disabled #FFE082、deep #FF8F00 |
| 文字 | primary #333、secondary #666、tertiary #999、placeholder #BFBFBF |
| 背景 | content #F5F5F7、card #FFFFFF |
| 邊框 | input #D9D9D9、card #EBEBEB |
| 狀態色(v1) | open #FF9800、completed #4CAF50、cancelled #9E9E9E |
| 買/賣色 | buy #4CAF50、sell #F44336 |
| 字級 | body 13、page-title 16/600、button 14/500 |
| 間距 | 4px 基數 |
| 圓角 | 4px |

買/賣與狀態以不同樣式區分:類型為實心標籤,狀態為淡底 + 語意色外框 + 圓點。

---

## 6. 頁面與互動

### Web(使用者端 + 後台)

| 路徑 | 頁面 | 說明 |
|------|------|------|
| / | 掛單頁 | 買/賣切換、表單、總額自動計算、驗證、送出建立 open 訂單 → 導向我的掛單 |
| /my-orders | 我的掛單 | 表格顯示自己訂單與狀態;open 可取消 → cancelled |
| /admin | 後台訂單管理 | 讀全部訂單、狀態篩選(全部/open/completed/cancelled)、點列進詳情 |
| /admin/orders/:id | 訂單詳情 | 完整資訊與正確狀態;open 可標記完成 → completed |

使用者端採頂部導覽 + 置中內容;後台沿用既有側邊欄 + Header 版型。

### App(使用者端)

| Tab | Screen | 說明 |
|-----|--------|------|
| 掛單 | V1CreateOrderScreen | 買/賣切換、表單與驗證、總額計算、呼叫後端建立 open 訂單 |
| 我的掛單 | V1MyOrdersScreen | 卡片式列表顯示狀態;open 可取消 → cancelled;下拉重新整理 |

底部 tab 導覽切換「掛單 / 我的掛單」。

---

## 7. i18n

- 全部 UI 文案走 i18n 訊息鍵,不寫死字串。
- 語系檔放 shared(order 命名空間),v1 僅 zh-TW。
- Web:react-i18next,合併既有後台語系與 shared 訂單語系。
- App:react-i18next,固定 zh-TW(未來可接 expo-localization 讀取裝置語系)。

---

## 8. 技術約束

- Web:Vite + React + MUI v9 + react-router + react-i18next;API 透過 shared fetch client。
- App:Expo + React Native + RN StyleSheet + React Navigation + react-i18next;API 透過 shared fetch client。
- 後端:go-zero;訂單沿用 listings;伺服器端驗證輸入;統一回傳信封。
- 資料存取一律透過後端 REST,不以 localStorage 作為主要資料源。
- Node 22 LTS(.nvmrc 與各 package.json engines 統一)。

---

## 9. 驗收條件

### Web
- 可選買/賣、填表送出、成功建立訂單,且有基本驗證。
- 建立後出現在「我的掛單」,狀態為 open。
- 可取消 open 訂單,狀態變 cancelled。
- 後台讀全部訂單,可見完整資訊與正確狀態,且與使用者端一致。
- 後台可將 open 標記為 completed。
- `npm install && npm run dev` 可啟動操作。

### App
- 可選買/賣、填表送出、成功建立訂單,且有基本驗證。
- 建立後出現在「我的掛單」,狀態為 open,可取消為 cancelled。
- `npx expo start` 可於模擬器或裝置操作。

### 共同
- 狀態僅 open / completed / cancelled。
- 設計 token 由各平台各自擁有(Web theme / App theme),值一致使視覺語言一致;買/賣與狀態標籤有明確視覺區分。
- 跨平台:App 建立的訂單,Web 後台可透過後端取得並顯示且狀態正確。
- UI 文案皆走 i18n(v1 僅 zh-TW)。
