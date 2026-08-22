# v1 錢包搬遷 — 剩餘 Slice 實作指南（C / E / H）

## 背景

`backend/internal`（legacy，go-zero 分層）的錢包相關功能正在搬遷到 `backend/cmd/v1-p2p-exchange-service`（Clean Architecture）。已核准的 Plan 共 10 個 slice，Slice 0、A、D、G、I、B、F 已完成並個別 commit（`git log` 可查）。本文件涵蓋剩下三個高風險 slice：**C（Crypto 提領 + Tron Withdraw Job）→ E（ECPay Webhook）→ H（Backend 法幣提領審核）**，依此順序實作即可完成整個 Plan。

三個 slice 都是資金異動或對外公開端點，每個都含至少一個已確認的資金完整性缺陷修正（非原樣搬遷）。實作時請比照 Slice B/F 已建立的慣例：

- 服務層自己開 `db.TransactCtx`，在 session 內組合多個 repository 的 `*InTx` 方法（範例：`cmd/v1-p2p-exchange-service/internal/service/order/expired_consumer.go`、`internal/service/crypto_deposit/scanner_runner.go`、`internal/service/fiat_withdraw/service.go`）
- 任何狀態轉換的 UPDATE 一律帶 `WHERE 目前狀態 = 期望值`，用 `RowsAffected()` 判斷是否真的轉換成功，0 筆視為冪等（已被處理過），不得重複執行資金異動
- 每完成一個 slice，執行 `/commit`，只 stage 該 slice 實際修改的檔案，commit message 需說明搬了什麼、修了什麼 P0

實作前先讀過 `cmd/v1-p2p-exchange-service/internal/repository/wallet/repository.go` 現況（Slice 0/B/F 已新增：`FindByUserID`、`FindOne`、`ListLedgers`、`Deposit`、`DepositWithLedgerType`、`DepositWithLedgerTypeInTx`、`FreezeInTx`、`DeductFrozenBalance`、`UnfreezeBalance`、`SumCurrencyBalance`，以及既有 `Freeze`、`UnfreezeInTx`、`TransferInTx`、`AcquireLocks`）。**這幾個 slice 都需要再新增方法，但不能修改上述任何既有方法的簽章與行為。**

---

## Slice C — Crypto 提領 API + Tron Withdraw Job

### 路由

- `POST /app/wallets/crypto/withdraw`（App JWT）
- `GET /app/wallets/crypto/withdrawals`（App JWT）

### 已確認的缺陷（比 P0-3 更嚴重，務必修正）

legacy `internal/job/tron_withdraw_job.go` 有兩層問題：

**P0-3（已在 Plan 中列出）**：`internal/model/crypto_withdrawal.go` 的 `UpdateBroadcasting`、`UpdateConfirmed` 都沒有 `WHERE status=期望值` 守衛：
```go
// UpdateBroadcasting 缺守衛
`UPDATE crypto_withdrawals SET status='broadcasting', tx_hash=$2, broadcast_at=$3, updated_at=NOW() WHERE id=$1`
// UpdateConfirmed 缺守衛
`UPDATE crypto_withdrawals SET status='confirmed', confirmed_at=$2, updated_at=NOW() WHERE id=$1`
```
`confirmBroadcasting` 確認完成後才呼叫 `Wallet.DeductFrozenBalance(...)`（交易外，另一個獨立 DB 往返）——跟 Slice B 修過的 P0-4 是同一種 bug 模式：重複觸發確認流程會重複扣款。

**P0-5（新發現，未列在原始 Plan，比 P0-3 更嚴重）**：`broadcastPending` → `processSingleWithdrawal` 在**呼叫鏈上廣播之前**，唯一的併發保護是 Redis 鎖 `lock:crypto_withdraw:{id}`（60 秒 TTL，`AcquireLock` 沒有 ownership token，Redis 連線失敗時 fail-open 直接放行——這個鎖實作的缺陷已在 Slice 0 之前的安全覆核中確認過）。`ListPending` 抓出的記錄在真正廣播（`TriggerTRC20Transfer` + `BroadcastTransaction`，呼叫外部鏈上 RPC，**不可逆**）之前，**完全沒有 DB 層的原子「認領」動作**。若 Redis 鎖失效（TTL 提前過期、Redis 短暫不可用、legacy/v1 併行運作各自跑一份 job），同一筆 pending 提領可能被兩個執行緒同時廣播成功，等於平台熱錢包對外**真實多付一次錢**，且鏈上交易不可回滾。

**Slice C 必須修正**：在觸發鏈上廣播之前，先用 DB 條件式 UPDATE 原子「認領」該筆記錄（例如 `UPDATE crypto_withdrawals SET status='broadcasting', updated_at=NOW() WHERE id=$1 AND status='pending'`，`RowsAffected=0` 就跳過不廣播），Redis 鎖可以保留作為額外的效能優化（減少多副本間的無謂重複嘗試），但**不能是唯一防線**。廣播成功後再用 `tx_hash`/`broadcast_at` 補一次 UPDATE（這次不需要再帶 status 守衛，因為認領動作已經把 status 從 pending 轉走，正常路徑只有認領到的那個執行緒會走到這裡；但仍建議 `WHERE status='broadcasting'` 保險）。若廣播失敗（`broadcastWithdrawal` 回 error），要把已認領的記錄改回可重試或標成 `failed` 並解凍餘額——參考 legacy 現有的失敗處理（`UpdateFailed` + `Wallet.UnfreezeBalance`），一樣要注意 `UpdateFailed` 該不該帶狀態守衛（認領到 broadcasting 的記錄失敗才轉 failed，避免跟其他路徑打架）。

