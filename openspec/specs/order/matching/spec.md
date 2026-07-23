# 訂單撮合（Order Matching）

## 概述

撮合（Matching）發生在吃單方呼叫 `POST /app/orders` 時。系統驗證掛單條件、計算法幣金額、鎖定資產並建立訂單。

---

## 不變式

1. **掛單必須 active**：撮合前驗證 listing.status = `active`。
2. **剩餘量足夠**：listing.remainingAmount >= 吃單的 cryptoAmount。
3. **sell listing：吃單者（買方）必須有收款帳號**：待補充（目前僅 sell listing 掛單者需要帳號）。
4. **buy listing：吃單者（賣方）必須有銀行帳號**：系統自動取第一筆，若無則拒絕撮合。
5. **fiatAmount 計算精度**：`fiatAmount = cryptoAmount * price`，以 Decimal 計算，捨入規則一致。
6. **Escrow 原子性**：訂單建立與資產凍結必須在同一事務中完成。

---

## 撮合流程

```
1. 驗證 listing 狀態與剩餘量
2. 確定賣方（根據掛單類型）
3. 查詢賣方銀行帳號
4. 計算法幣金額與手續費
5. AcquireLock（賣方錢包）
6. TransactCtx：
   a. Wallet.FreezeInTx（凍結賣方 cryptoAmount）
   b. INSERT orders（status=matched, paymentDeadline=now+paymentTimeLimit）
   c. UPDATE listings SET remaining_amount = remaining_amount - amount
   d. INSERT escrow_records（action=lock）
   e. INSERT order_status_logs（→ matched）
7. 發送通知（買方、賣方）
8. 發布 NATS order.status.changed
```

### 掛單類型對應的角色

| listing.type | 掛單者角色 | 吃單者角色 | 賣方（持有 USDT） | 銀行帳號來源 |
|-------------|-----------|-----------|----------------|------------|
| `sell` | 賣方 | 買方 | 掛單者（listing.userID） | 掛單者的 paymentMethodID |
| `buy` | 買方 | 賣方 | 吃單者（呼叫者） | 吃單者第一筆 paymentMethod |

---

## 邊界條件

### BC-1：吃單量超過 remainingAmount
- 情境：listing 剩餘 50 USDT，吃單者要吃 60 USDT。
- 規則：驗證 cryptoAmount <= listing.remainingAmount，否則回傳 400 "insufficient listing amount"。

### BC-2：buy listing 吃單者（賣方）無銀行帳號
- 情境：吃單者還沒有設定任何 paymentMethod。
- 規則：查詢 paymentMethod 為空時回傳 400 "no payment method"，引導用戶先新增收款帳號。

### BC-3：並發吃單（同一筆 listing 被多人同時吃）
- 情境：listing 剩餘 100 USDT，A 吃 60、B 吃 60 幾乎同時到達。
- 規則：FreezeInTx 與 UPDATE listings 在同一事務中執行，並以 Redis lock per listing 串行化。先到者成功，後到者回傳 409 或餘量不足錯誤。

### BC-4：吃自己的掛單
- 情境：掛單者本人呼叫吃單 API。
- 規則：目前未明確禁止，應補充驗證 listing.userID != 呼叫者 userID，回傳 400。

### BC-5：paymentDeadline 計算時區
- 情境：伺服器與用戶時區不同導致期限顯示偏差。
- 規則：paymentDeadline 以 UTC 儲存，API 回傳 ISO 8601（UTC），前端自行轉換本地時間顯示。

### BC-6：Freeze 成功但 INSERT orders 失敗（事務回滾）
- 情境：DB 在 INSERT orders 時出現約束違反或超時。
- 規則：整個 TransactCtx 回滾，Freeze 操作隨之回滾，賣方資產不受影響。

---

## API 摘要

### 吃單（建立訂單）
`POST /app/orders`

請求欄位：
| 欄位 | 型別 | 必填 | 說明 |
|------|------|------|------|
| listingID | int64 | 是 | 要吃的掛單 ID |
| cryptoAmount | string | 是 | 要成交的加密貨幣數量 |

回應：
| 欄位 | 說明 |
|------|------|
| id | 訂單 ID |
| orderNo | 訂單編號（uuid） |
| status | `matched` |
| fiatAmount | 應付法幣金額 |
| paymentDeadline | 付款截止時間（ISO 8601 UTC） |
| bankInfo | 賣方銀行帳號資訊（帳號、戶名、行碼） |

---

## 相關檔案

| 層級 | 路徑 |
|------|------|
| Handler | `backend/internal/handler/apporderhandler.go` |
| Logic | `backend/internal/logic/apporderlogic.go: CreateOrderLogic` |
| Model | `backend/internal/model/order.go`, `listing.go`, `wallet.go` |
