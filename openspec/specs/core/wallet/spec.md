# 錢包核心（Wallet Core）

## 狀態機

錢包無「狀態」欄位，但餘額在任意時刻分為兩個部分：

```
total_balance = available_balance + frozen_balance
```

### 餘額操作類型

| 操作 | available | frozen | 說明 |
|------|-----------|--------|------|
| Freeze（凍結） | - amount | + amount | 訂單建立，鎖定賣方資產 |
| Unfreeze（解凍） | + amount | - amount | 訂單取消/退款，歸還資產 |
| TransferIn（轉入） | + amount | 0 | 買方收到加密貨幣 |
| TransferOut（轉出） | - amount | 0 | 直接扣除可用餘額（平台出金） |
| FreezeAndTransfer（凍結→轉移） | seller: 0, buyer: +amount | seller: -amount | Escrow release：凍結直接轉至買方 |
| Deposit（入金） | + amount | 0 | 鏈上確認後增加可用餘額 |
| Withdraw（出金） | - amount | 0 | 提領請求凍結可用餘額（待確認） |

---

## 不變式

1. **非負約束**：任何時刻 `available_balance >= 0` 且 `frozen_balance >= 0`。
2. **凍結不超過可用**：Freeze 操作前必須驗證 `available_balance >= amount`。
3. **原子性**：跨用戶的 Freeze/Transfer 操作必須在同一 TransactCtx 內完成，防止部分成功。
4. **分散式鎖**：任何修改 wallet 的操作（Freeze/Unfreeze/Transfer）必須先 AcquireLock（per user per currency）。多用戶操作按固定順序鎖定（小 ID 先鎖）防止死鎖。
5. **帳本一致性**：每次 wallet 變動必須同步寫入 wallet_ledger 表，ledger 總和應等於 wallet 當前餘額（審計用）。
6. **幣種隔離**：不同幣種的錢包完全獨立，不共享餘額或鎖。

---

## 邊界條件

### BC-1：Freeze 時 available_balance 不足
- 情境：賣方 USDT 餘額為 100，嘗試吃單 150 USDT。
- 規則：Freeze 前 SELECT wallet FOR UPDATE 確認餘額，不足則回傳 `insufficient balance` 錯誤，拒絕建立訂單。

### BC-2：同時兩筆訂單鎖定同一用戶錢包（死鎖風險）
- 情境：用戶 A 與用戶 B 互相吃對方的掛單，兩個 goroutine 同時嘗試以相反順序鎖定錢包。
- 規則：當需要同時持有兩個用戶的鎖時，一律以 userID 升序獲取（小 ID 先）。目前 confirm 邏輯：先鎖 sellerID 再鎖 buyerID（若有此需求需補充保證順序）。

### BC-3：鏈上入金金額精度（Tron USDT）
- 情境：Tron USDT 以 6 位小數計算，後端以 Decimal string 儲存。
- 規則：`amount` 欄位統一以 `NUMERIC(30,8)` 儲存，禁止使用 float64 進行金額計算。

### BC-4：AcquireLock 逾時後資料庫已完成更新
- 情境：lock 等待 10 秒超時，但 goroutine A 已完成 DB 更新卻尚未釋放鎖。
- 規則：AcquireLock 使用 Redis SET NX 加 TTL，鎖本身有過期機制；lock 釋放後由 defer unlock() 執行。逾時回傳錯誤讓呼叫方重試。

### BC-5：TransferInTx 中賣方 frozen 不足（資料不一致）
- 情境：frozen_balance 已在先前操作中異常扣除，confirm 時 frozen 不足以轉移。
- 規則：TransferInTx 應先驗證 `frozen_balance >= amount`，不足時回滾整個事務並回傳告警。

### BC-6：法幣出金申請後加密貨幣出金同時提出
- 情境：用戶同時申請 TWD 提領與 USDT 提領。
- 規則：兩者操作不同幣種的餘額，各自加鎖互不干擾。但 available_balance 需各自足夠。

---

## API 摘要

### 查詢錢包列表
`GET /app/wallets`

- 回傳用戶所有幣種的錢包（available + frozen）

### 查詢帳本
`GET /app/wallets/:currency/ledgers`

- 回傳指定幣種的交易明細（入金/出金/凍結/解凍/轉帳）

### 加密貨幣入金資訊
`GET /app/wallets/crypto/deposit-info`

- 回傳用戶的鏈上收款地址（Tron TRX/USDT）

### 加密貨幣出金
`POST /app/wallets/crypto/withdraw`

- 前置條件：available_balance >= amount + 手續費
- 副作用：建立 CryptoWithdrawal 記錄，由 TronWithdrawJob 批次處理

### 法幣入金（ECPay）
`POST /app/wallets/fiat/deposit`

- 建立 FiatDeposit 記錄，回傳 ECPay 付款連結；Webhook 確認後增加餘額

### 法幣出金（銀行轉帳）
`POST /app/wallets/fiat/withdraw`

- 建立 FiatWithdrawal 記錄（status=pending），等待後台審核

---

## 相關檔案

| 層級 | 路徑 |
|------|------|
| Model | `backend/internal/model/wallet.go` |
| Ledger Model | `backend/internal/model/walletledger.go` |
| Freeze/Unfreeze/Transfer | `wallet.go: FreezeInTx / UnfreezeInTx / TransferInTx` |
| Lock Key | `backend/internal/model/model.go: WalletLockKey` |
| Tron Scanner | `backend/internal/job/tron_scanner_job.go` |
| Tron Withdraw | `backend/internal/job/tron_withdraw_job.go` |
| ECPay Webhook | `backend/internal/handler/webhookhandler.go` |
