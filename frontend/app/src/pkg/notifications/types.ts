export type NotificationChannel = 'orders' | 'system'
export type NotificationPriority = 'default' | 'normal' | 'high'

export interface NotificationChannelConfig {
  id: NotificationChannel
  name: string
  importance: number
  vibrationPattern?: number[]
}
