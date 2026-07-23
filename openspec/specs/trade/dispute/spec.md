# 糾紛處理（Trade Dispute）

## 概述

當買方已付款（訂單狀態為 `paid`）但雙方無法達成共識時，任一方可發起糾紛，由後台管理員介入仲裁。

---

## 狀態機

糾紛狀態屬於訂單狀態機的一部分，詳見 [`core/order-lifecycle`](../../core/order-lifecycle/spec.md)。

```
paid → disputed → completed (admin: release USDT to buyer)
               → cancelled  (admin: refund USDT to seller)
```

---

## 不變式

1. **糾紛只能從 paid 發起**：`matched` 狀態下無法糾紛（買方尚未付款）。
2. **雙方皆可發起**：買方（付款後對方不確認）或賣方（對方標記付款但未收到款項）都可發起。
3. **進入 disputed 後不可 confirm**：賣方在糾紛期間無法單方面確認收款。
4. **admin 仲裁後不可撤銷**：`completed` 或 `cancelled` 為終態。
5. **資產凍結期間**：糾紛期間賣方的加密貨幣維持凍結，等待仲裁結果。

---

## 仲裁規則

| admin 決定 | 操作 | 結果 |
|-----------|------|------|
| `complete`（買方勝） | TransferInTx（凍結→買方）| 訂單→completed，USDT 轉至買方 |
| `refund`（賣方勝） | UnfreezeInTx + RestoreAmountInTx | 訂單→cancelled，USDT 退還賣方，掛單恢復餘量 |

---

## 邊界條件

### BC-1：糾紛發起後買方緊急匯款
- 情境：買方發起糾紛後，實際匯款到帳，賣方確認收款。
- 規則：disputed 狀態下賣方無法 confirm，必須透過 admin 仲裁 complete。

### BC-2：admin 仲裁期間掛單已被取消（listing.status != active）
- 情境：admin 選擇 refund，但 listing 已 completed 或 cancelled。
- 規則：RestoreAmountInTx 僅在 listing.status=active 或 paused 時恢復餘量；若已 completed/cancelled，僅解凍資產不恢復餘量。

### BC-3：雙方同時發起糾紛
- 情境：買方與賣方在毫秒內同時呼叫 dispute API。
- 規則：dispute API 以 `UPDATE WHERE status='paid'` 條件更新，只有一個能成功。

### BC-4：admin 帳號操作後被停用
- 情境：admin A 完成仲裁後帳號被停用，操作記錄是否保留。
- 規則：order_status_logs 記錄 operatorType=admin，操作記錄永久保留，不受帳號狀態影響。

### BC-5：disputed 持續過長時間（無人仲裁）
- 情境：糾紛建立 30 天無人處理。
- 規則：目前無自動逾時機制，由後台監控介入。未來可加入 SLA 告警。

---

## API 摘要

### 發起糾紛
`PUT /app/orders/:id/dispute`

- 前置條件：訂單狀態為 `paid`，呼叫者為買方或賣方
- 副作用：status→disputed，通知後台管理員
- 回傳：更新後的訂單狀態

### 後台仲裁
`PUT /backend/orders/:id/resolve`

請求：`{ action: "complete" | "refund", reason: string }`

- 前置條件：訂單狀態為 `disputed`
- action=complete 副作用：TransactCtx(status→completed, EscrowRecord(release), Wallet.TransferInTx)
- action=refund 副作用：TransactCtx(status→cancelled, EscrowRecord(refund), Wallet.UnfreezeInTx, Listing.RestoreAmountInTx)
- 發布：NATS `order.status.changed`

---

## 相關檔案

| 層級 | 路徑 |
|------|------|
| App Handler | `backend/internal/handler/apporderhandler.go` |
| App Logic | `backend/internal/logic/apporderlogic.go: DisputeOrderLogic` |
| Backend Logic | `backend/internal/logic/backendorderlogic.go: BackendResolveOrderLogic` |
| Backend Handler | `backend/internal/handler/backendorderhandler.go` |
