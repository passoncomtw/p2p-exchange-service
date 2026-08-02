# P2P Exchange Service

## Spec-Driven Development (SDD)

本專案採用 OpenSpec 進行規格驅動開發。

- **規格目錄**：[`openspec/specs/`](openspec/specs/) — 各領域業務規格
- **變更追蹤**：[`openspec/changes/`](openspec/changes/) — 進行中與已封存的變更提案
- **專案概覽**：[`openspec/project.md`](openspec/project.md) — 技術棧、領域表、核心狀態機

常用指令（需安裝 `@fission-ai/openspec`）：
- `/opsx:propose "描述"` — 提出新需求並產生規格與任務
- `/opsx:apply` — 根據任務清單開始實作
- `/opsx:explore "問題"` — 探索問題與釐清需求

---

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

---

## App 開發工具（Dev Tools）

### 啟動開發伺服器

```bash
cd frontend/app
npx expo start
```

**必須使用 `npx expo start`**，直接用 Xcode / Android Studio 啟動 Metro 時，Expo CLI 鍵盤指令（`shift+m` 等）不會作用。

### Redux DevTools

套件：`redux-devtools-expo-dev-plugin`（已安裝，`configureStore.ts` 已設定）

**開啟步驟：**

1. 在模擬器或實機上開啟 Dev Menu：
   - iOS 模擬器：`cmd + d`
   - Android 模擬器：`cmd + m`（Mac）/ `ctrl + m`（Windows/Linux）
   - 實機：物理搖晃裝置
2. 選擇 **Open DevTools Plugin**
3. Chrome 另開新視窗，顯示 Redux state 樹、action 歷程與 diff

或在 `npx expo start` 的終端按 `shift+m`，選擇 `redux-devtools-expo-dev-plugin`。

> **注意**：React Native Debugger（獨立桌面 App）不支援 Expo 56 / RN 0.85 的新架構（Bridgeless），請使用上述 Expo DevTools Plugin 方式。

### React Native JS Debugger（Fusebox）

用於 JS 執行、斷點、Console、網路請求（Expo Network 分頁）偵錯，與 Redux DevTools 是**不同視窗**。

- 開啟網址：`http://localhost:8081/json`，找到對應裝置的 `devtoolsFrontendUrl`
- 或從 Dev Menu 選擇 **Open JS Debugger**

### 查看已連線裝置

```bash
curl http://localhost:8081/json
```

---

## 通知系統架構

App (React Native) 與 Web (React + Vite) 統一使用 Redux 佇列管理通知，不在 component 層直接呼叫 `Alert.alert` 或內嵌錯誤 UI 顯示 API 回應。

### App (`frontend/app`)

| 元件 | 路徑 | 說明 |
|------|------|------|
| `notificationSlice` | `src/navigation/store/slices/notificationSlice.ts` | 佇列狀態，不持久化 |
| `NotificationHandler` | `src/components/NotificationHandler.tsx` | 根元件，監聽佇列呼叫 `Alert.alert` |
| `errorSaga` | `src/navigation/store/sagas/errorSaga.ts` | 統一攔截所有 `*Failure` action，dispatch `pushNotification` |

**觸發方式：**
```ts
// saga 成功路徑
yield put(pushNotification({ type: 'success', message: '掛單建立成功' }))

// screen 直接 API call 的 catch
dispatch(pushNotification({ type: 'error', message: t('order.message.submitFailed') }))
```

### Web (`frontend/web`)

| 元件 | 路徑 | 說明 |
|------|------|------|
| `notificationSlice` | `src/slices/notificationSlice.js` | 佇列狀態，不持久化 |
| `NotificationHandler` | `src/components/NotificationHandler.jsx` | MUI Snackbar，掛載於 `App.jsx` |

**通知類型：** `success` / `error` / `warning` / `info`

**保留使用 `Alert.alert` 的情境：**
- 需要使用者確認的操作（取消掛單確認、仲裁確認等）
- 帶有導航按鈕的提示（尚未新增收款帳戶）
- 欄位格式驗證（金額格式錯誤）

