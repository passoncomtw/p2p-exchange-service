import { Platform } from 'react-native'
import * as Device from 'expo-device'
import * as Notifications from 'expo-notifications'
import Constants from 'expo-constants'
import { navigationRef } from '@/navigation/navigationRef'
import logger from '@pkg/logger'
import { NOTIFICATION_CHANNELS } from './channels'

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: false,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
})

async function setupAndroidChannels(): Promise<void> {
  if (Platform.OS !== 'android') return
  for (const ch of NOTIFICATION_CHANNELS) {
    await Notifications.setNotificationChannelAsync(ch.id, {
      name: ch.name,
      importance: ch.importance,
      vibrationPattern: ch.vibrationPattern,
    })
  }
}

export async function registerPushToken(): Promise<string | null> {
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

  await setupAndroidChannels()

  const projectId = Constants.expoConfig?.extra?.eas?.projectId
  if (!projectId) {
    logger.warn('缺少 EAS projectId，無法取得 push token')
    return null
  }

  const tokenData = await Notifications.getExpoPushTokenAsync({ projectId })
  return tokenData.data
}

export function navigateToOrder(orderId: unknown): void {
  if (orderId && navigationRef.isReady()) {
    navigationRef.navigate('OrderDetail', { id: Number(orderId) })
  }
}

export async function handleColdStartNotification(): Promise<void> {
  const response = await Notifications.getLastNotificationResponseAsync()
  if (response) {
    navigateToOrder(response.notification.request.content.data?.orderId)
  }
}

export function onNotificationReceived(
  cb: (notification: Notifications.Notification) => void,
): Notifications.EventSubscription {
  return Notifications.addNotificationReceivedListener(cb)
}

export function onNotificationResponse(
  cb: (response: Notifications.NotificationResponse) => void,
): Notifications.EventSubscription {
  return Notifications.addNotificationResponseReceivedListener(cb)
}
