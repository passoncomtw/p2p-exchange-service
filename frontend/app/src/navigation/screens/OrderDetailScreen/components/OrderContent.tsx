import React, { useState, useEffect } from 'react';
import { View, Text, Pressable, StyleSheet, Platform, Alert } from 'react-native';
import { theme, commonStyles } from '@/theme';
import { formatDateTime } from '@/utils/formatUtils';
import { ORDER_STATUS_MAP } from '@/constants/orders';
import { User } from '@/interfaces/store';
import { Order } from '@/interfaces/order';
import { useAppDispatch } from '@/navigation/store/hooks';
import { markOrderAsPaidRequest, applyOrderRequest, fetchOrderDetailRequest, cancelOrderRequest, disputeOrderRequest } from '@/navigation/store/actions/ordersActions';

function useCountdown(deadline: string | null): string | null {
  const [remaining, setRemaining] = useState<string | null>(null);

  useEffect(() => {
    if (!deadline) return;

    const calc = () => {
      const diff = new Date(deadline).getTime() - Date.now();
      if (diff <= 0) {
        setRemaining('已逾時');
        return false;
      }
      const m = Math.floor(diff / 60000);
      const s = Math.floor((diff % 60000) / 1000);
      setRemaining(`${m}:${s.toString().padStart(2, '0')}`);
      return true;
    };

    if (!calc()) return;
    const id = setInterval(() => {
      if (!calc()) clearInterval(id);
    }, 1000);
    return () => clearInterval(id);
  }, [deadline]);

  return remaining;
}


