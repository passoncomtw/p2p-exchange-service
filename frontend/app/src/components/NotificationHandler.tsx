import * as React from 'react';
import { Alert } from 'react-native';
import { useAppDispatch, useAppSelector } from '@/navigation/store/hooks';
import { popNotification } from '@/navigation/store/slices/notificationSlice';

const TITLE: Record<string, string> = {
  success: '成功',
  error: '錯誤',
  warning: '警告',
  info: '通知',
};

export function NotificationHandler() {
  const dispatch = useAppDispatch();
  const current = useAppSelector((s) => s.notification.queue[0]);

  React.useEffect(() => {
    if (!current) return;
    Alert.alert(
      current.title ?? TITLE[current.type] ?? '通知',
      current.message,
      [{ text: '確認', onPress: () => dispatch(popNotification()) }],
    );
  }, [current?.id]);

  return null;
}
