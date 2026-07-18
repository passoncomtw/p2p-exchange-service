import { useEffect, useRef } from 'react';
import { Platform } from 'react-native';
import * as Device from 'expo-device';
import * as Notifications from 'expo-notifications';
import Constants from 'expo-constants';
import { profileApi } from '@/apis/profileApi';
import { navigationRef } from '@/navigation/navigationRef';
import logger from '@pkg/logger';

Notifications.setNotificationHandler({
  handleNotification: async () => ({
    shouldShowAlert: true,
    shouldPlaySound: true,
    shouldSetBadge: false,
    shouldShowBanner: true,
    shouldShowList: true,
  }),
});

async function registerForPushNotifications(): Promise<string | null> {
  if (!Device.isDevice) {
    return null; // 模擬器不支援 push
  }

  const { status: existingStatus } = await Notifications.getPermissionsAsync();
  let finalStatus = existingStatus;

  if (existingStatus !== 'granted') {
    const { status } = await Notifications.requestPermissionsAsync();
    finalStatus = status;
  }

  if (finalStatus !== 'granted') {
    logger.warn('推送通知權限未授予');
    return null;
  }

  if (Platform.OS === 'android') {
    await Notifications.setNotificationChannelAsync('default', {
      name: 'P2P 訂單通知',
      importance: Notifications.AndroidImportance.MAX,
      vibrationPattern: [0, 250, 250, 250],
    });
  }

  const projectId = Constants.expoConfig?.extra?.eas?.projectId;
  if (!projectId) {
    logger.warn('缺少 EAS projectId，無法取得 push token');
    return null;
  }

  const tokenData = await Notifications.getExpoPushTokenAsync({ projectId });
  return tokenData.data;
}

export function usePushNotifications(isAuthenticated: boolean): void {
  const notificationListener = useRef<Notifications.EventSubscription | null>(null);
  const responseListener = useRef<Notifications.EventSubscription | null>(null);

  useEffect(() => {
    if (!isAuthenticated) return;

    registerForPushNotifications()
      .then(async (token) => {
        if (!token) return;
        try {
          await profileApi.registerPushToken(token);
          logger.info('Push token 已註冊', { token });
        } catch (err) {
          logger.warn('Push token 註冊失敗', err);
        }
      })
      .catch((err) => logger.warn('取得 push token 失敗', err));

    notificationListener.current = Notifications.addNotificationReceivedListener((notification) => {
      logger.info('收到推送通知', {
        title: notification.request.content.title,
        body: notification.request.content.body,
      });
    });

    responseListener.current = Notifications.addNotificationResponseReceivedListener((response) => {
      const orderId = response.notification.request.content.data?.orderId;
      if (orderId && navigationRef.isReady()) {
        navigationRef.navigate('OrderDetail', { id: Number(orderId) });
      }
    });

    return () => {
      notificationListener.current?.remove();
      responseListener.current?.remove();
    };
  }, [isAuthenticated]);
}
