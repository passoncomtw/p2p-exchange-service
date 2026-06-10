import React from 'react';
import { View, Text, StyleSheet, Pressable } from 'react-native';
import { theme } from '@/theme';
import { ORDER_STATUS_MAP } from '@/constants/orders';
import type { Order } from '@/interfaces';

interface OrderItemProps {
  order: Order;
  onPress: (orderId: string) => void;
}

const fmt = (num: number): string =>
  (num ?? 0).toLocaleString('zh-TW');

const getStatusStyle = (status: string) => {
  switch (status) {
    case 'completed':
      return { container: styles.statusCompleted, text: styles.statusTextCompleted };
    case 'cancelled':
    case 'timeout':
      return { container: styles.statusCancelled, text: styles.statusTextCancelled };
    case 'disputed':
      return { container: styles.statusDispute, text: styles.statusTextDispute };
    case 'paid':
    case 'releasing':
      return { container: styles.statusPendingRelease, text: styles.statusTextPendingRelease };
    default:
      return { container: styles.statusPendingPayment, text: styles.statusTextPendingPayment };
  }
};

export default function OrderItem({ order, onPress }: OrderItemProps) {
  const statusInfo = ORDER_STATUS_MAP[order.status] ?? { label: order.status };
  const statusStyle = getStatusStyle(order.status);

  return (
    <Pressable
      style={({ pressed }) => [styles.container, pressed && styles.containerPressed]}
      onPress={() => onPress(order.id.toString())}
    >
      <View style={[styles.statusContainer, statusStyle.container]}>
        <Text style={[styles.statusText, statusStyle.text]}>{statusInfo.label}</Text>
      </View>

      <View style={styles.row}>
        <Text style={styles.label}>訂單編號</Text>
        <Text style={styles.text}>{order.orderNo}</Text>
      </View>
      <View style={styles.row}>
        <Text style={styles.label}>掛單類型</Text>
        <Text style={styles.text}>{order.listingType === 'buy' ? '買幣' : '賣幣'}</Text>
      </View>
      <View style={styles.row}>
        <Text style={styles.label}>數量</Text>
        <Text style={styles.text}>{fmt(order.cryptoAmount)} {order.cryptoCurrency}</Text>
      </View>
      <View style={styles.row}>
        <Text style={styles.label}>交易金額</Text>
        <Text style={styles.text}>{order.fiatCurrency} {fmt(order.fiatAmount)}</Text>
      </View>
      <View style={styles.row}>
        <Text style={styles.label}>單價</Text>
        <Text style={styles.text}>{fmt(order.price)}</Text>
      </View>
      {order.cancelReason ? (
        <View style={styles.row}>
          <Text style={styles.label}>取消原因</Text>
          <Text style={[styles.text, { color: theme.status.error }]}>{order.cancelReason}</Text>
        </View>
      ) : null}
      <View style={styles.row}>
        <Text style={styles.label}>創建時間</Text>
        <Text style={styles.text}>{order.createdAt}</Text>
      </View>
    </Pressable>
  );
}

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#fff',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  containerPressed: { backgroundColor: '#F8F8F8' },
  statusContainer: {
    alignSelf: 'flex-start',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
    marginBottom: 12,
  },
  statusPendingPayment:  { backgroundColor: '#FFF3E0' },
  statusPendingRelease:  { backgroundColor: '#E3F2FD' },
  statusCompleted:       { backgroundColor: '#E8F5E9' },
  statusCancelled:       { backgroundColor: '#F5F5F5' },
  statusDispute:         { backgroundColor: '#FFEBEE' },
  statusText:            { fontSize: 13, fontWeight: '600' },
  statusTextPendingPayment:  { color: theme.status.warning },
  statusTextPendingRelease:  { color: theme.status.info },
  statusTextCompleted:       { color: theme.status.success },
  statusTextCancelled:       { color: theme.text.tertiary },
  statusTextDispute:         { color: theme.status.error },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 10,
  },
  label: { fontSize: 14, color: '#999', flex: 0.4 },
  text:  { fontSize: 14, color: '#333', fontWeight: '500', flex: 0.6, textAlign: 'right' },
});