`confirmBroadcasting` 的確認+扣款要比照 Slice B 的 `confirmAndCredit` 模式，收進同一個 `db.TransactCtx`：`ConfirmInTx`（`WHERE status='broadcasting'` + RowsAffected）→ 0 筆就跳過 → 非 0 筆才呼叫 `DeductFrozenBalanceInTx`（見下）。

### 需要新增的 repository 方法

`walletrepo.WalletRepository` 新增（比照 Slice B 對 `DepositWithLedgerType`/`DepositWithLedgerTypeInTx` 的重構方式，抽公用 helper，不改既有 `DeductFrozenBalance` 的簽章與行為）：
```go
// DeductFrozenBalanceWithLedgerTypeInTx 與 DeductFrozenBalance 相同的守衛（frozen_balance >= amount
// 條件式 UPDATE + RowsAffected 檢查），但可指定帳本類型並在呼叫端傳入的 session 內執行。
DeductFrozenBalanceWithLedgerTypeInTx(ctx context.Context, session sqlx.Session, userID int64, currency, amount, ledgerType string) error
```
legacy 的 `DeductFrozenBalance` 原本就有 `ledgerType` 參數（`crypto_withdraw` 用在這裡、`fiat_withdraw` 用在 Slice H），v1 目前的版本沒有這個參數（Slice 0 遺留的已知缺口），這裡要補上。

新建 `cmd/v1-p2p-exchange-service/internal/repository/crypto_withdraw/repository.go`（package `cryptowithdrawrepo`），介面比照 `cryptodepositrepo` 的風格：
```go
type CryptoWithdrawRepository interface {
    CreateInTx(ctx context.Context, session sqlx.Session, w *entity.CryptoWithdrawal) (*entity.CryptoWithdrawal, error)
    ListPending(ctx context.Context, limit int) ([]*entity.CryptoWithdrawal, error)
    ListBroadcasting(ctx context.Context, limit int) ([]*entity.CryptoWithdrawal, error)
    // ClaimForBroadcastInTx 原子認領一筆 pending 記錄準備廣播，RowsAffected=0 代表已被認領。
    ClaimForBroadcastInTx(ctx context.Context, session sqlx.Session, id int64) (int64, error)
    UpdateBroadcastedInTx(ctx context.Context, session sqlx.Session, id int64, txHash string, broadcastAt time.Time) error
    UpdateFailedInTx(ctx context.Context, session sqlx.Session, id int64) error
    ConfirmInTx(ctx context.Context, session sqlx.Session, id int64, confirmedAt time.Time) (int64, error)
    ListByUserID(ctx context.Context, userID int64, limit, offset int64) ([]*entity.CryptoWithdrawal, error)
    CountByUserID(ctx context.Context, userID int64) (int64, error)
}
```

### Service

`cmd/v1-p2p-exchange-service/internal/service/crypto_withdraw/`（package `cryptowithdraw_service`），兩部分：

**查詢/建立服務**（對應 legacy `appcryptowithdrawlogic.go`，完整原始碼如下）：
```go
const minUSDTWithdraw = "10" // 最低 10 USDT

func (l *AppCryptoWithdrawLogic) Withdraw(userID int64, req *types.CryptoWithdrawRequest) (*types.CryptoWithdrawResponse, error) {
	if req.ToAddress == "" { return nil, apperrors.New(400, "toAddress 為必填") }
	if req.Amount == "" { return nil, apperrors.New(400, "amount 為必填") }
	if _, err := tron.TronBase58ToBytes(req.ToAddress); err != nil { return nil, apperrors.New(400, "無效的 Tron 地址") }
	amount, _, err := new(big.Float).Parse(req.Amount, 10)
	if err != nil || amount.Sign() <= 0 { return nil, apperrors.New(400, "無效的提領金額") }
	minAmount, _, _ := new(big.Float).Parse(minUSDTWithdraw, 10)
	if amount.Cmp(minAmount) < 0 { return nil, fmt.Errorf("最低提領金額為 %s USDT", minUSDTWithdraw) }

	// v1 版本：Freeze 與建立提領記錄要收進同一個 db.TransactCtx（比照 Slice F 的 FreezeInTx + CreateInTx 模式），
	// 不要像 legacy 那樣 Freeze 成功、Create 失敗才「best-effort」交易外 unfreeze——那個路徑本身沒有 P0
	// 等級的資金風險（unfreeze 失敗頂多是使用者資金卡住，不會憑空產生餘額），但既然 Slice F 已經建立了
	// 「凍結+建單同一交易」的慣例，這裡應該一致採用，不必沿用 legacy 較弱的作法。
	amountF, _ := amount.Float64()
	_ = amountF // legacy 用 float64 呼叫 Freeze；v1 请改用 FreezeInTx(ctx, session, userID, "USDT", req.Amount)（字串版本，避免精度轉換）

	w := &model.CryptoWithdrawal{UserID: userID, Currency: "USDT", Amount: req.Amount, ToAddress: req.ToAddress}
	// db.TransactCtx(ctx, func(ctx, session) error {
	//     if err := walletRepo.FreezeInTx(ctx, session, userID, "USDT", req.Amount); err != nil { return err }
	//     created, err := repo.CreateInTx(ctx, session, w); ...
	// })
	return &types.CryptoWithdrawResponse{ID: w.ID, Status: "pending"}, nil
}

func (l *AppListCryptoWithdrawalsLogic) List(userID int64, req *types.AppListCryptoWithdrawalsRequest) (*types.AppListCryptoWithdrawalsResponse, error) {
	limit := req.Limit
	if limit <= 0 || limit > 100 { limit = 20 }
	rows, err := l.svcCtx.CryptoWithdraw.ListByUserID(l.ctx, userID, limit, req.Offset)
	total, err := l.svcCtx.CryptoWithdraw.CountByUserID(l.ctx, userID)
	// map to CryptoWithdrawalItem{ID, Currency, Amount, ToAddress, TxHash, Status, ConfirmedAt, CreatedAt}
	return &types.AppListCryptoWithdrawalsResponse{List: items, Total: total}, nil
}
```

