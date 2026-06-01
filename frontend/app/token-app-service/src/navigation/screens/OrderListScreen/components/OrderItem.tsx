/**
 * OrderItem - 訂單項目組件
 */

import React from 'react';
import { View, Text, StyleSheet, Pressable, StyleProp, ViewStyle, TextStyle } from 'react-native';
import { theme } from '@/theme';
import { ORDER_STATUS_MAP } from '@/constants/orders';
import { OrderItem as ApiOrderItem } from '@/interfaces';

interface OrderItemProps {
  order: ApiOrderItem;
  onPress: (orderId: string) => void;
}

const formatNumber = (num: number): string => {
  if (num === undefined || num === null) {
    return '0';
  }
  return num.toLocaleString('zh-TW');
};

// 根據狀態類型設置樣式（與訂單詳情頁面保持一致）
const getStatusStyle = (status: number): { container: StyleProp<ViewStyle>, text: StyleProp<TextStyle> } => {
  switch (status) {
    case 2:
      return {
        container: styles.statusCompleted,
        text: styles.statusTextCompleted,
      };
    case 3:
      return {
        container: styles.statusCancelled,
        text: styles.statusTextCancelled,
      };
    case 4:
      return {
        container: styles.statusDispute,
        text: styles.statusTextDispute,
      };
    case 1:
      return {
        container: styles.statusPendingRelease,
        text: styles.statusTextPendingRelease,
      };
    case 0:
    default:
      return {
        container: styles.statusPendingPayment,
        text: styles.statusTextPendingPayment,
      };
  }
};

export default function OrderItem({ order, onPress }: OrderItemProps) {
  

  // 根據狀態類型獲取時間標籤和值
  const getTimeLabel = () => {
    if (order.finishAt) return '取消時間';
    return '創建時間';
  };

  const statusStyle = getStatusStyle(order.status);
  const isBuyPendingOrder = order.pendingOrder.type === 0;
  const bankCard = isBuyPendingOrder ? order.bankcard : order.pendingOrder.bankcard;
  
  return (
    <Pressable
      style={({ pressed }) => [
        styles.container,
        pressed && styles.containerPressed,
      ]}
      onPress={() => onPress(order.id)}
    >
      {/* 訂單狀態 */}
      <View style={[styles.statusContainer, statusStyle.container]}>
        <Text style={[styles.statusText, statusStyle.text]}>
          {ORDER_STATUS_MAP[order.status].label}
        </Text>
      </View>

      {/* 訂單詳情 */}
      <View style={styles.row}>
        <Text style={styles.label}>訂單編號</Text>
        <Text style={styles.text}>{order.id}</Text>
      </View>
      <View style={styles.row}>
        <Text style={styles.label}>掛單類型</Text>
        <Text style={styles.text}>{order.pendingOrder.type === 0 ? '買幣' : '賣幣'}</Text>
      </View>
      <View style={styles.row}>
        <Text style={styles.label}>數量</Text>
        <Text style={styles.text}>{formatNumber(order.amount)} E幣</Text>
      </View>
      <View style={styles.row}>
        <Text style={styles.label}>交易金額</Text>
        <Text style={styles.text}>CNY ¥{formatNumber(order.amount)}</Text>
      </View>
      <View style={styles.row}>
        <Text style={styles.label}>{getTimeLabel()}</Text>
        <Text style={styles.text}>{order.createdAt}</Text>
      </View>

      {/* 付款資訊（銀行卡資訊） */}
        <>
          <View style={styles.divider} />
          <Text style={styles.sectionTitle}>付款資訊</Text>
            <View style={styles.row}>
              <Text style={styles.label}>銀行</Text>
            <Text style={styles.text}>{bankCard.bank.bankName}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.label}>卡號</Text>
            <Text style={styles.text}>{bankCard.cardNumber}</Text>
          </View>
          <View style={styles.row}>
            <Text style={styles.label}>戶名</Text>
            <Text style={styles.text}>{bankCard.name}</Text>
          </View>
        </>
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
  containerPressed: {
    backgroundColor: '#F8F8F8',
  },
  statusContainer: {
    alignSelf: 'flex-start',
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
    marginBottom: 12,
  },
  statusPendingPayment: {
    backgroundColor: '#FFF3E0', // 待付款 - 警告色背景（淺黃）
  },
  statusPendingRelease: {
    backgroundColor: '#E3F2FD', // 待放行 - 資訊色背景（淺藍）
  },
  statusCompleted: {
    backgroundColor: '#E8F5E9', // 已完成 - 成功色背景（淺綠）
  },
  statusCancelled: {
    backgroundColor: '#F5F5F5', // 已取消 - 灰色背景
  },
  statusDispute: {
    backgroundColor: '#FFEBEE', // 申訴中 - 錯誤色背景（淺紅）
  },
  statusText: {
    fontSize: 13,
    fontWeight: '600',
  },
  statusTextPendingPayment: {
    color: theme.status.warning, // '#FF9800'
  },
  statusTextPendingRelease: {
    color: theme.status.info, // '#2196F3'
  },
  statusTextCompleted: {
    color: theme.status.success, // '#4CAF50'
  },
  statusTextCancelled: {
    color: theme.text.tertiary, // '#999999'
  },
  statusTextDispute: {
    color: theme.status.error, // '#F44336'
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 10,
  },
  label: {
    fontSize: 14,
    color: '#999',
    flex: 0.4,
  },
  text: {
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
    flex: 0.6,
    textAlign: 'right',
  },
  divider: {
    height: 1,
    backgroundColor: '#E5E5E5',
    marginVertical: 12,
  },
  sectionTitle: {
    fontSize: 15,
    fontWeight: '600',
    color: '#333',
    marginBottom: 10,
  },
});

