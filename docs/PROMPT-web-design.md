# Prompt:實作 P2P Exchange v1 Web 原型

> 將以下整段內容貼給 Claude(design)。它是一份自足的實作指示;若可存取本 repo,請優先讀取其中引用的檔案。

---

## 角色與目標

你是一位資深設計師轉資深前端工程師。請實作一個 P2P 加密貨幣交易平台「v1」的 **Web 原型**(使用者端 + 後台),串接既有後端服務(非 mock)。沿用既有視覺與技術棧,避免過度工程化(YAGNI / KISS),遵循 Clean Code。所有註解與文案用繁體中文,不使用 emoji。

v1 只做一件事的完整閉環:使用者可建立掛單(買幣/賣幣),且後台能透過後端取得並顯示該筆訂單的完整資訊與正確狀態。

## 先讀以下既有資產(若可存取)

- `docs/SPEC-v1.md` — 前端規格(資料模型、狀態機、API 契約、驗收條件)。
- `docs/HANDOFF-v1.md` — 每個畫面的 token / 元件 / 狀態 / 邊界 / 無障礙規格。
- `DESIGN.md` — 設計系統與 token 來源。
- `shared/src/**` — 共用核心(型別、驗證、狀態機、i18n、API client),**直接匯入,勿重複實作**。
- `frontend/web/src/theme/index.js` — Web 設計 token 與 `statusColor`/`typeColor`(token 不與 App 共用)。

## 技術棧(沿用既有,勿替換)

- Vite + React + MUI v9(以 `sx` 設樣式)+ react-router v7 + react-i18next。
- 資料存取透過 `shared` 的 fetch API client(`createApiClient`),base URL 取自 `import.meta.env.VITE_API_URL`;不要用 localStorage 當主要資料源。
- Node 22。

## 設計 token(Web,定義於 frontend/web/src/theme)

- 主色 `#FFC107`(hover `#FFB300`、disabled `#FFE082`、deep `#FF8F00`)。
- 背景 content `#F5F5F7` / card `#FFFFFF`;文字 `#333/#666/#999`;邊框 `#D9D9D9 / #EBEBEB`;danger `#F44336`。
- 狀態色:open `#FF9800`、completed `#4CAF50`、cancelled `#9E9E9E`(用 `statusColor()`)。
- 類型色:buy `#4CAF50`、sell `#F44336`(用 `typeColor()`)。
- 字級 body 13、page-title 16/600、button 14/500;間距 4px 基數;圓角 4;按鈕高 36。
- 類型標籤=實心;狀態標籤=淡底+語意色外框+圓點,兩者視覺需明確區分。

## 資料模型與狀態機(來自 shared)

Order:`{ id, type('buy'|'sell'), asset('USDT'), fiat('TWD'), price, quantity, totalAmount(=price×quantity), paymentMethod('bank_transfer'|'convenience_store'), status('open'|'completed'|'cancelled'), createdBy, createdAt, updatedAt }`。

狀態機:`open → cancelled`(使用者端取消)、`open → completed`(後台標記完成),其餘不開放。請使用 shared 的 `canCancel` / `canComplete` / `validateCreateOrder` / `calcTotalAmount`,勿自行重寫規則。

## API 契約(免登入,回傳信封 `{ code, message, data }`,成功 code=0)

- `POST /v1/orders` 建立 → Order
- `GET /v1/orders/mine` 我的 → `{ list: Order[] }`
- `POST /v1/orders/:id/cancel` 取消 → `{ ok: true }`
- `GET /v1/admin/orders?status=` 全部(可篩選)→ `{ list: Order[] }`
- `GET /v1/admin/orders/:id` 詳情 → Order
- `POST /v1/admin/orders/:id/complete` 標記完成 → `{ ok: true }`

一律透過 shared API client 呼叫,不要直接寫 fetch/axios。

## 要實作的頁面與路由

使用者端(免登入,簡潔頂部導覽 + 置中內容):

1. `/` 掛單頁:買/賣切換、表單(幣種、法幣、單價、數量、付款方式)、總額自動計算、欄位驗證(必填、price/quantity > 0)、送出建立 open 訂單 → 成功提示並導向 `/my-orders`。
2. `/my-orders` 我的掛單:表格(類型、幣種、單價、數量、總額、付款方式、狀態、建立時間);open 可取消 → cancelled(需確認對話框)。

後台(沿用既有側邊欄 160px + Header 56px 版型):

3. `/admin` 訂單管理:讀全部訂單、狀態 Tabs 篩選(全部/open/completed/cancelled)、點列進詳情。
4. `/admin/orders/:id` 詳情:完整資訊與正確狀態,open 可標記完成 → completed(需確認對話框)。狀態須與使用者端一致。

## 狀態、邊界、無障礙(每頁皆需)

- 所有狀態:default / hover / disabled / submitting / loading / 空狀態 / 錯誤。
- 送出中停用按鈕避免重複送出;成功/失敗以 Snackbar 回饋。
- 空狀態文案:我的掛單「尚無掛單」、後台「尚無訂單」。
- 表單欄位皆有 label;錯誤訊息顯示於對應欄位;狀態色非唯一資訊(輔以文字)。
- 詳情找不到 id 時顯示載入失敗。

## i18n

全部文案走 i18n 鍵(來自 shared 的 `order.*` 命名空間,合併進 web i18n resources),v1 僅 zh-TW,不寫死字串。

## 交付與驗收

- 可 `npm install && npm run dev` 啟動。
- 建立 → 我的掛單可見且為 open → 後台可見且狀態一致 → 取消(cancelled)/標記完成(completed)各驗一次。
- `npm run build` 需成功。
- 程式碼風格:自我說明命名、最小必要註解、優先共用 shared、勿引入多餘相依。

請依序:先確認既有資產與 token,再實作四個頁面,每完成一頁做一次自我驗收(欄位、狀態、API、i18n),最後跑一次端到端驗收與 build。
