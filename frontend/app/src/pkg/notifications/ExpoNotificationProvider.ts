import { Platform } from 'react-native'
import * as Device from 'expo-device'
import * as Notifications from 'expo-notifications'
import Constants from 'expo-constants'
import type { INotificationProvider, NotificationMessage } from '@frontend-pkg/notifications'
import logger from '@pkg/logger'

const ANDROID_CHANNELS: Array<{
  id: string
  name: string
  importance: number
  vibrationPattern?: number[]
}> = [
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

function toMessage(content: Notifications.NotificationContent): NotificationMessage {
  return {
    title: content.title ?? undefined,
    body: content.body ?? undefined,
    data: content.data as Record<string, unknown>,
  }
}

export class ExpoNotificationProvider implements INotificationProvider {
  constructor() {
    Notifications.setNotificationHandler({
      handleNotification: async () => ({
        shouldShowAlert: true,
        shouldPlaySound: true,
        shouldSetBadge: false,
        shouldShowBanner: true,
        shouldShowList: true,
      }),
    })
  }

  async setup(): Promise<string | null> {
    if (!Device.isDevice) return null

    const { status: existingStatus } = await Notifications.getPermissionsAsync()
    let finalStatus = existingStatus

    if (existingStatus !== 'granted') {
      const { status } = await Notifications.requestPermissionsAsync()
      finalStatus = status
    }

    if (finalStatus !== 'granted') {
      logger.warn('推送通知權限未授予')
      return null
    }

    if (Platform.OS === 'android') {
      for (const ch of ANDROID_CHANNELS) {
        await Notifications.setNotificationChannelAsync(ch.id, {
          name: ch.name,
          importance: ch.importance,
          vibrationPattern: ch.vibrationPattern,
        })
      }
    }

    const projectId = Constants.expoConfig?.extra?.eas?.projectId
    if (!projectId) {
      logger.warn('缺少 EAS projectId，無法取得 push token')
      return null
    }

    const tokenData = await Notifications.getExpoPushTokenAsync({ projectId })
    return tokenData.data
  }

  onMessage(cb: (message: NotificationMessage) => void): () => void {
    const sub = Notifications.addNotificationReceivedListener((notification) => {
      cb(toMessage(notification.request.content))
    })
    return () => sub.remove()
  }

  onNotificationTap(cb: (message: NotificationMessage) => void): () => void {
    const sub = Notifications.addNotificationResponseReceivedListener((response) => {
      cb(toMessage(response.notification.request.content))
    })
    return () => sub.remove()
  }

  async getInitialNotification(): Promise<NotificationMessage | null> {
    try {
      const response = await Notifications.getLastNotificationResponseAsync()
      if (!response) return null
      return toMessage(response.notification.request.content)
    } catch {
      return null
    }
  }
}