---

## NATS JetStream

### 背景與技術決策

| 項目 | 決定 |
|---|---|
| MQ 服務 | Synadia Cloud Free Plan（含 JetStream） |
| Stream Retention | WorkQueue — 訊息被 consumer ack 後立即刪除 |
| 連線方式 | `.creds` 憑證檔（Synadia Cloud）或 User/Password（本地） |

### Stream 架構

所有 Stream 使用 WorkQueue Retention，每個 Stream 只能有一個 Consumer。

| Stream | Subjects | 保留 | 用途 |
|--------|----------|------|------|
| `P2P_ORDERS` | `order.*`, `order.payment.*` | 7 天 | 訂單生命週期事件 |
| `P2P_LEDGER` | `ledger.*` | 30 天 | 平台內部帳本操作 |
| `P2P_NOTIFY` | `notify.buyer.*`, `notify.seller.*`, `notify.admin.*` | 3 天 | WebSocket 推播 |

Stream 於服務啟動時自動初始化，若已存在則跳過建立（冪等）。

### Synadia Cloud 連線設定

| 項目 | 值 |
|---|---|
| Server URL | `tls://connect.ngs.global` |
| Cluster | `ngsprod-aws-apeast2`（Taipei, Taiwan），RTT 約 9–20ms |
| Account ID | AC5PRLIT |
| JetStream | 已啟用 |
| Max Connections | 10 |
| Max Msg Payload | 512 KiB |
| Network Data | 10 GiB/month |
| Storage | 5 GiB |

**設定方式：**

```yaml
# backend/etc/config.yaml
Nats:
  URL: "tls://connect.ngs.global"
  CredsPath: "/path/to/NGS-Default-CLI.creds"
  StreamName: "p2p-exchange"
  ConsumerName: "p2p-exchange-consumer"
```

憑證檔（`.creds`）從 Synadia Cloud → Get Connected 下載，**不得 commit 至 Git**。建議放置路徑：`~/.config/nats/NGS-Default-CLI.creds`

**本地開發（使用 Docker Compose NATS）：**

```yaml
Nats:
  URL: "nats://localhost:30422"
  User: "nats"
  Password: "nats@local123"
  StreamName: "p2p-exchange"
  ConsumerName: "p2p-exchange-consumer"
```

### NATS CLI 連線（Synadia Cloud）

```bash
nats context add "NGS-Default-CLI" \
  --server "tls://connect.ngs.global" \
  --creds "~/.config/nats/NGS-Default-CLI.creds" \
  --select
```

### Stream 手動建立指令

Retention policy 無法透過 `stream edit` 修改，必須刪除後重建。

```bash
# 刪除
nats stream rm P2P_ORDERS --force
nats stream rm P2P_LEDGER --force
nats stream rm P2P_NOTIFY --force

# 重建
nats stream add P2P_ORDERS \
  --subjects "order.*,order.payment.*" \
  --storage file --retention workq \
  --max-age 7d --replicas 1 --discard old --defaults

nats stream add P2P_LEDGER \
  --subjects "ledger.*" \
  --storage file --retention workq \
  --max-age 30d --replicas 1 --discard old --defaults

nats stream add P2P_NOTIFY \
  --subjects "notify.buyer.*,notify.seller.*,notify.admin.*" \
  --storage file --retention workq \
  --max-age 3d --replicas 1 --discard old --defaults
```

### Go 模組路徑

| 元件 | 路徑 |
|------|------|
| Client / Stream 初始化 | `backend/internal/infra/mq/client.go` |
| Publisher | `backend/internal/infra/mq/publisher.go` |
| Subscriber（Durable Consumer） | `backend/internal/infra/mq/subscriber.go` |
| Scheduler（週期發布） | `backend/internal/infra/mq/scheduler.go` |
| WebSocket 事件路由 | `backend/internal/job/ws_event_job.go` |
| WS 訊息格式 / Subject 常數 | `backend/pkg/ws/message.go` |