**Tron Withdraw Runner**（`pkg/schedule.Runner`，比照 `TronScannerRunner` 的寫法，掛進 `schedule_runner` group）。完整移植 `internal/job/tron_withdraw_job.go`（已附於下方「附錄」），關鍵修正點已在上面「已確認的缺陷」段落說明：`ClaimForBroadcastInTx` 認領 → 呼叫鏈上 API（`tron.Client.TriggerTRC20Transfer` → `tron.SignRawDataHex`（用 `cfg.Tron.HotWalletPrivateKey`）→ `tron.Client.BroadcastTransaction`）→ 成功則 `UpdateBroadcastedInTx`，失敗則 `UpdateFailedInTx` + `UnfreezeBalance`（這步在交易外沒關係，因為此時 status 已經是 failed，不會被其他流程碰）。`confirmBroadcasting` 比照 Slice B 的 `confirmAndCredit` 模式，`ConfirmInTx` + `DeductFrozenBalanceWithLedgerTypeInTx(ledgerType="crypto_withdraw")` 同一交易。`extractTxID` 那段手刻字串解析可以照抄（沒有安全疑慮，只是從 JSON 字串挖 `txID` 欄位），或改用標準 `encoding/json` 更穩健，兩者皆可。

### Interfaces

```go
type CryptoWithdrawRequest struct {
    ToAddress string `json:"toAddress"`
    Amount    string `json:"amount"`
}
type CryptoWithdrawResponse struct {
    ID     int64  `json:"id"`
    Status string `json:"status"`
}
type AppListCryptoWithdrawalsRequest struct {
    Limit  int64 `form:"limit,optional,default=20"`
    Offset int64 `form:"offset,optional,default=0"`
}
type CryptoWithdrawalItem struct {
    ID          int64   `json:"id"`
    Currency    string  `json:"currency"`
    Amount      string  `json:"amount"`
    ToAddress   string  `json:"toAddress"`
    TxHash      *string `json:"txHash"`
    Status      string  `json:"status"`
    ConfirmedAt *string `json:"confirmedAt"`
    CreatedAt   string  `json:"createdAt"`
}
type AppListCryptoWithdrawalsResponse struct {
    List  []CryptoWithdrawalItem `json:"list"`
    Total int64                  `json:"total"`
}
```

### 驗收條件

- `go build ./...`、`go vet`、`gofmt -l` 乾淨
- 併發測試證明：`ClaimForBroadcastInTx` 對同一筆記錄併發呼叫兩次，只有一次 `RowsAffected=1`，另一次為 0（不會廣播兩次）
- 併發測試證明：`ConfirmInTx` 重複觸發不會重複扣款（比照 Slice B `scanner_runner_test.go` 的測試風格）
- **完成條件（非程式碼）**：Plan 的併行切換規則要求——這個 slice 要在正式環境啟用前，必須先在 Tron Nile Testnet 完成上述併發測試的實測驗證（不只是 mock 測試），因為鏈上交易不可逆，風險最高
- 不新增 DB migration；不修改既有 `Freeze`/`UnfreezeInTx`/`TransferInTx`/`DepositWithLedgerType(InTx)`/`FreezeInTx`/`DeductFrozenBalance`/`UnfreezeBalance` 的簽章與行為

### Rollback

關閉 `TronWithdrawRunner` 對應的 schedule 註冊（`service/module.go` 移除該 `fx.Provide` annotate 區塊）+ revert commit + 移除路由。**若已發生重複廣播上鏈，鏈上交易不可逆，無法程式回滾，必須人工介入**（這也是為何測試網驗證是完成條件之一）。

---

## Slice E — ECPay Webhook

### 路由

- `POST /webhook/ecpay/notify`（**public，無 JWT**，唯一沒有身分驗證保護的端點，靠 CheckMacValue 簽章驗證身分）

### 已確認的缺陷（P0-2，必須修正）

legacy `internal/logic/webhookecpaylogic.go`：`FindByMerchantTradeNo`（無鎖讀取）→ 檢查 `status != "pending"`（check-then-act，非原子）→ `UpdatePaid`（無 `WHERE status='pending'` 守衛）→ **交易外**呼叫 `Wallet.DepositWithLedgerType`。ECPay 在未收到 `1|OK` 前會重送通知，兩個幾乎同時到達的通知都可能讀到 `pending` 而重複入帳；反過來，`UpdatePaid` 成功但入帳失敗，會產生「狀態顯示已付款、但錢從未入帳」且無法重試的永久漏帳（因為狀態已非 pending，冪等檢查會讓後續重試直接短路回 `1|OK`）。

