# P2P C2C 加密貨幣交易所 — 專案概覽

## 專案目標

一個 P2P（Peer-to-Peer）C2C 加密貨幣交易平台，讓用戶之間直接撮合買賣加密貨幣與法幣，平台擔任擔保方。

## 技術棧

| 層級 | 技術 |
|------|------|
| Backend | Go 1.23+, go-zero v1.10.1 (REST framework) |
| Database | PostgreSQL (Neon cloud), go-zero sqlx |
| Cache | Redis (分散式鎖、session) |
| Message Queue | NATS JetStream |
| WebSocket | gorilla/websocket v1.5.3 |
| App (Mobile) | React Native, Expo SDK 56, TypeScript, Redux/Saga |
| Web (Admin) | React 18, MUI v5, Vite, JSX |
| Auth | JWT (App 與 Backend 各自獨立 secret) |
| 金流 | ECPay（法幣入金 webhook）|
| 加密貨幣 | Tron（TRX/USDT），鏈上掃描 + 熱錢包出金 |

## 領域表

| Domain | 說明 | 主要 Spec 位置 |
|--------|------|---------------|
| core/order-lifecycle | 訂單狀態機（matched→paid→releasing→completed，以及 cancelled/timeout/disputed） | `specs/core/order-lifecycle/` |
| core/trade-lifecycle | 掛單（Listing）生命週期（active→paused/completed/cancelled） | `specs/core/trade-lifecycle/` |
| core/wallet | 錢包餘額、凍結/解凍、轉帳原子性 | `specs/core/wallet/` |
| core/identity | 用戶身份（App 用戶 vs 後台帳號）| `specs/core/identity/` |
| kyc/verification | KYC 驗證流程（預留）| `specs/kyc/verification/` |
| order/posting | 掛單發布規則 | `specs/order/posting/` |
| order/matching | 訂單撮合邏輯 | `specs/order/matching/` |
| order/cancellation | 訂單取消條件 | `specs/order/cancellation/` |
| trade/execution | 付款確認流程 | `specs/trade/execution/` |
| trade/dispute | 糾紛處理流程 | `specs/trade/dispute/` |
| trade/completion | 交易完成與結算 | `specs/trade/completion/` |
| wallet/deposit | 入金（法幣 ECPay / 加密貨幣鏈上）| `specs/wallet/deposit/` |
| wallet/withdrawal | 出金（法幣銀行轉帳 / 加密貨幣鏈上）| `specs/wallet/withdrawal/` |
| wallet/freeze | 凍結/解凍機制 | `specs/wallet/freeze/` |
| cs-portal | 客服後台：交易查詢、糾紛處理、手動放行 | `specs/cs-portal/` |
| admin-portal | 管理後台：用戶管理、訂單監控、風控 | `specs/admin-portal/` |

## 核心狀態機

### 訂單（Order）狀態
```
matched → paid → releasing → completed
        ↘ cancelled (任意階段均可)
        ↘ timeout   (付款期限超過)
        ↘ disputed  (糾紛中)
```

### 掛單（Listing）狀態
```
active ⇌ paused
active → completed (成交量耗盡)
active → cancelled (用戶手動取消)
```

### 法幣提領（FiatWithdrawal）狀態
```
pending → approved
        → rejected
```

### 加密貨幣出入金狀態
```
CryptoDeposit:    pending → confirmed / failed
CryptoWithdraw:   pending → broadcasting → confirmed / failed
FiatDeposit:      pending → paid / failed
```

## 架構約定

- **Clean Architecture**：handler（HTTP 層）→ service/logic（商業邏輯）→ repository/model（資料存取）
- **Response 格式**：`{ code, message, data }`
- **分散式鎖**：Redis `AcquireLock`，操作錢包前必須持鎖
- **事務**：跨表操作使用 `sqlx.TransactCtx`
- **NATS subjects**：
  - `order.status.changed`
  - `wallet.balance.changed`
  - `order.timeout.check`
  - `notification.push`
- **WebSocket**：
  - `/ws/app`：App 用戶端，per-userID 連線
  - `/ws/backend`：後台廣播，所有管理員共享

## 讀者指南

Spec 採用「現狀描述 + 不變式 + 邊界條件」格式。每份 spec 包含：
1. **狀態機** — 合法狀態與轉換觸發條件
2. **不變式** — 任何時刻必須成立的業務規則
3. **邊界條件** — 容易遺漏的角落案例
4. **API 摘要** — 相關 endpoint（非完整 OpenAPI，聚焦行為）

## Q&A 追蹤

| # | 問題 | 狀態 | 答案/決策 |
|---|------|------|-----------|
| 1 | 訂單付款期限超時後，由 job 自動取消還是用戶觸發？ | 已解答 | NATS `order.timeout.check` consumer 自動取消，釋放凍結資產 |
| 2 | 平台費用在哪個時間點扣除？ | 待釐清 | — |
| 3 | KYC 是否影響掛單或交易額度？ | 待規劃 | — |