const Footer = (props: { order: Order; user: User }) => {
  const { order, user } = props;
  const dispatch = useAppDispatch();
  const orderId = order.id.toString();
  const isBuyer = user.id === order.buyerId;
  const isSeller = user.id === order.sellerId;
  const countdown = useCountdown(order.status === 'matched' ? order.paymentDeadline : null);

  const refetch = () => dispatch(fetchOrderDetailRequest({ orderId }));

  const onPaidSuccess = () => {
    Alert.alert('成功', '已標記為已付款');
    refetch();
  };
  const onApplySuccess = () => {
    Alert.alert('成功', '訂單已放行');
    refetch();
  };
  const onCancelSuccess = () => {
    Alert.alert('成功', '訂單已取消，賣家資產已解凍');
    refetch();
  };
  const onDisputeSuccess = () => {
    Alert.alert('申訴已提交', '客服人員將介入處理，請耐心等候');
    refetch();
  };

  const confirmDispute = () => {
    if (Platform.OS === 'ios') {
      Alert.prompt(
        '提交申訴',
        '請說明申訴原因（例如：已付款但賣家未放行）',
        [
          { text: '返回', style: 'cancel' },
          {
            text: '確認申訴',
            style: 'destructive',
            onPress: (reason?: string) => {
              const trimmed = reason?.trim() || '買家已付款，賣家未放行';
              dispatch(
                disputeOrderRequest({
                  orderId,
                  reason: trimmed,
                  onSuccess: onDisputeSuccess,
                  onError: (e) => Alert.alert('錯誤', e || '申訴提交失敗'),
                })
              );
            },
          },
        ],
        'plain-text',
        '',
      );
    } else {
      Alert.alert('提交申訴', '確定要對此訂單提交申訴嗎？客服人員將介入處理。', [
        { text: '返回', style: 'cancel' },
        {
          text: '確認申訴',
          style: 'destructive',
          onPress: () =>
            dispatch(
              disputeOrderRequest({
                orderId,
                reason: '買家已付款，賣家未放行',
                onSuccess: onDisputeSuccess,
                onError: (e) => Alert.alert('錯誤', e || '申訴提交失敗'),
              })
            ),
        },
      ]);
    }
  };

  const confirmCancel = () => {
    if (Platform.OS === 'ios') {
      Alert.prompt(
        '取消訂單',
        '請輸入取消原因',
        [
          { text: '返回', style: 'cancel' },
          {
            text: '確認取消',
            style: 'destructive',
            onPress: (reason?: string) => {
              const trimmed = reason?.trim() || '主動取消';
              dispatch(
                cancelOrderRequest({
                  orderId,
                  reason: trimmed,
                  onSuccess: onCancelSuccess,
                  onError: (e) => Alert.alert('錯誤', e || '取消失敗'),
                })
              );
            },
          },
        ],
        'plain-text',
        '',
      );
    } else {
      Alert.alert('確認取消', '確定要取消此訂單嗎？', [
        { text: '返回', style: 'cancel' },
        {
          text: '確認取消',
          style: 'destructive',
          onPress: () =>
            dispatch(
              cancelOrderRequest({
                orderId,
                reason: '主動取消',
                onSuccess: onCancelSuccess,
                onError: (e) => Alert.alert('錯誤', e || '取消失敗'),
              })
            ),
        },
      ]);
    }
  };

  if (order.status === 'matched' && isBuyer) {
    return (
      <View>
        {countdown && (
          <View style={styles.countdownRow}>
            <Text style={styles.countdownLabel}>付款截止倒數</Text>
            <Text style={[styles.countdownValue, countdown === '已逾時' && styles.countdownExpired]}>
              {countdown}
            </Text>
          </View>
        )}
        <View style={styles.buttonContainer}>
          <Pressable
            style={styles.buttonPrimary}
            onPress={() =>
              Alert.alert('確認付款', '確認已完成匯款？', [
                { text: '取消', style: 'cancel' },
                {
                  text: '確認',
                  onPress: () =>
                    dispatch(
                      markOrderAsPaidRequest({
                        orderId,
                        onSuccess: onPaidSuccess,
                        onError: (e) => Alert.alert('錯誤', e || '標記付款失敗'),
                      })
                    ),
                },
              ])
            }
          >
            <Text style={styles.buttonPrimaryText}>匯款已完成</Text>
          </Pressable>
          <Pressable style={styles.buttonDanger} onPress={confirmCancel}>
            <Text style={styles.buttonDangerText}>取消訂單</Text>
          </Pressable>
        </View>
      </View>
    );
  }

  if (order.status === 'paid' && isSeller) {
    return (
      <View style={styles.buttonContainer}>
        <Pressable
          style={styles.buttonPrimary}
          onPress={() =>
            Alert.alert('確認放行', '確認已收到款項並放行訂單？', [
              { text: '取消', style: 'cancel' },
              {
                text: '確認',
                onPress: () =>
                  dispatch(
                    applyOrderRequest({
                      orderId,
                      onSuccess: onApplySuccess,
                      onError: (e) => Alert.alert('錯誤', e || '放行訂單失敗'),
                    })
                  ),
              },
            ])
          }
        >
          <Text style={styles.buttonPrimaryText}>確認放行</Text>
        </Pressable>
      </View>
    );
  }

  if (order.status === 'matched' && isSeller) {
    return (
      <View>
        <View style={styles.section}>
          <Text style={styles.infoLabel}>等待買方付款</Text>
          {countdown && <Text style={styles.hintText}>買方付款截止：{countdown}</Text>}
        </View>
        <View style={styles.buttonContainer}>
          <Pressable style={styles.buttonDanger} onPress={confirmCancel}>
            <Text style={styles.buttonDangerText}>取消訂單</Text>
          </Pressable>
        </View>
      </View>
    );
  }

  if (order.status === 'paid' && isBuyer) {
    return (
      <View>
        <View style={styles.section}>
          <Text style={styles.infoLabel}>等待賣方確認放行</Text>
          <Text style={styles.hintText}>若賣方長時間未放行，可提交申訴由客服介入</Text>
        </View>
        <View style={styles.buttonContainer}>
          <Pressable style={styles.buttonDanger} onPress={confirmDispute}>
            <Text style={styles.buttonDangerText}>提交申訴</Text>
          </Pressable>
        </View>
      </View>
    );
  }

  return null;
};