**v1 修正**：`UpdatePaid` 與入帳收進同一個 `db.TransactCtx`：`UPDATE fiat_deposits SET status='paid', ... WHERE id=$1 AND status='pending'` + `RowsAffected` 檢查，0 筆直接視為已處理（回 `1|OK`，不重複入帳）；非 0 筆才在同一 session 呼叫 `walletRepo.DepositWithLedgerTypeInTx(ctx, session, deposit.UserID, deposit.Currency, deposit.Amount, "fiat_deposit", "")`（這個方法 Slice B 已經加好，直接用）。

### 簽章驗證邏輯——逐位元組保留清單（已通過安全覆核，不能改）

`pkg/ecpay/ecpay.go`（**共用套件，v1 直接 import，不要複製或重寫**，`CheckMacValue`/`VerifyCheckMacValue`/`AioCheckOutParams` 三個函式 Slice D 已在用）：

1. 排除 `CheckMacValue` 自身再排序（`sort.Strings`，位元組序）
2. `HashKey=` 前綴 + `&HashIV=` 後綴的確切位置（secret-prefix + secret-suffix 構造）
3. 對**整串**（含 `&` 與 `=`）做 `url.QueryEscape`，不是逐值 escape 再組串
4. `ToLower` 在 SHA256 **之前**，`ToUpper` 在 SHA256 **之後**
5. 缺 `CheckMacValue` key 直接 `return false`（不能用 map 零值空字串續算）
6. handler 端把 **所有** form 欄位放進 map（不要改成白名單擷取已知欄位，ECPay 未來新增欄位會導致驗證全面失敗）
7. **入帳金額必須取自 DB 的 `deposit.Amount`，不能用通知帶的 `TradeAmt`**——這是唯一擋住金額竄改的機制，絕對不可「順手優化」成用通知金額

**最容易搬丟、風險最高的一點**：`ECPayConf.IsEnabled()`（`MerchantID`/`HashKey`/`HashIV` 皆非空才 true）這個 gate 一定要在 webhook 入口保留（`conf.IsEnabled()` 否則回 `apperrors.New(503, "ECPay not configured")`）。**如果這個 gate 沒搬過去，且環境變數 `ECPAY_HASH_KEY`/`ECPAY_HASH_IV` 沒設，會變成用空字串當金鑰，任何人都能自算合法簽章，簽章驗證形同虛設**。Slice D 已經確認 v1 目前 `ECPay.IsEnabled()` 會是 false（除非環境變數已注入），這裡務必用同一個 `cfg.ECPay.IsEnabled()`。

### 需要新增的 repository 方法

`fiat_deposit` repository（Slice D 建的 `cmd/v1-p2p-exchange-service/internal/repository/fiat_deposit/repository.go`，目前只有 `Create`/`ListByUserID`/`CountByUserID`）新增：
```go
FindByMerchantTradeNo(ctx context.Context, tradeNo string) (*entity.FiatDeposit, error) // 查無回 sqlx.ErrNotFound
// ConfirmPaidInTx 原子把 pending → paid 並記錄 ECPay 回傳資訊，RowsAffected=0 代表已處理過。
ConfirmPaidInTx(ctx context.Context, session sqlx.Session, id int64, ecpayOrderNo, paymentType string, paidAt time.Time) (int64, error)
UpdateFailed(ctx context.Context, id int64) error // 非交易，付款失敗時單獨呼叫即可（不涉及資金異動）
```

### Service / Handler

新建 `cmd/v1-p2p-exchange-service/internal/service/ecpay_webhook/`（或直接放進既有 `fiat_deposit_service`，兩種都合理，選一種跟現有慣例最一致的）：
```go
func (s *...) HandleNotify(ctx context.Context, params map[string]string) (ok bool, message string) {
    conf := s.cfg.ECPay
    if !conf.IsEnabled() { return false, "ECPay not configured" }
    if !ecpay.VerifyCheckMacValue(params, conf.HashKey, conf.HashIV) { return false, "signature mismatch" }

    tradeNo := params["MerchantTradeNo"]
    deposit, err := s.repo.FindByMerchantTradeNo(ctx, tradeNo)
    if err != nil {
        if errors.Is(err, sqlx.ErrNotFound) { return false, "deposit not found" }
        return false, "internal error"
    }

    if params["RtnCode"] != "1" {
        if deposit.Status == statusPending { _ = s.repo.UpdateFailed(ctx, deposit.ID) }
        return true, "OK" // 付款失敗仍回 1|OK，避免 ECPay 一直重送
    }

    err = s.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
        affected, err := s.repo.ConfirmPaidInTx(ctx, session, deposit.ID, params["TradeNo"], params["PaymentType"], time.Now())
        if err != nil { return err }
        if affected == 0 { return nil } // 已處理過，冪等
        return s.walletRepo.DepositWithLedgerTypeInTx(ctx, session, deposit.UserID, deposit.Currency, deposit.Amount, "fiat_deposit", "")
    })
    if err != nil { return false, "internal error" }
    return true, "OK"
}
```
Handler（新建 `cmd/v1-p2p-exchange-service/internal/server/handlers/ecpay_webhook.go`）比照 legacy `webhookecpayhandler.go`：`r.ParseForm()` → 所有 form 欄位塞進 `map[string]string`（`vs[0]`）→ 呼叫 service → **回應格式是純文字，不是 JSON**：成功 `w.Header().Set("Content-Type","text/plain")` + `w.Write([]byte("1|OK"))`（HTTP 200）；失敗 `w.Write([]byte("0|"+message))`（**同樣 HTTP 200**，legacy 用 `http.StatusOK` 包裝失敗訊息，不是 4xx/5xx——ECPay 的通知協定期待這個格式，不要改成標準 JSON error response）。

