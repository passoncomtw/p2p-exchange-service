# P2P Exchange Service

## 原型設計 (Claude design)

v1 原型(使用者端手機 + 後台桌面)設計連結:

https://claude.ai/design/p/2c7b1880-aaf0-4e7c-9bd1-496d0b4560af?file=P2P+Exchange+v1.dc.html

匯出的原始檔位於 `prototypes/P2P-Exchange-v1.html`(Claude design 元件格式 `.dc.html`,需搭配設計執行階段 `support.js` 才能渲染;完整可互動版請於上方連結檢視)。詳見 `prototypes/README.md`。

## Mobile App 發版流程 (iOS / Android)

### 前置條件

在 GitHub 倉庫的 Settings → Secrets and variables → Actions 設定以下 Secrets：

| Secret | 用途 |
|---|---|
| `EXPO_TOKEN` | Expo 帳號 Access Token（[expo.dev](https://expo.dev) → Access Tokens） |
| `EXPO_APPLE_SPECIFIC_PASSWORD` | Apple ID 專用密碼（用於 TestFlight 上傳） |
| `GOOGLE_SERVICE_ACCOUNT_KEY` | Google Play 服務帳號 JSON 完整內容（CI 注入，不 commit） |

Google Play 服務帳號：`expo-ci@ecoinwallet.iam.gserviceaccount.com`，金鑰檔案名稱 `ecoinwallet-6394ffc01290.json`（已加入 `.gitignore`，由 CI 從 Secret 注入）。

### 觸發方式

Workflow 名稱：`expo-build-submit`，僅支援手動觸發：

1. GitHub → Actions → `expo-build-submit` → Run workflow
2. 輸入語義化版本號，例如 `v1.0.0`
3. 選擇平台：`ios`、`android`、或 `all`

### 流程說明

```
輸入版本號 + 平台
    ↓
版本號格式驗證（必須符合 vX.Y.Z）
    ↓
安裝依賴（yarn install）
    ↓
寫入 app.json expo.version
    ↓
EAS Build（在 Expo 伺服器上編譯，--wait 等待完成）
    ↓
iOS  → Submit to TestFlight（EXPO_APPLE_SPECIFIC_PASSWORD）
Android → 注入 Google Service Account Key → Submit to Google Play Internal
```

選擇 `all` 時，iOS 與 Android 會在 Expo 伺服器上同時編譯，完成後依序 submit。

### EAS 設定

`frontend/app/eas.json`：

- `production` build：distribution 為 `store`，iOS/Android 均使用 `credentialsSource: remote`（EAS 管理簽名憑證）
- Android submit：track 為 `internal`（Google Play Internal Testing）
- iOS submit：`ascAppId: 6477443899`，`appleTeamId: W6K7F9HNX5`

### 安全注意事項

- Google Service Account Key **不可 commit**（`.gitignore` 已封鎖 `ecoinwallet-*.json`）
- 定期到 GCP Console → IAM → 服務帳號 → 金鑰，輪換金鑰並更新 GitHub Secret

---

## P2P 訂單流程

P2P 交易的核心是「掛單 → 吃單 → 線下付款 → 確認 → 放幣」，依掛單類型分為兩種路徑。

### 角色定義

| 角色 | 平台 | 說明 |
|------|------|------|
| 買家 (Buyer) | App | 想用法幣（TWD）購買加密貨幣（USDT） |
| 賣家 (Seller) | App | 持有加密貨幣（USDT），想換成法幣（TWD） |
| 後台管理員 (Admin) | Web 後台 | 僅在訂單爭議時介入仲裁 |
| 系統 | 後端 API | 媒合驗證、法幣計算、托管（Escrow）管理、狀態流轉 |

### 買幣掛單流程 (type="buy")

買家 A 發布買幣掛單，賣家 B 來吃單。

```
買家 A (App)              系統                        賣家 B (App)
    |                       |                             |
    |-- 建立 buy listing -->|                             |
    |   (不需綁銀行帳號)     |-- listing active ---------->|
    |                       |                             |
    |                       |              瀏覽市場掛單 --|
    |                       |                             |
    |                       |<-- 吃單 POST /app/orders ---|
    |                       |    (輸入要賣的 USDT 數量)     |
    |                       |                             |
    |                       |--- 媒合驗證 ----------------|
    |                       |   1. listing active          |
    |                       |   2. remaining >= 數量       |
    |                       |   3. 取 B 的銀行帳號         |
    |                       |   4. 法幣計算                |
    |                       |   5. escrow lock B 的 USDT   |
    |                       |                             |
    |<-- matched 通知 ------|------ matched 通知 -------->|
    |                       |                             |
    |-- 線下匯 TWD -------->|                  (至 B 帳號) |
    |-- 標記已付款 -------->|                             |
    |                       |-------- paid -------------->|
    |                       |                             |
    |                       |<------ 確認收款到帳 ---------|
    |                       |                             |
    |<-- completed ---------|--- escrow release USDT ---->|
    |   (收到 USDT)          |      (釋放至 A 帳戶)         |
```

### 賣幣掛單流程 (type="sell")

賣家 A 發布賣幣掛單，買家 B 來吃單。

```
賣家 A (App)              系統                        買家 B (App)
    |                       |                             |
    |-- 建立 sell listing ->|                             |
    |   (必須綁銀行帳號)     |-- listing active ---------->|
    |                       |                             |
    |                       |              瀏覽市場掛單 --|
    |                       |                             |
    |                       |<-- 吃單 POST /app/orders ---|
    |                       |    (輸入要買的 USDT 數量)     |
    |                       |                             |
    |                       |--- 媒合驗證 ----------------|
    |                       |   1. listing active          |
    |                       |   2. remaining >= 數量       |
    |                       |   3. 沿用 A 的銀行帳號       |
    |                       |   4. 法幣計算                |
    |                       |   5. escrow lock A 的 USDT   |
    |                       |                             |
    |<-- matched 通知 ------|------ matched 通知 -------->|
    |                       |                             |
    |                       |          線下匯 TWD ------->|
    |                       |                  (至 A 帳號) |
    |                       |<-------- 標記已付款 ---------|
    |                       |                             |
    |<-------- paid --------|                             |
    |                       |                             |
    |-- 確認收款到帳 ------>|                             |
    |                       |                             |
    |                       |--- escrow release USDT ---->|
    |                       |      (釋放至 B 帳戶)         |
    |                       |------ completed ----------->|
```

### 兩種掛單的差異比較

| | 買幣掛單 (buy) | 賣幣掛單 (sell) |
|---|---|---|
| 掛單者角色 | 買家 | 賣家 |
| 吃單者角色 | 賣家 | 買家 |
| 銀行帳號來源 | 吃單的賣家提供（系統自動取第一筆） | 掛單的賣家綁定（建單時必填） |
| Escrow lock | 鎖吃單者（賣家）的 USDT | 鎖掛單者（賣家）的 USDT |
| 法幣匯款方向 | 買家 → 賣家帳號 | 買家 → 賣家帳號 |
| 確認收款者 | 賣家（吃單者） | 賣家（掛單者） |
| Escrow release | 釋放 USDT 至買家（掛單者） | 釋放 USDT 至買家（吃單者） |

### 訂單狀態流轉

```
matched ──[買方付款]──> paid ──[賣方確認]──> releasing ──> completed
    |                    |
    |                    └──[爭議]──> disputed ──[admin]──> completed (release)
    |                                                  └──> cancelled (refund)
    └──[取消]──> cancelled
```

| 狀態 | 說明 | 可操作者 |
|------|------|---------|
| matched | 訂單建立，等待買方付款 | buyer/seller 可取消 |
| paid | 買方已標記付款 | seller 確認收款 / buyer or seller 發起爭議 |
| releasing | 賣方確認，系統放行中（中間狀態） | 系統自動 |
| completed | 訂單完成，USDT 已釋放 | - |
| cancelled | 訂單取消，USDT 退還賣方 | - |
| disputed | 爭議中，等待後台仲裁 | admin resolve |

### 法幣轉換計算

```
fiatAmount       = cryptoAmount x listing.price
platformFeeAmt   = fiatAmount x platformFeeRate
paymentFeeAmt    = fiatAmount x paymentFeeRate
totalFee         = platformFeeBase + platformFeeAmt + paymentFeeBase + paymentFeeAmt
totalAmount      = fiatAmount + totalFee  （買家實際應付金額）
```

第一階段所有手續費均為 0，`totalAmount = fiatAmount`。

### 托管（Escrow）機制

| 時機 | 動作 | 說明 |
|------|------|------|
| 建立訂單 | lock | 鎖定賣方的 USDT |
| 賣方確認收款 | release | 釋放 USDT 至買方 |
| 取消訂單 | refund | 退還 USDT 至賣方 |
| Admin 仲裁 complete | release | 釋放 USDT 至買方 |
| Admin 仲裁 refund | refund | 退還 USDT 至賣方 |

目前 Escrow 為純紀錄型設計，尚未對接鏈上操作或錢包餘額系統。

---

## 預設帳號

執行 `go run backend/cmd/seed/main.go -f backend/etc/config.yaml` 建立以下帳號。

### 後台管理員

| 帳號 | 密碼 | 角色 |
|---|---|---|
| admin001 | admin@1234 | admin |

### App 使用者

| 帳號 | 密碼 |
|---|---|
| testdemo001 | a12345678 |
| testdemo002 | a12345678 |
| testdemo003 | a12345678 |
