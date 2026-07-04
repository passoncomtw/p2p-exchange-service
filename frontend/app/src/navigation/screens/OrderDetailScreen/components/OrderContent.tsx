import React from 'react';
import { View, Text, Pressable, StyleSheet, Platform, Alert } from 'react-native';
import { theme, commonStyles } from '@/theme';
import { formatDateTime } from '@/utils/formatUtils';
import { ORDER_STATUS_MAP } from '@/constants/orders';
import { User } from '@/interfaces/store';
import { Order } from '@/interfaces/order';
import { useAppDispatch } from '@/navigation/store/hooks';
import type { AppDispatch } from '@/navigation/store/configureStore';
import { markOrderAsPaidRequest, applyOrderRequest } from '@/navigation/store/actions/ordersActions';

const handleMarkAsPaid = (orderId: string, dispatch: AppDispatch) => () => {
  Alert.alert(
    '確認付款',
    '確認已完成匯款？',
    [
      { text: '取消', style: 'cancel' },
      {
        text: '確認',
        onPress: () => {
          dispatch(markOrderAsPaidRequest({
            orderId,
            onSuccess: () => Alert.alert('成功', '已標記為已付款'),
            onError: (error) => Alert.alert('錯誤', error || '標記付款失敗'),
          }));
        },
      },
    ]
  );
};

const handleApplyOrder = (orderId: string, dispatch: AppDispatch) => () => {
  Alert.alert(
    '確認放行',
    '確認已收到款項並放行訂單？',
    [
      { text: '取消', style: 'cancel' },
      {
        text: '確認',
        onPress: () => {
          dispatch(applyOrderRequest({
            orderId,
            onSuccess: () => Alert.alert('成功', '訂單已放行'),
            onError: (error) => Alert.alert('錯誤', error || '放行訂單失敗'),
          }));
        },
      },
    ]
  );
};

const Footer = (props: { order: Order; user: User; dispatch: AppDispatch }) => {
  const { order, user, dispatch } = props;
  const orderId = order.id.toString();
  const isBuyer = user.id === order.buyerId;
  const isSeller = user.id === order.sellerId;

  if (order.status === 'matched' && isBuyer) {
    return (
      <View style={styles.buttonContainer}>
        <Pressable style={styles.buttonPrimary} onPress={handleMarkAsPaid(orderId, dispatch)}>
          <Text style={styles.buttonPrimaryText}>匯款已完成</Text>
        </Pressable>
      </View>
    );
  }

  if (order.status === 'paid' && isSeller) {
    return (
      <View style={styles.buttonContainer}>
        <Pressable style={styles.buttonPrimary} onPress={handleApplyOrder(orderId, dispatch)}>
          <Text style={styles.buttonPrimaryText}>確認放行</Text>
        </Pressable>
      </View>
    );
  }

  if (order.status === 'matched' && isSeller) {
    return (
      <View style={styles.section}>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>等待買方付款</Text>
        </View>
      </View>
    );
  }

  if (order.status === 'paid' && isBuyer) {
    return (
      <View style={styles.section}>
        <View style={styles.infoRow}>
          <Text style={styles.infoLabel}>等待賣方確認放行</Text>
        </View>
      </View>
    );
  }

  return null;
};

const OrderContent = (props: { order: Order; user: User }) => {
  const { order, user } = props;
  const dispatch = useAppDispatch();

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

      <Footer order={order} user={user} dispatch={dispatch} />
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
});

export default OrderContent;
