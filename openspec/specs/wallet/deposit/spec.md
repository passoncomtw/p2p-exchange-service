# 入金（Wallet Deposit）

## 子領域

入金分為兩種渠道，各自獨立：

| 渠道 | 幣種 | 觸發方式 |
|------|------|---------|
| 法幣入金（ECPay） | TWD | 用戶在 ECPay 付款，Webhook 回呼確認 |
| 加密貨幣入金（Tron 鏈上） | TRX / USDT | 鏈上掃描 Scanner Job 自動偵測 |

---

## 法幣入金（ECPay）

### 狀態機

| 狀態 | 說明 |
|------|------|
| `pending` | 已建立入金單，等待用戶付款 |
| `paid` | ECPay Webhook 確認付款成功，餘額已增加 |
| `failed` | 付款失敗或逾時 |

```
pending → paid     (ECPay Webhook 通知付款成功)
pending → failed   (付款失敗/逾時)
```

### 不變式

1. **Webhook 驗簽**：ECPay 回呼必須驗證 HashKey/HashIV 簽名，簽名錯誤一律拒絕。
2. **冪等**：同一筆 ECPay MerchantTradeNo 的 Webhook 可能重複回呼，系統需先查詢 fiat_deposits 狀態，已處理過的跳過。
3. **原子性**：Webhook 收到後，更新 fiat_deposit.status + wallet.available_balance 必須在同一事務中完成。

### 邊界條件

**BC-1：Webhook 重複回呼**
- ECPay 可能在付款後多次發送 Webhook（如網路重試）。
- 規則：SELECT fiat_deposit WHERE merchant_trade_no 若狀態已為 `paid`，直接回傳 OK，不重複加帳。

**BC-2：Webhook 到達時系統維護**
- 情境：DB 連線中斷，Webhook 處理失敗。
- 規則：回傳非 200 讓 ECPay 重試；系統恢復後重新處理。

**BC-3：HashKey/HashIV 洩漏**
- 安全規則：HashKey 與 HashIV 僅存於環境變數，不 commit 至 git。

---

## 加密貨幣入金（Tron 鏈上）

### 狀態機

| 狀態 | 說明 |
|------|------|
| `pending` | 偵測到鏈上交易，等待確認數 |
| `confirmed` | 達到確認數門檻，餘額已增加 |
| `failed` | 交易在鏈上失敗（status=FAIL） |

```
pending → confirmed  (達到確認數)
pending → failed     (鏈上交易失敗)
```

### 不變式

1. **確認數門檻**：必須達到設定的確認數（confirmations >= threshold）才增加餘額。
2. **防重複入帳**：以 txHash + toAddress 為唯一鍵，避免同一鏈上交易重複入帳。
3. **原子性**：INSERT crypto_deposits + wallet available_balance 更新在同一事務。
4. **地址驗證**：只有屬於系統用戶的地址才觸發入帳。

### 邊界條件

**BC-1：同一 txHash 掃描多次**
- Scanner Job 定期掃描，可能多次看到同一 pending 交易。
- 規則：INSERT IGNORE / ON CONFLICT DO NOTHING，以 txHash 防止重複記錄。

**BC-2：鏈上重組（Chain Reorg）**
- 情境：確認數足夠後入帳，隨後鏈上重組導致交易消失。
- 規則：目前未處理（設置足夠高的確認數門檻降低風險）。

**BC-3：TRC20 USDT 精度換算**
- Tron USDT（TRC20）以 6 位小數計算（1 USDT = 1,000,000 sun）。
- 規則：掃描後換算為 8 位小數的 Decimal string 儲存，與系統其他金額格式一致。

---

## API 摘要

### 法幣入金建立
`POST /app/wallets/fiat/deposit`

請求：`{ amount: string, currency: "TWD" }`
回應：`{ paymentUrl: "https://ecpay..." }`

### 加密貨幣入金地址查詢
`GET /app/wallets/crypto/deposit-info`

回應：`{ address: "T...", currency: "USDT", network: "TRC20" }`

### ECPay Webhook（系統內部）
`POST /webhook/ecpay/notify`

- 由 ECPay 主動呼叫，不需 JWT
- 驗簽後更新 fiat_deposit.status → paid 並增加 TWD 餘額

---

## 相關檔案

| 層級 | 路徑 |
|------|------|
| ECPay Webhook | `backend/internal/handler/webhookhandler.go` |
| Fiat Deposit Logic | `backend/internal/logic/appwalletlogic.go` |
| Tron Scanner | `backend/internal/job/tron_scanner_job.go` |
| Model | `backend/internal/model/fiatdeposit.go`, `cryptodeposit.go` |
