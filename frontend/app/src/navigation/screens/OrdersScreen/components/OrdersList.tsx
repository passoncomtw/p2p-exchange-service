/**
 * OrdersList - 掛單列表組件
 */

import React from 'react';
import { View, Text, StyleSheet, Pressable, ScrollView, Alert } from 'react-native';

export interface PendingOrder {
  id: string;
  type: 'buy' | 'sell';
  status: 'active' | 'locked';
  amount: number;
  totalPrice: number;
  minAmount: number;
  paymentTimeout: number;
  createdAt: string;
}

interface OrdersListProps {
  orders: PendingOrder[];
  showSuccessAlert?: boolean;
  onDelete: (orderId: string) => void;
}

export default function OrdersList({ 
  orders, 
  showSuccessAlert = true,
  onDelete,
}: OrdersListProps) {
  const formatNumber = (num: number) => {
    return num.toLocaleString('zh-TW');
  };

  return (
    <ScrollView style={styles.container} showsVerticalScrollIndicator={false}>
      {/* 成功提示 */}
      {showSuccessAlert && (
        <View style={styles.successAlert}>
          <View style={styles.successAlertHeader}>
            <Text style={styles.successAlertIcon}>✅</Text>
            <Text style={styles.successAlertTitle}>掛單成功！</Text>
          </View>
          <Text style={styles.successAlertText}>
            您的掛單已成功建立，請保持在線等待交易對手
          </Text>
        </View>
      )}

      {/* 掛單卡片列表 */}
      {orders.map((order) => (
        <View key={order.id} style={styles.pendingCard}>
          {/* 頂部：類型 + 狀態 */}
          <View style={styles.pendingHeader}>
            <View style={[
              styles.pendingType,
              order.type === 'buy' ? styles.pendingTypeBuy : styles.pendingTypeSell
            ]}>
              <Text style={styles.pendingTypeText}>
                {order.type === 'buy' ? '買入' : '賣出'}
              </Text>
            </View>
            <View style={[
              styles.pendingStatus,
              order.status === 'active' ? styles.pendingStatusActive : styles.pendingStatusLocked
            ]}>
              <Text style={[
                styles.pendingStatusText,
                order.status === 'active' ? styles.pendingStatusTextActive : styles.pendingStatusTextLocked
              ]}>
                {order.status === 'active' ? '進行中' : '已鎖定'}
              </Text>
            </View>
          </View>

          {/* 詳細資訊 */}
          <View style={styles.labelTextRow}>
            <Text style={styles.label}>數量</Text>
            <Text style={styles.text}>{formatNumber(order.amount)} E幣</Text>
          </View>
          <View style={styles.labelTextRow}>
            <Text style={styles.label}>交易金額</Text>
            <Text style={styles.text}>CNY ¥{formatNumber(order.totalPrice)}</Text>
          </View>
          <View style={styles.labelTextRow}>
            <Text style={styles.label}>最小交易量</Text>
            <Text style={styles.text}>{formatNumber(order.minAmount)} E幣</Text>
          </View>
          <View style={styles.labelTextRow}>
            <Text style={styles.label}>支付時效</Text>
            <Text style={styles.text}>{order.paymentTimeout} 分鐘</Text>
          </View>
          <View style={styles.labelTextRow}>
            <Text style={styles.label}>創建時間</Text>
            <Text style={styles.text}>{order.createdAt}</Text>
          </View>

          {/* 操作按鈕 */}
          <View style={styles.pendingActions}>
            <Pressable 
              style={({ pressed }) => [
                styles.btnSmall,
                styles.btnRed,
                pressed && styles.btnPressed,
              ]}
              onPress={() => onDelete(order.id)}
            >
              <Text style={styles.btnSmallTextWhite}>刪除</Text>
            </Pressable>
          </View>
        </View>
      ))}
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
  },
  successAlert: {
    backgroundColor: '#E8F5E9',
    padding: 16,
    margin: 12,
    borderRadius: 8,
    borderLeftWidth: 4,
    borderLeftColor: '#4CAF50',
  },
  successAlertHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  successAlertIcon: {
    fontSize: 24,
    marginRight: 12,
  },
  successAlertTitle: {
    fontWeight: 'bold',
    color: '#2E7D32',
    fontSize: 16,
  },
  successAlertText: {
    color: '#666',
    fontSize: 14,
    lineHeight: 20,
  },
  pendingCard: {
    backgroundColor: '#fff',
    margin: 12,
    marginTop: 0,
    padding: 16,
    borderRadius: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  pendingHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  pendingType: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
  },
  pendingTypeBuy: {
    backgroundColor: '#E3F2FD',
  },
  pendingTypeSell: {
    backgroundColor: '#FFF3E0',
  },
  pendingTypeText: {
    fontSize: 14,
    fontWeight: '600',
  },
  pendingStatus: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
  },
  pendingStatusActive: {
    backgroundColor: '#E8F5E9',
  },
  pendingStatusLocked: {
    backgroundColor: '#FFF3E0',
  },
  pendingStatusText: {
    fontSize: 14,
    fontWeight: '500',
  },
  pendingStatusTextActive: {
    color: '#2E7D32',
  },
  pendingStatusTextLocked: {
    color: '#F57C00',
  },
  labelTextRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  label: {
    fontSize: 14,
    color: '#999',
  },
  text: {
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
  },
  pendingActions: {
    flexDirection: 'row',
    justifyContent: 'flex-end',
    marginTop: 16,
    gap: 8,
  },
  btnSmall: {
    paddingVertical: 8,
    paddingHorizontal: 16,
    borderRadius: 6,
    minWidth: 60,
    alignItems: 'center',
  },
  btnGray: {
    backgroundColor: '#F5F5F5',
  },
  btnGreen: {
    backgroundColor: '#4CAF50',
  },
  btnRed: {
    backgroundColor: '#F44336',
  },
  btnPressed: {
    opacity: 0.8,
  },
  btnSmallText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#333',
  },
  btnSmallTextWhite: {
    fontSize: 14,
    fontWeight: '500',
    color: '#fff',
  },
});

