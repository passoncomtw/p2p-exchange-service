import * as Notifications from 'expo-notifications'
import { NotificationChannelConfig } from './types'

export const NOTIFICATION_CHANNELS: NotificationChannelConfig[] = [
  {
    id: 'orders',
    name: '訂單通知',
    importance: Notifications.AndroidImportance.HIGH,
    vibrationPattern: [0, 250, 250, 250],
  },
  {
    id: 'system',
    name: '系統公告',
    importance: Notifications.AndroidImportance.DEFAULT,
  },
]
