# 掛單發布（Order Posting）

## 狀態機

掛單發布為單次動作，不涉及狀態機，建立後直接進入 `active`。
掛單生命週期詳見 [`core/trade-lifecycle`](../../core/trade-lifecycle/spec.md)。

---

## 不變式

1. **幣種合法性**：`cryptoCurrency` 與 `fiatCurrency` 必須在 `currencies` 表中存在且啟用。
2. **金額正數**：`totalAmount > 0`，`price > 0`，`minOrderFiat <= maxOrderFiat`。
3. **sell listing 須有收款帳號**：type=sell 的掛單者必須有至少一筆 paymentMethod 記錄。
4. **不凍結資產**：掛單建立時不鎖定任何資產，資產凍結發生在吃單（訂單建立）時。
5. **手續費參數**：platformFeeBase/Rate 與 paymentFeeBase/Rate 由平台設定填入，不由用戶自定義（目前均為 0）。

---

## 邊界條件

### BC-1：sell listing 掛單者帳戶 USDT 餘額不足
- 情境：賣方掛出 1000 USDT 但帳戶只有 500 USDT。
- 規則：掛單建立時**不**驗證餘額（不凍結），允許掛單；若吃單時餘額不足，吃單失敗。
- 影響：掛單在市場上可見但無法成交，可能造成用戶體驗差（待討論是否在建立時驗證）。

### BC-2：minOrderFiat > maxOrderFiat
- 情境：用戶誤填 min=10000, max=1000。
- 規則：API 驗證 minOrderFiat <= maxOrderFiat，否則回傳 400。

### BC-3：paymentTimeLimit 為零或過短
- 情境：設置付款期限為 0 分鐘。
- 規則：paymentTimeLimit 必須 >= 系統最低限制（建議 15 分鐘），否則回傳 400。

### BC-4：用戶同時有多筆 active 掛單
- 情境：同一用戶發布 10 筆 sell listing。
- 規則：目前無限制並發掛單數量。未來可加入每用戶最大 active 掛單數的風控規則。

### BC-5：price 精度
- 情境：USDT/TWD 匯率輸入 30.12345678901234（超長小數）。
- 規則：price 以 `NUMERIC(20,8)` 儲存，超過 8 位小數時截斷或四捨五入（需確認 DB schema）。

---

## API 摘要

### 建立掛單
`POST /app/listings`

請求欄位：
| 欄位 | 型別 | 必填 | 說明 |
|------|------|------|------|
| type | string | 是 | `buy` 或 `sell` |
| cryptoCurrency | string | 是 | 如 `USDT` |
| fiatCurrency | string | 是 | 如 `TWD` |
| totalAmount | string | 是 | 加密貨幣總量（Decimal string） |
| price | string | 是 | 每單位加密貨幣的法幣價格 |
| minOrderFiat | string | 是 | 單筆最低法幣金額 |
| maxOrderFiat | string | 是 | 單筆最高法幣金額 |
| paymentTimeLimit | int | 是 | 付款期限（分鐘） |
| paymentMethodID | int64 | sell only | 綁定的收款帳號 ID |

回應：建立的 listing 詳情（含 id, status=active, remainingAmount=totalAmount）

---

## 相關檔案

| 層級 | 路徑 |
|------|------|
| Handler | `backend/internal/handler/applistinghandler.go` |
| Logic | `backend/internal/logic/applistinglogic.go: CreateListingLogic` |
| Model | `backend/internal/model/listing.go` |
| 請求型別 | `backend/internal/types/types.go: CreateListingRequest` |
