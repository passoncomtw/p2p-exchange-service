/**
 * OrdersScreen - 掛單頁面
 */

import React, { useMemo } from 'react';
import { View, Text, StyleSheet, Pressable, StatusBar, Alert, ActivityIndicator } from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '../../store/hooks';
import { fetchPendingOrdersRequest, deletePendingOrderRequest } from '../../store/actions/ordersActions';
import EmptyState from './components/EmptyState';
import OrdersList, { PendingOrder } from './components/OrdersList';
import type { PendingOrder as ApiPendingOrder } from '@/apis/ordersApi';
import logger from '@pkg/logger';

/**
 * 將 API 返回的掛單資料轉換為組件需要的格式
 */
function mapApiOrderToComponentOrder(apiOrder: ApiPendingOrder, type: 'buy' | 'sell'): PendingOrder {
  return {
    id: apiOrder.id,
    type,
    status: apiOrder.status === 0 ? 'active' : 'locked',
    amount: apiOrder.amount,
    totalPrice: apiOrder.balance, // 使用 balance 作為總價
    minAmount: apiOrder.minAmount,
    paymentTimeout: apiOrder.transactionMinutes,
    createdAt: apiOrder.createdAt || new Date().toISOString(),
  };
}

export default function OrdersScreen() {
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const { buy, sell, loading, error, deleting } = useAppSelector((state) => state.orders);

  // 將 buy 和 sell 轉換為組件格式的陣列
  const orders = useMemo(() => {
    const result: PendingOrder[] = [];
    if (buy && buy.id) {
      result.push(mapApiOrderToComponentOrder(buy, 'buy'));
    }
    if (sell && sell.id) {
      result.push(mapApiOrderToComponentOrder(sell, 'sell'));
    }
    return result;
  }, [buy, sell]);

  // 當頁面獲得焦點時，重新取得掛單列表
  useFocusEffect(
    React.useCallback(() => {
      logger.info('OrdersScreen - 頁面聚焦，取得掛單列表');
      dispatch(fetchPendingOrdersRequest());
    }, [dispatch])
  );

  // 檢查是否可以新增掛單（最多一買一賣）
  // 確保檢查 id 屬性，避免空對象被誤認為有掛單
  const canAddBuy = !buy || !buy.id;
  const canAddSell = !sell || !sell.id;
  const canAddOrder = canAddBuy || canAddSell;

  logger.info('OrdersScreen - 掛單狀態', {
    hasBuyId: buy?.id,
    hasSellId: sell?.id,
    canAddBuy,
    canAddSell,
    canAddOrder,
  });

  const handleCreateOrder = () => {
    if (!canAddOrder) {
      Alert.alert('提示', '您已有買入和賣出掛單，無法新增更多');
      return;
    }

    // 顯示選擇買幣或賣幣的對話框
    const options: any[] = [];
    
    if (canAddBuy) {
      options.push({
        text: '買幣',
        onPress: () => {
          // 導航到購買掛單頁面
          (navigation as any).push('CreateOrderBuy', { type: 'buy' });
        },
      });
    }
    
    if (canAddSell) {
      options.push({
        text: '賣幣',
        onPress: () => {
          // 導航到出售掛單頁面
          (navigation as any).push('CreateOrderSell', { type: 'sell' });
        },
      });
    }

    options.push({
      text: '取消',
      style: 'cancel',
    });

    Alert.alert('選擇掛單類型', '請選擇要建立的掛單類型', options);
  };

  const handleDeleteSuccess = () => {
    logger.info('OrdersScreen - 刪除掛單成功');
  };

  const handleDeleteError = (error: string) => {
    logger.warn('OrdersScreen - 刪除掛單失敗', { error });
    Alert.alert('錯誤', error);
  };

  const handleDelete = (orderId: string) => {
    if (deleting) {
      logger.warn('OrdersScreen - 刪除進行中，請勿重複操作');
      return;
    }

    Alert.alert(
      '確認刪除',
      '確定要刪除此掛單嗎？',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '刪除',
          style: 'destructive',
          onPress: () => {
            logger.info('OrdersScreen - 開始刪除掛單', { orderId });
            dispatch(deletePendingOrderRequest({
              orderId,
              onSuccess: handleDeleteSuccess,
              onError: handleDeleteError,
            }));
          },
        },
      ]
    );
  };

  // 判斷是否為空狀態
  const isEmpty = !loading && orders.length === 0;

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />
      
      {/* 頂部導航 */}
      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>掛單</Text>
        <Pressable onPress={handleCreateOrder} style={styles.addButton}>
          <Text style={styles.addButtonText}>➕</Text>
        </Pressable>
      </View>

      {/* 載入中狀態 */}
      {loading && (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#007AFF" />
          <Text style={styles.loadingText}>載入中...</Text>
        </View>
      )}

      {/* 錯誤訊息 */}
      {error && !loading && (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
          <Pressable
            style={styles.retryButton}
            onPress={() => dispatch(fetchPendingOrdersRequest())}
          >
            <Text style={styles.retryButtonText}>重試</Text>
          </Pressable>
        </View>
      )}

      {/* 根據是否有掛單顯示對應組件 */}
      {!loading && !error && (
        isEmpty ? (
        <EmptyState onCreateOrder={handleCreateOrder} />
      ) : (
        <OrdersList
          orders={orders}
          showSuccessAlert={false}
          onDelete={handleDelete}
        />
        )
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  topBar: {
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 16,
    paddingTop: 60,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5E5',
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  topBarTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#000',
  },
  addButton: {
    padding: 8,
  },
  addButtonText: {
    fontSize: 24,
    color: '#007AFF',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    marginTop: 12,
    fontSize: 14,
    color: '#666',
  },
  errorContainer: {
    padding: 16,
    alignItems: 'center',
  },
  errorText: {
    fontSize: 14,
    color: '#F44336',
    marginBottom: 12,
    textAlign: 'center',
  },
  retryButton: {
    paddingHorizontal: 24,
    paddingVertical: 8,
    backgroundColor: '#007AFF',
    borderRadius: 6,
  },
  retryButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
});
