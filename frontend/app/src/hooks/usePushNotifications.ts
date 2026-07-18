import { useEffect, useRef } from 'react'
import { NotificationManager } from '@frontend-pkg/notifications'
import { ExpoNotificationProvider } from '@pkg/notifications'
import { navigationRef } from '@/navigation/navigationRef'
import { profileApi } from '@/apis/profileApi'
import logger from '@pkg/logger'

const manager = new NotificationManager(new ExpoNotificationProvider())

function navigateToOrder(orderId: unknown) {
  if (orderId && navigationRef.isReady()) {
    navigationRef.navigate('OrderDetail', { id: Number(orderId) })
  }
}

export function usePushNotifications(isAuthenticated: boolean): void {
  const unsubscribers = useRef<Array<() => void>>([])

  useEffect(() => {
    if (!isAuthenticated) return

    manager
      .setup()
      .then(async (token) => {
        if (!token) return
        try {
          await profileApi.registerPushToken(token)
          logger.info('Push token 已註冊', { token })
        } catch (err) {
          logger.warn('Push token 註冊失敗', err)
        }
      })
      .catch((err) => logger.warn('取得 push token 失敗', err))

    manager.getInitialNotification().then((msg) => {
      if (msg) navigateToOrder(msg.data?.orderId)
    })

    unsubscribers.current = [
      manager.onMessage((msg) => {
        logger.info('收到推送通知', { title: msg.title, body: msg.body })
      }),
      manager.onNotificationTap((msg) => {
        navigateToOrder(msg.data?.orderId)
      }),
    ]

    return () => {
      unsubscribers.current.forEach((unsub) => unsub())
      unsubscribers.current = []
    }
  }, [isAuthenticated])
}