const OrderContent = (props: { order: Order; user: User }) => {
  const { order, user } = props;

  return (
    <View>
      <View style={styles.section}>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>訂單編號</Text>
          <View style={styles.infoValueRow}>
            <Text style={styles.infoValueMonospace}>{order.orderNo}</Text>
          </View>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>訂單狀態</Text>
          <View style={styles.infoValueRow}>
            <Text style={styles.infoValue}>{ORDER_STATUS_MAP[order.status]?.label ?? order.status}</Text>
          </View>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>掛單類型</Text>
          <View style={styles.infoValueRow}>
            <Text style={styles.infoValue}>{order.listingType === 'buy' ? '買幣掛單' : '賣幣掛單'}</Text>
          </View>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>訂單成立時間</Text>
          <View style={styles.infoValueRow}>
            <Text style={styles.infoValue}>{formatDateTime(order.createdAt)}</Text>
          </View>
        </View>
      </View>

      <View style={styles.section}>
        <View style={styles.infoRow}>
          <Text style={styles.sectionTitle}>交易資料</Text>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>加密貨幣數量</Text>
          <View style={styles.infoValueRow}>
            <Text style={styles.infoValue}>{order.cryptoAmount} {order.cryptoCurrency}</Text>
          </View>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>單價</Text>
          <View style={styles.infoValueRow}>
            <Text style={styles.infoValue}>{order.price} {order.fiatCurrency}</Text>
          </View>
        </View>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>法幣金額</Text>
          <View style={styles.infoValueRow}>
            <Text style={styles.infoValue}>{order.fiatAmount} {order.fiatCurrency}</Text>
          </View>
        </View>
        {order.totalFee > 0 && (
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>手續費</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{order.totalFee}</Text>
            </View>
          </View>
        )}
      </View>

      <Footer order={order} user={user} />
    </View>
  );
};

const styles = StyleSheet.create({
  section: {
    backgroundColor: theme.background.primary,
    padding: theme.spacing.lg,
    marginTop: theme.spacing.md,
  },
  sectionTitle: {
    fontSize: theme.fontSize.lg,
    fontWeight: '600',
    color: theme.text.primary,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: theme.spacing.md,
  },
  infoLabel: {
    fontSize: theme.fontSize.md,
    color: theme.text.secondary,
  },
  infoValue: {
    fontSize: theme.fontSize.md,
    color: theme.text.primary,
    fontWeight: '500',
  },
  infoValueMonospace: {
    fontSize: theme.fontSize.md,
    color: theme.text.primary,
    fontWeight: '500',
    fontFamily: Platform.OS === 'ios' ? 'Courier' : 'monospace',
  },
  infoValueRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  buttonContainer: {
    padding: theme.spacing.lg,
    backgroundColor: theme.background.primary,
    borderTopWidth: 1,
    borderTopColor: theme.border.default,
  },
  buttonPrimary: {
    ...commonStyles.buttonPrimary,
    marginBottom: theme.spacing.md,
  },
  buttonPrimaryText: {
    ...commonStyles.buttonPrimaryText,
  },
  countdownRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: theme.spacing.lg,
    paddingVertical: theme.spacing.md,
    backgroundColor: '#FFF8E1',
  },
  countdownLabel: {
    fontSize: theme.fontSize.md,
    color: theme.text.secondary,
  },
  countdownValue: {
    fontSize: theme.fontSize.xl,
    fontWeight: '700',
    color: '#F57C00',
  },
  countdownExpired: {
    color: theme.status.error,
  },
  hintText: {
    fontSize: theme.fontSize.sm,
    color: theme.text.tertiary,
    marginTop: theme.spacing.sm,
  },
  buttonDanger: {
    borderWidth: 1,
    borderColor: theme.status.error,
    borderRadius: 8,
    paddingVertical: theme.spacing.md,
    alignItems: 'center',
    marginTop: theme.spacing.sm,
  },
  buttonDangerText: {
    fontSize: theme.fontSize.md,
    fontWeight: '600',
    color: theme.status.error,
  },
});

export default OrderContent;
