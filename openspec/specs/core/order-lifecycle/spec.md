# 訂單生命週期（Order Lifecycle）

## 狀態機

### 合法狀態

| 狀態 | 說明 |
|------|------|
| `matched` | 訂單已建立，等待買方付款 |
| `paid` | 買方已標記付款，等待賣方確認 |
| `releasing` | 賣方已確認收款，系統正在釋放加密貨幣（中間態） |
| `completed` | 交易完成，加密貨幣已轉至買方 |
| `cancelled` | 訂單取消，加密貨幣退還賣方 |
| `timeout` | 付款期限超過，系統自動取消 |
| `disputed` | 糾紛中，等待後台仲裁 |

### 狀態轉換圖

```
matched ──[買方 pay]──────────────────> paid
        ──[取消]────────────────────── > cancelled

paid    ──[賣方 confirm]──────────────> releasing ──> completed
        ──[買方/賣方 dispute]──────── > disputed

disputed ──[admin complete]──────────> completed
         ──[admin refund]────────────> cancelled

matched ──[NATS timeout job]─────────> timeout
timeout  是 cancelled 的特殊形式，資產釋放邏輯相同
```

### 觸發者對照

| 轉換 | 觸發者 | API Endpoint |
|------|--------|-------------|
| → `matched` | 買方或賣方（吃單方） | `POST /app/orders` |
| → `paid` | 買方（付款方） | `PUT /app/orders/:id/pay` |
| → `releasing` | 賣方（確認方） | `PUT /app/orders/:id/confirm`（內部自動） |
| → `completed` | 系統（confirm 觸發）或 admin（仲裁） | confirm 邏輯 / `PUT /backend/orders/:id/resolve` |
| → `cancelled` | 買方或賣方（matched 階段）或 admin（仲裁） | `PUT /app/orders/:id/cancel` / resolve(refund) |
| → `timeout` | NATS `order.timeout.check` consumer | 內部 job |
| → `disputed` | 買方或賣方（paid 階段） | `PUT /app/orders/:id/dispute` |

---

## 不變式

1. **Escrow 完整性**：訂單存在期間，賣方錢包中對應加密貨幣的凍結量必須 >= 訂單 `cryptoAmount`（直到 completed 或 cancelled 釋放）。
2. **狀態不可逆**：`completed` 和 `cancelled` 為終態，任何狀態均不得從這兩個狀態往回轉換。
3. **timeout 屬於 cancelled**：`timeout` 觸發後，資產釋放邏輯與 `cancelled` 完全相同（UnfreezeInTx + RestoreAmountInTx）。
4. **dispute 只能從 paid 進入**：`matched` 狀態下不允許發起糾紛（買方尚未付款，沒有爭議標的）。
5. **confirm 前必須是 paid**：賣方確認只能在 `paid` 狀態下觸發，防止重複放幣。
6. **付款期限（paymentDeadline）**：在訂單建立（`matched`）時設定，通常為建立時間 + listing.paymentTimeLimit 分鐘。

---

## 邊界條件

### BC-1：付款期限到期時訂單已進入 paid 狀態
- 情境：用戶剛好在 paymentDeadline 前 1 秒按下付款，timeout job 此時也被觸發。
- 規則：timeout job 只能取消 `matched` 狀態的訂單。若狀態已是 `paid`，跳過處理。
- 實作依據：job 在執行前應先 SELECT FOR UPDATE 確認狀態。

### BC-2：同一用戶同時發出兩個 confirm 請求（網路重試）
- 情境：賣方點擊「確認收款」後網路延遲，用戶重試。
- 規則：confirm 邏輯在 TransactCtx 內先 SELECT 檢查狀態；若已是 `completed`，冪等回傳成功。
- 目前實作：尚未明確處理，`UPDATE orders SET status='completed' WHERE id=$1 AND status='paid'` 的 affected rows 檢查可防止雙重執行。

### BC-3：dispute 發起後 admin 操作延遲，賣方嘗試 confirm
- 情境：訂單進入 `disputed` 後，賣方仍嘗試呼叫 confirm。
- 規則：confirm API 在進入前應驗證訂單狀態為 `paid`，`disputed` 不允許 confirm，回傳 400。

### BC-4：cancel 與 timeout 競速（Race Condition）
- 情境：用戶手動取消與 timeout job 同時觸發。
- 規則：兩個路徑均使用 TransactCtx + `WHERE status='matched'` 條件更新，只有一個能成功（另一個 affected=0）。Redis 鎖確保 Wallet 操作串行。

### BC-5：掛單剩餘量在訂單建立後被其他吃單消耗至負數
- 情境：高並發下多個吃單請求同時到達，listing.remainingAmount 可能超賣。
- 規則：`RestoreAmountInTx` 使用 `WHERE remaining_amount + $amount <= total_amount` 條件；建立訂單時需加鎖或使用 `SELECT FOR UPDATE`。
- 現況：目前透過 Redis lock per listing 防止並發（待確認實作）。

### BC-6：分散式鎖逾時（Lock Acquisition Timeout）
- 情境：confirm 或 cancel 時，AcquireLock 等待超過 10 秒。
- 規則：回傳 `ErrInternal`（500），不修改任何狀態，由用戶重試。

---

## API 摘要

### 建立訂單
`POST /app/orders`

- 前置條件：listing 狀態為 `active`，剩餘量足夠，吃單方有銀行帳號（sell listing 沿用掛單者）
- 副作用：凍結賣方加密貨幣、更新 listing.remainingAmount、建立 EscrowRecord(lock)、寫入 OrderStatusLog
- 回傳：訂單詳情（含 paymentDeadline、對方銀行帳號）

### 標記付款
`PUT /app/orders/:id/pay`

- 前置條件：訂單狀態為 `matched`，呼叫者為買方
- 副作用：更新狀態為 `paid`、記錄 paidAt、發送通知給賣方、發布 NATS `order.status.changed`

### 確認收款
`PUT /app/orders/:id/confirm`

- 前置條件：訂單狀態為 `paid`，呼叫者為賣方
- 副作用：TransactCtx 內完成 status→completed、EscrowRecord(release)、Wallet.TransferInTx（凍結轉移至買方）、OrderStatusLog
- 發布：NATS `order.status.changed`（status=completed）

### 取消訂單
`PUT /app/orders/:id/cancel`

- 前置條件：訂單狀態為 `matched`，呼叫者為買方或賣方
- 副作用：TransactCtx 內 status→cancelled、EscrowRecord(refund)、Wallet.UnfreezeInTx、Listing.RestoreAmountInTx

### 發起爭議
`PUT /app/orders/:id/dispute`

- 前置條件：訂單狀態為 `paid`，呼叫者為買方或賣方
- 副作用：status→disputed、通知後台管理員

### 後台仲裁
`PUT /backend/orders/:id/resolve`

- 前置條件：訂單狀態為 `disputed`
- action=complete：TransactCtx(status→completed, EscrowRecord(release), Wallet.TransferInTx)
- action=refund：TransactCtx(status→cancelled, EscrowRecord(refund), Wallet.UnfreezeInTx, Listing.RestoreAmountInTx)
- 發布：NATS `order.status.changed`

---

## 相關檔案

| 層級 | 路徑 |
|------|------|
| Handler | `backend/internal/handler/apporderhandler.go` |
| Logic | `backend/internal/logic/apporderlogic.go` |
| Backend Logic | `backend/internal/logic/backendorderlogic.go` |
| Model | `backend/internal/model/order.go` |
| Timeout Job | `backend/internal/job/expired_order_job.go` |
| WS 事件 | `backend/internal/job/ws_event_job.go` |
