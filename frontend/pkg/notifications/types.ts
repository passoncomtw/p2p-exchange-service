export type NotificationChannel = 'orders' | 'system'
export type NotificationPriority = 'default' | 'normal' | 'high'

export interface NotificationMessage {
  title?: string
  body?: string
  data?: Record<string, unknown>
  channel?: NotificationChannel
  priority?: NotificationPriority
}

/**
 * 平台無關的 notification provider 介面。
 * 各平台（Expo / Firebase / WebSocket）各自實作此介面，
 * NotificationManager 透過此介面操作，不依賴具體技術。
 */
export interface INotificationProvider {
  /** 初始化：請求權限、建立 channel、回傳 push token（無 token 時回傳 null） */
  setup(): Promise<string | null>
  /** 前景收到通知時的 listener，回傳 unsubscribe 函數 */
  onMessage(cb: (message: NotificationMessage) => void): () => void
  /** 用戶點擊通知時的 listener，回傳 unsubscribe 函數 */
  onNotificationTap(cb: (message: NotificationMessage) => void): () => void
  /** App 從完全關閉狀態被通知喚起，取得那則通知（cold start） */
  getInitialNotification(): Promise<NotificationMessage | null>
}