### 路由

`POST /webhook/ecpay/notify` 掛在 **public routes**（不經過 App/Backend JWT 群組，比照 `server.go` 目前 `/api/v1/version`、`/app/auth/login`、`/backend/auth/login` 那種直接 `server.AddRoute` 的寫法，不要包進 `rest.WithJwt(...)`）

### 驗收條件

- golden test：固定參數 + 固定 HashKey/HashIV，斷言 `ecpay.CheckMacValue` 輸出與已知正確值逐位元組相同（`pkg/ecpay/ecpay_test.go` 已有 `TestCheckMacValue_OfficialVector`，可以直接引用同一組向量或另開一個 v1 端的 handler 層 golden test）
- 併發測試：兩個 goroutine 同時對同一筆 `MerchantTradeNo` 呼叫 `HandleNotify`（`RtnCode=1`），斷言 `wallet_ledgers` 只多一筆、餘額只增加一次
- 測試 `conf.IsEnabled()=false`（HashKey 空字串）時回 503/`false`，不會被當成驗證通過
- 測試簽章錯誤回 `0|signature mismatch`
- 不新增 DB migration；不修改既有 wallet repository 方法簽章與行為

### 併行與切換（部署面，非程式碼）

此 slice 上線時，ECPay 商店後台的 Notify URL 設定需切換指向 v1 端點，且此為完成條件之一——不能兩邊同時接收通知，否則 v1 做的原子性修正會被繞過（legacy 端仍是 TOCTOU 重複入帳）。

### Rollback

移除 `/webhook/ecpay/notify` 路由註冊前，**必須先確認 ECPay 後台設定的 Notify URL 沒有指向這個端點**，否則會造成付款成功但永遠收不到通知。已入帳紀錄不回滾。

---

## Slice H — Backend 法幣提領審核

### 路由

- `GET /backend/fiat-withdrawals`（Backend JWT）
- `PUT /backend/fiat-withdrawals/:id/review`（Backend JWT）

### 已確認的缺陷（P0-1，必須修正，這是整個安全覆核裡最先發現的）

legacy `internal/logic/backendreviewfiatwithdrawlogic.go`：讀狀態（`FindByID`，無鎖）→ `if w.Status != "pending"` check-then-act → 資金異動（`DeductFrozenBalance` 或 `UnfreezeBalance`）→ **另一個獨立交易**才更新 `fiat_withdrawals` 狀態（`UpdateApproved`/`UpdateRejected`，同樣沒有 `WHERE status='pending'` 守衛）。並發雙擊、前端重送、或「資金異動成功但狀態更新失敗」都會導致同一筆提領被處理兩次——approve 兩次等於後台憑空多扣一次（使用者少收一次卻被系統認為已撥款兩次，對平台是損失），reject 兩次等於使用者憑空拿回 2 倍金額。

**v1 修正**：狀態轉換與資金異動收進同一個 `db.TransactCtx`：
```go
db.TransactCtx(ctx, func(ctx, session) error {
    affected, err := repo.ClaimForReviewInTx(ctx, session, id) // UPDATE ... SET status='reviewing'（或直接做下面兩選一）WHERE status='pending'
    if affected == 0 { return apierrors.New(409, "此申請已審核完畢") }
    if action == "approve" {
        if err := walletRepo.DeductFrozenBalanceWithLedgerTypeInTx(ctx, session, w.UserID, w.Currency, w.Amount, "fiat_withdraw"); err != nil { return err }
        return repo.UpdateApprovedInTx(ctx, session, id, reviewerID, now)
    }
    if err := walletRepo.UnfreezeInTx(ctx, session, w.UserID, w.Currency, amountFloat); err != nil { return err } // 注意：既有 UnfreezeInTx 簽章是 float64，見下方說明
    return repo.UpdateRejectedInTx(ctx, session, id, reviewerID, now, reason)
})
```
更簡單的寫法：不用額外的「reviewing」中繼狀態，直接讓 `UpdateApprovedInTx`/`UpdateRejectedInTx` 自己帶 `WHERE status='pending'` 守衛 + 回傳 `RowsAffected`，在同一個 TransactCtx 內先做資金異動（此時還沒鎖狀態，理論上有極小窗口）**或**先做狀態轉換（`RowsAffected=0` 直接回衝突）再做資金異動——**後者更安全**，因為狀態轉換失敗（0 筆）就不會執行資金異動，兩個並發審核只有一個能通過 `WHERE status='pending'` 的第一道 UPDATE。建議順序：**先狀態轉換、確認 RowsAffected=1、再做資金異動**，資金異動失敗則整個 transaction 回滾（狀態轉換也一起回滾，不會卡在「狀態已改但錢沒動」的中間態）。

