@AGENTS.md

# 通知系統架構

## 設計原則

統一使用 Redux 通知佇列管理所有 API 成功/失敗訊息，不在 component 層直接呼叫 `Alert.alert`。

## 核心元件

- **`notificationSlice`** (`src/navigation/store/slices/notificationSlice.ts`)
  - 維護 `queue: Notification[]`，FIFO 佇列
  - `pushNotification({ type, message, title? })` — 加入通知
  - `popNotification()` — 消耗第一則通知
  - 不持久化（不在 redux-persist whitelist）

- **`NotificationHandler`** (`src/components/NotificationHandler.tsx`)
  - 掛載於 `AppInner`（Provider 內、PersistGate 內）
  - 監聽 `queue[0]`，出現新通知時呼叫 `Alert.alert`
  - 使用者點擊確認後 dispatch `popNotification()`

## Saga 整合

| 觸發來源 | 處理方式 |
|---------|---------|
| 所有 `*Failure` action | `errorSaga` 統一 dispatch `pushNotification({ type: 'error', message })` |
| 各 saga 成功路徑 | 各自 dispatch `pushNotification({ type: 'success', message })` |
| 直接 API call 的 screen | 在 try/catch 裡 dispatch `pushNotification` |

## 使用規範

1. **不要在 component 呼叫 `Alert.alert` 顯示 API 回應**
2. 保留以下情況使用 `Alert.alert`：
   - 需要使用者確認的操作（如：取消掛單確認）
   - 帶有導航按鈕的提示（如：尚未新增收款帳戶）
   - input validation（如：金額格式錯誤）
3. API 成功/失敗訊息一律透過 `pushNotification` 進入佇列

```ts
// 正確：在 screen 的 try/catch 裡
dispatch(pushNotification({ type: 'success', message: '操作成功' }))
dispatch(pushNotification({ type: 'error', message: '操作失敗' }))

// 正確：在 saga 的成功路徑裡
yield put(pushNotification({ type: 'success', message: '掛單建立成功' }))
```
