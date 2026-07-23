# 出金（Wallet Withdrawal）

## 子領域

| 渠道 | 幣種 | 處理方式 |
|------|------|---------|
| 法幣出金（銀行轉帳）| TWD | 後台人工審核，審核後手動轉帳 |
| 加密貨幣出金（Tron） | TRX / USDT | 自動批次廣播，TronWithdrawJob 執行 |

---

## 法幣出金（FiatWithdrawal）

### 狀態機

| 狀態 | 說明 |
|------|------|
| `pending` | 用戶申請，等待後台審核 |
| `approved` | 後台審核通過，已手動完成銀行轉帳 |
| `rejected` | 後台審核拒絕，TWD 餘額退還用戶 |

```
pending → approved  (後台 review action=approve)
pending → rejected  (後台 review action=reject，退還餘額)
```

### 不變式

1. **申請時扣除可用餘額**：用戶申請出金時即刻從 available_balance 扣除申請金額（凍結）。
2. **approved 不退款**：approved 後不可撤銷。
3. **rejected 退款**：rejected 時系統將申請金額退回 available_balance。
4. **用戶銀行帳號**：出金申請必須包含目標銀行帳號資訊（用戶自填），後台審核時可見。
5. **後台審核者標記**：approved/rejected 需記錄操作的後台帳號 ID（審計用）。

### 邊界條件

**BC-1：用戶申請金額超過可用餘額**
- 規則：申請前驗證 available_balance >= amount，不足回傳 400 "insufficient balance"。

**BC-2：審核中用戶再次申請出金**
- 情境：第一筆 pending，用戶又發起第二筆。
- 規則：目前未限制（多筆 pending 並行）。可加入風控：每用戶同時最多 1 筆 pending（待討論）。

**BC-3：後台 approve 後銀行轉帳失敗**
- 情境：後台點擊 approve 後，實際銀行轉帳因帳號錯誤退票。
- 規則：目前系統不追蹤銀行轉帳結果，approve 為最終狀態。退票情況需人工處理（待補充流程）。

**BC-4：同一筆 fiat_withdrawal 被雙擊 approve**
- 情境：後台人員快速雙擊確認。
- 規則：review API 先查詢狀態，已非 `pending` 時回傳 400 "already reviewed"。

---

## 加密貨幣出金（CryptoWithdrawal）

### 狀態機

| 狀態 | 說明 |
|------|------|
| `pending` | 用戶申請，等待 TronWithdrawJob 處理 |
| `broadcasting` | Job 已廣播交易至鏈上，等待確認 |
| `confirmed` | 鏈上確認，出金完成 |
| `failed` | 廣播失敗或鏈上交易失敗 |

```
pending → broadcasting  (TronWithdrawJob 廣播)
broadcasting → confirmed  (鏈上確認)
broadcasting → failed     (鏈上失敗)
pending → failed          (私鑰錯誤、餘額不足等）
```

### 不變式

1. **申請時扣除可用餘額**：用戶申請時即刻從 available_balance 扣除（預防雙重提領）。
2. **私鑰安全**：HotWalletPrivateKey 僅存於環境變數，不 commit 至 git。
3. **手續費預留**：出金金額需包含鏈上手續費或系統另外收取手續費（待確認）。
4. **地址驗證**：目標地址必須是合法的 Tron 地址格式（以 T 開頭的 Base58 地址）。

### 邊界條件

**BC-1：廣播後私鑰輪換**
- 情境：TronWithdrawJob 廣播交易後，熱錢包私鑰被更換。
- 規則：broadcasting 狀態的交易由 Scanner Job 追蹤確認，不依賴私鑰。

**BC-2：熱錢包 USDT 餘額不足**
- 情境：平台熱錢包餘額低於用戶提領金額。
- 規則：TronWithdrawJob 處理前驗證熱錢包餘額，不足時將 status→failed 並通知管理員補充熱錢包。

**BC-3：鏈上費用（TRX）不足**
- 情境：廣播 USDT 交易時需要消耗 TRX 作為手續費，熱錢包 TRX 不足。
- 規則：廣播前驗證 TRX 餘額，不足時暫停出金並告警。

---

## API 摘要

### 法幣出金申請
`POST /app/wallets/fiat/withdraw`

請求：`{ amount: string, currency: "TWD", bankCode: string, bankAccount: string, accountName: string }`
回應：FiatWithdrawal 記錄（status=pending）

### 加密貨幣出金申請
`POST /app/wallets/crypto/withdraw`

請求：`{ currency: "USDT", amount: string, toAddress: string }`
回應：CryptoWithdrawal 記錄（status=pending）

### 後台法幣出金審核
`PUT /backend/fiat-withdrawals/:id/review`

請求：`{ action: "approve" | "reject", reason: string }`
- approve：更新 status→approved
- reject：更新 status→rejected，退還 available_balance

---

## 相關檔案

| 層級 | 路徑 |
|------|------|
| App Logic | `backend/internal/logic/appwalletlogic.go` |
| Backend Logic | `backend/internal/logic/backendwithdrawlogic.go` |
| Tron Withdraw Job | `backend/internal/job/tron_withdraw_job.go` |
| Model | `backend/internal/model/fiatwithdrawal.go`, `cryptowithdrawal.go` |
| Handler | `backend/internal/handler/backendwithdrawhandler.go` |