`DeductFrozenBalance`/`UnfreezeBalance` 本身（Slice 0）已經有 `frozen_balance >= amount` 的條件式 UPDATE + RowsAffected 檢查（防止扣成負數），這裡是**額外**要修正「審核申請本身被重複處理」這一層，兩層防護不衝突、都需要。

### 既有方法簽章注意

- `WalletRepository.UnfreezeInTx` 現有簽章是 `(ctx, session, userID, currency string, amount float64)`（float64！這是既有給 order escrow 用的舊簽章，Slice F 沒有動它）。這裡如果要傳字串金額（legacy `FiatWithdrawal.Amount` 是字串），要嘛用 `strconv.ParseFloat` 轉換後呼叫既有 `UnfreezeInTx`，要嘛（更一致的做法）新增一個字串版本的 `UnfreezeInTxStr`／複用 Slice 0 已有的 `UnfreezeBalance` 內部邏輯抽出一個接受 session 的 helper。**選哪種做法都可以，但不要修改 `UnfreezeInTx` 既有簽章**（order 那邊在用），新增方法即可。
- approve 需要的 `DeductFrozenBalanceWithLedgerTypeInTx`（帶 `ledgerType="fiat_withdraw"`）跟 Slice C 需要的是**同一個方法**（Slice C 用 `ledgerType="crypto_withdraw"`）。**如果先做 Slice C，這個方法在 Slice C 就會加好，Slice H 直接複用；如果先做 Slice H，就在這裡加，Slice C 複用。哪個 slice 先做就在哪個 slice 加，不要重複定義。**

### 需要新增的 repository 方法

`fiat_withdraw` repository（Slice F 建的 `cmd/v1-p2p-exchange-service/internal/repository/fiat_withdraw/repository.go`，目前只有 `CreateInTx`/`ListByUserID`/`CountByUserID`）新增：
```go
FindByID(ctx context.Context, id int64) (*entity.FiatWithdrawal, error) // 查無回 sqlx.ErrNotFound
ListByStatus(ctx context.Context, status string, limit, offset int64) ([]*entity.FiatWithdrawal, error) // status="" 或 "all" 回全部
CountByStatus(ctx context.Context, status string) (int64, error)
// UpdateApprovedInTx / UpdateRejectedInTx：WHERE status='pending' 守衛 + RowsAffected。
UpdateApprovedInTx(ctx context.Context, session sqlx.Session, id, reviewerID int64, reviewedAt time.Time) (int64, error)
UpdateRejectedInTx(ctx context.Context, session sqlx.Session, id, reviewerID int64, reviewedAt time.Time, reason string) (int64, error)
```

### Service

擴充既有 `backendAdminService`（`cmd/v1-p2p-exchange-service/internal/service/backend_admin/`，Slice G/I 已經在擴充這個 service，比照它們的模式，這次再加建構子參數 `fiatWithdrawRepo`）：
```go
ListFiatWithdrawals(ctx context.Context, status string, limit, offset int64) (*backend_interface.BackendListFiatWithdrawalsResponse, error)
ReviewFiatWithdrawal(ctx context.Context, reviewerID, id int64, action, reason string) error
```
`ReviewFiatWithdrawal` 驗證邏輯（比照 legacy）：`action` 必須是 `"approve"` 或 `"reject"`；`reject` 時 `reason` 不可空；`FindByID` 查無回 404；交易組法見上方「已確認的缺陷」段落。`reviewerID` **一律從 handler 的 `ctxUID(r)`（Backend JWT context）取得，絕對不接受 request body 傳入**——這是 Slice G 已經建立的安全底線，這裡延續同一個規則。

### Interfaces

```go
type BackendListFiatWithdrawalsRequest struct {
    Status string `form:"status,optional,default=pending"`
    Limit  int64  `form:"limit,optional,default=20"`
    Offset int64  `form:"offset,optional,default=0"`
}
type FiatWithdrawalItem struct {
    ID           int64   `json:"id"`
    UserID       int64   `json:"userId"`
    Currency     string  `json:"currency"`
    Amount       string  `json:"amount"`
    BankCode     string  `json:"bankCode"`
    BankAccount  string  `json:"bankAccount"` // 後台可見完整帳號（不遮蔽，跟 App 端列表不同）
    AccountName  string  `json:"accountName"`
    Status       string  `json:"status"`
    ReviewedBy   *int64  `json:"reviewedBy"`
    RejectReason *string `json:"rejectReason"`
    CreatedAt    string  `json:"createdAt"`
}
type BackendListFiatWithdrawalsResponse struct {
    List  []FiatWithdrawalItem `json:"list"`
    Total int64                `json:"total"`
}
type BackendReviewFiatWithdrawalRequest struct {
    ID     int64  `path:"id"`
    Action string `json:"action"` // approve | reject
    Reason string `json:"reason,optional"`
}
```

### Handler

擴充既有 `BackendHandler`（`server/handlers/backend.go`），加 `ListFiatWithdrawals`、`ReviewFiatWithdrawal` 兩個 method，比照 Slice G 的 `adminUID := ctxUID(r)` 寫法。

### 驗收條件

