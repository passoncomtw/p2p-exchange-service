# 掛單生命週期（Trade/Listing Lifecycle）

## 狀態機

### 合法狀態

| 狀態 | 說明 |
|------|------|
| `active` | 掛單上架中，可被吃單 |
| `paused` | 暫停，不可被吃單，掛單者可重新啟用 |
| `completed` | 加密貨幣剩餘量歸零，自動結束 |
| `cancelled` | 掛單者手動取消 |

### 狀態轉換圖

```
active ⇌ paused          （掛單者手動切換，預留 API）
active → completed        （remainingAmount = 0 時系統自動設定）
active → cancelled        （掛單者手動取消，remainingAmount > 0）
paused → cancelled        （掛單者從暫停狀態取消）
```

### 觸發者對照

| 轉換 | 觸發者 | API Endpoint |
|------|--------|-------------|
| 建立（active） | 掛單者 | `POST /app/listings` |
| active → cancelled | 掛單者 | `PUT /app/listings/:id/cancel` |
| active → completed | 系統（訂單 confirm 觸發） | 內部邏輯 |
| active ⇌ paused | 掛單者 | 待實作 |

---

## 不變式

1. **剩餘量非負**：`remainingAmount >= 0`，且 `remainingAmount <= totalAmount`。
2. **completed 時剩餘量為零**：掛單狀態轉為 `completed` 當且僅當 `remainingAmount == 0`。
3. **cancelled 時資產已歸還**：取消掛單時，若掛單者有凍結資產（sell listing），需先確認無進行中的訂單，或等待訂單結清後才可取消。
4. **sell listing 必須綁定銀行帳號**：type=sell 的掛單建立時，掛單者必須有至少一筆 paymentMethod 記錄。
5. **掛單類型不可更改**：建立後 type（buy/sell）、cryptoCurrency、fiatCurrency 均不可修改。

---

## 邊界條件

### BC-1：取消 active 掛單時有進行中的訂單
- 情境：掛單 A 有一筆狀態為 `matched` 的訂單，掛單者嘗試取消 A。
- 規則：取消前需檢查是否有非終態訂單（matched/paid/releasing/disputed）。若有，拒絕取消並回傳錯誤提示。
- 現況：目前 cancel API 未明確檢查此條件（待補強）。

### BC-2：多個吃單同時消耗 remainingAmount 至零
- 情境：高並發下，兩筆吃單同時成立，合計消耗量恰好等於剩餘量。
- 規則：每筆訂單 confirm 時在 TransactCtx 內執行 `remainingAmount - amount`，若結果為 0 則同步將 listing.status 設為 `completed`。Redis lock per listing 防止並發超賣。

### BC-3：buy listing 的銀行帳號來源
- 情境：buy listing 掛單者本身不需要銀行帳號（買方只需匯款給賣方）。
- 規則：buy listing 吃單時（賣方吃單），系統自動取賣方第一筆 paymentMethod。若賣方無 paymentMethod，拒絕吃單。

### BC-4：掛單建立時 totalAmount 精度問題
- 情境：加密貨幣金額（USDT）為小數，`totalAmount` 以 Decimal/string 儲存。
- 規則：比較 remainingAmount 與 0 必須使用 Decimal 比較，不可用浮點 `== 0`。

### BC-5：sell listing 掛單者凍結資產時機
- 情境：建立 sell listing 時是否立即凍結加密貨幣？
- 規則：目前設計為**訂單建立（吃單）時**才凍結，非掛單建立時凍結。掛單本身不鎖資產，允許掛單者的 USDT 餘額低於 totalAmount（接受部分成交）。

---

## API 摘要

### 建立掛單
`POST /app/listings`

- 前置條件：
  - type=sell：掛單者有至少一筆 paymentMethod
  - 加密貨幣、法幣幣種為系統支援的幣種
- 副作用：建立 listing（status=active）
- 不凍結資產（等待吃單時才凍結）

### 取消掛單
`PUT /app/listings/:id/cancel`

- 前置條件：掛單者本人，listing 狀態為 `active` 或 `paused`
- 副作用：status→cancelled
- 待補強：檢查是否有進行中的訂單

### 查詢市場掛單
`GET /app/listings`

- 回傳 active 的掛單列表，支援 type/currency 篩選

### 查詢自己的掛單
`GET /app/listings/mine`

- 回傳目前登入用戶的所有掛單

---

## 相關檔案

| 層級 | 路徑 |
|------|------|
| Handler | `backend/internal/handler/applistinghandler.go` |
| Logic | `backend/internal/logic/applistinglogic.go` |
| Model | `backend/internal/model/listing.go` |
