import { useEffect, useRef } from 'react'
import * as Notifications from 'expo-notifications'
import { profileApi } from '@/apis/profileApi'
import {
  registerPushToken,
  handleColdStartNotification,
  navigateToOrder,
  onNotificationReceived,
  onNotificationResponse,
} from '@pkg/notifications'
import logger from '@pkg/logger'

export function usePushNotifications(isAuthenticated: boolean): void {
  const notificationListener = useRef<Notifications.EventSubscription | null>(null)
  const responseListener = useRef<Notifications.EventSubscription | null>(null)

  useEffect(() => {
    if (!isAuthenticated) return

    registerPushToken()
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

    handleColdStartNotification()

    notificationListener.current = onNotificationReceived((notification) => {
      logger.info('收到推送通知', {
        title: notification.request.content.title,
        body: notification.request.content.body,
      })
    })

    responseListener.current = onNotificationResponse((response) => {
      navigateToOrder(response.notification.request.content.data?.orderId)
    })

    return () => {
      notificationListener.current?.remove()
      responseListener.current?.remove()
    }
  }, [isAuthenticated])
}