- 併發測試：兩個 goroutine 同時對同一筆 `pending` 提領呼叫 `ReviewFiatWithdrawal`（一個 approve 一個 reject，或兩個都 approve），斷言只有一次成功、資金只異動一次、`wallet_ledgers` 只多一筆
- 已審核過的申請（status 非 pending）再次呼叫回明確的衝突錯誤（建議 409），不是靜默成功
- `reviewerID` 來源測試：確認來自 JWT context，request body 帶任何 `reviewerID`/`adminUID` 欄位都不會被採用（介面根本不定義這個欄位）
- 不新增 DB migration；不修改既有 `Freeze`/`FreezeInTx`/`UnfreezeInTx`/`DepositWithLedgerType(InTx)`/`DeductFrozenBalance`/`UnfreezeBalance` 的簽章與行為

### Rollback

revert commit + 移除路由；已審核完成的提領不回滾，除非稽核發現雙重撥款才需人工對帳。

---

## 附錄：legacy 完整原始碼（供實作時逐行比對）

### tron_withdraw_job.go

```go
package job

const withdrawLockPrefix = "lock:crypto_withdraw:"

func StartTronWithdrawJob(ctx context.Context, conf config.TronConf, deps TronWithdrawDeps) {
	if !conf.IsEnabled() { logx.Info("[tron-withdraw] disabled: hot wallet not configured"); return }
	client := tron.NewClient(conf.TronGridURL, conf.TronGridAPIKey)
	interval := time.Duration(conf.WithdrawIntervalSeconds) * time.Second
	go func() {
		for {
			select {
			case <-ctx.Done(): return
			case <-time.After(interval):
				broadcastPending(ctx, conf, client, deps)
				confirmBroadcasting(ctx, conf, client, deps)
			}
		}
	}()
}

func broadcastPending(ctx context.Context, conf config.TronConf, client *tron.Client, deps TronWithdrawDeps) {
	pending, err := deps.CryptoWithdraw.ListPending(ctx) // WHERE status='pending' ORDER BY created_at ASC LIMIT 50
	if err != nil { logx.Errorf("[tron-withdraw] ListPending: %v", err); return }
	for _, w := range pending { processSingleWithdrawal(ctx, conf, client, w, deps) }
}

func processSingleWithdrawal(ctx context.Context, conf config.TronConf, client *tron.Client, w *model.CryptoWithdrawal, deps TronWithdrawDeps) {
	lockKey := fmt.Sprintf("lock:crypto_withdraw:%d", w.ID)
	if deps.RDB != nil {
		unlock, err := deps.RDB.AcquireLock(ctx, lockKey, 60*time.Second)
		if err != nil { return } // ⚠️ 唯一的併發保護，見上方 P0-5 說明——v1 必須改成 DB 條件式 UPDATE 認領
		defer unlock()
	}
	if err := broadcastWithdrawal(ctx, conf, client, w, deps); err != nil {
		logx.Errorf("[tron-withdraw] broadcast id=%d: %v", w.ID, err)
		_ = deps.CryptoWithdraw.UpdateFailed(ctx, w.ID)
		_ = deps.Wallet.UnfreezeBalance(ctx, w.UserID, w.Currency, w.Amount)
	}
}

func broadcastWithdrawal(ctx context.Context, conf config.TronConf, client *tron.Client, w *model.CryptoWithdrawal, deps TronWithdrawDeps) error {
	sunAmount, err := tron.USDTToSun(w.Amount)
	if err != nil { return err }
	trigger, rawDataHex, err := client.TriggerTRC20Transfer(ctx, conf.HotWalletAddress, w.ToAddress, conf.USDTContractAddress, sunAmount)
	if err != nil { return err }
	signature, err := tron.SignRawDataHex(rawDataHex, conf.HotWalletPrivateKey)
	if err != nil { return err }
	if err := client.BroadcastTransaction(ctx, trigger.Transaction, signature); err != nil { return err }
	txHash, err := extractTxID(trigger.Transaction)
	if err != nil { logx.Errorf("[tron-withdraw] extractTxID: %v", err); txHash = "unknown" }
	now := time.Now()
	if err := deps.CryptoWithdraw.UpdateBroadcasting(ctx, w.ID, txHash, now); err != nil { return err } // ⚠️ 缺 WHERE status 守衛
	logx.Infof("[tron-withdraw] broadcast id=%d txHash=%s", w.ID, txHash)
	return nil
}

func confirmBroadcasting(ctx context.Context, conf config.TronConf, client *tron.Client, deps TronWithdrawDeps) {
	broadcasting, err := deps.CryptoWithdraw.ListBroadcasting(ctx) // WHERE status='broadcasting' LIMIT 50
	if err != nil || len(broadcasting) == 0 { return }
	currentBlock, err := client.GetCurrentBlockNumber(ctx)
	if err != nil { return }
	for _, w := range broadcasting {
		if w.TxHash == nil { continue }
		blockNumber, _, err := client.GetTransactionDetail(ctx, *w.TxHash)
		if err != nil || blockNumber == 0 { continue }
		if (currentBlock - blockNumber) < int64(conf.ConfirmationBlocks) { continue }
		now := time.Now()
		if err := deps.CryptoWithdraw.UpdateConfirmed(ctx, w.ID, now); err != nil { continue } // ⚠️ 缺 WHERE status 守衛，P0-3
		if err := deps.Wallet.DeductFrozenBalance(ctx, w.UserID, w.Currency, w.Amount, "crypto_withdraw"); err != nil { // ⚠️ 交易外，P0-3
			logx.Errorf("[tron-withdraw] DeductFrozenBalance id=%d: %v", w.ID, err)
		} else {
			logx.Infof("[tron-withdraw] confirmed id=%d user=%d amount=%s", w.ID, w.UserID, w.Amount)
		}
	}
}

// extractTxID／indexOf：手刻字串解析 raw tx JSON 裡的 "txID":"<value>"，無安全疑慮，可照抄或改用 encoding/json。
```

### webhookecpaylogic.go + webhookecpayhandler.go

```go
func (l *WebhookECPayLogic) HandleNotify(params map[string]string) error {
	conf := l.svcCtx.Config.ECPay
	if !conf.IsEnabled() { return apperrors.New(503, "ECPay not configured") }
	if !ecpay.VerifyCheckMacValue(params, conf.HashKey, conf.HashIV) {
		l.Errorf("[ecpay-webhook] CheckMacValue mismatch: %v", params)
		return fmt.Errorf("signature mismatch")
	}
	tradeNo := params["MerchantTradeNo"]
	rtnCode := params["RtnCode"]
	ecpayOrderNo := params["TradeNo"]
	paymentType := params["PaymentType"]

	deposit, err := l.svcCtx.FiatDeposit.FindByMerchantTradeNo(l.ctx, tradeNo)
	if err != nil {
		if err == sqlx.ErrNotFound { return fmt.Errorf("deposit not found: %s", tradeNo) }
		return err
	}
	if deposit.Status != "pending" { return nil } // ⚠️ check-then-act，非原子，P0-2
	if rtnCode != "1" {
		if err := l.svcCtx.FiatDeposit.UpdateFailed(l.ctx, deposit.ID); err != nil { l.Errorf(...) }
		return nil
	}
	now := time.Now()
	if err := l.svcCtx.FiatDeposit.UpdatePaid(l.ctx, deposit.ID, ecpayOrderNo, paymentType, now); err != nil { return err } // ⚠️ 缺 WHERE status 守衛
	if err := l.svcCtx.Wallet.DepositWithLedgerType(l.ctx, deposit.UserID, deposit.Currency, deposit.Amount, "fiat_deposit"); err != nil { // ⚠️ 交易外
		l.Errorf("[ecpay-webhook] DepositWithLedgerType user=%d amount=%s: %v", deposit.UserID, deposit.Amount, err)
		return err
	}
	return nil
}

// handler：r.ParseForm() → 所有 form 值塞進 map[string]string → HandleNotify(params)
// 成功："1|OK"（text/plain, HTTP 200）；失敗："0|"+err.Error()（同樣 HTTP 200，不是 4xx/5xx）
```

`pkg/ecpay/ecpay.go` 全文已在 Slice D 用過，此處不重複貼；驗證邏輯見本文件 Slice E 段落的「逐位元組保留清單」。

### backendlistfiatwithdrawlogic.go + backendreviewfiatwithdrawlogic.go

```go
func (l *BackendListFiatWithdrawLogic) List(req *types.BackendListFiatWithdrawalsRequest) (*types.BackendListFiatWithdrawalsResponse, error) {
	rows, err := l.svcCtx.FiatWithdraw.ListByStatus(l.ctx, req.Status, req.Limit, req.Offset) // status="" 或 "all" 用 DESC 排序回全部；否則 status=$1 用 ASC 排序
	total, err := l.svcCtx.FiatWithdraw.CountByStatus(l.ctx, req.Status)
	// map to types.FiatWithdrawalItem（完整銀行帳號，不遮蔽——後台可見）
	return &types.BackendListFiatWithdrawalsResponse{List: list, Total: total}, nil
}

func (l *BackendReviewFiatWithdrawLogic) Review(reviewerID int64, req *types.BackendReviewFiatWithdrawalRequest) error {
	if req.Action != "approve" && req.Action != "reject" { return apperrors.New(400, "action 必須為 approve 或 reject") }
	if req.Action == "reject" && req.Reason == "" { return apperrors.New(400, "拒絕時必須填寫原因") }
	w, err := l.svcCtx.FiatWithdraw.FindByID(l.ctx, req.ID)
	if err != nil {
		if err == sqlx.ErrNotFound { return apperrors.ErrNotFound }
		return err
	}
	if w.Status != "pending" { return apperrors.New(400, "此申請已審核完畢") } // ⚠️ check-then-act，非原子，P0-1
	now := time.Now()
	if req.Action == "approve" {
		if err := l.svcCtx.Wallet.DeductFrozenBalance(l.ctx, w.UserID, w.Currency, w.Amount, "fiat_withdraw"); err != nil { return err } // ⚠️ 交易外
		if err := l.svcCtx.FiatWithdraw.UpdateApproved(l.ctx, w.ID, reviewerID, now); err != nil { return err } // ⚠️ 缺 WHERE status 守衛
	} else {
		if err := l.svcCtx.Wallet.UnfreezeBalance(l.ctx, w.UserID, w.Currency, w.Amount); err != nil { return err }
		if err := l.svcCtx.FiatWithdraw.UpdateRejected(l.ctx, w.ID, reviewerID, now, req.Reason); err != nil { return err }
	}
	return nil
}

// handler：l.Review(ctxUID(r), &req) —— reviewerID 一律來自 JWT context
```
