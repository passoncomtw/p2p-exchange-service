/**
 * OrderListScreen - 訂單頁面
 */

import React, { useState, useMemo, useRef, useEffect } from 'react';
import { View, Text, StyleSheet, FlatList, Pressable, StatusBar, ActivityIndicator } from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '../../store/hooks';
import { fetchOrderListRequest } from '../../store/actions/ordersActions';
import OrderItem from './components/OrderItem';
import logger from '@pkg/logger';

type OrderCategory = 'ongoing' | 'completed';
type OngoingTab = 'pending_payment' | 'pending_release';
type CompletedTab = 'completed' | 'cancelled';

/**
 * 將 API 訂單轉換為 UI 訂單格式
 */

export default function OrderListScreen() {
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const {
    orderList,
    orderListLoading,
    orderListError,
  } = useAppSelector((state) => state.orders);
  const { buyOrders, sellOrders } = useAppSelector((state) => state.market);
  const { user } = useAppSelector((state) => state.auth);

  const [category, setCategory] = useState<OrderCategory>('ongoing');
  const [ongoingTab, setOngoingTab] = useState<OngoingTab>('pending_payment');
  const [completedTab, setCompletedTab] = useState<CompletedTab>('completed');

  // 使用 ref 追蹤是否已經初始化過，避免重複呼叫
  const hasInitialized = useRef(false);
  const isFetching = useRef(false);

  // 當頁面聚焦時，取得訂單列表和掛單列表（僅在首次聚焦時呼叫）
  useFocusEffect(
    React.useCallback(() => {
      // 如果正在載入中或正在取得資料，則不重複呼叫
      if (orderListLoading || isFetching.current) {
        return;
      }

      // 如果已經初始化過，則不重複呼叫
      if (hasInitialized.current) {
        return;
      }

      logger.info('OrderListScreen - 取得訂單列表和掛單列表');
      hasInitialized.current = true;
      isFetching.current = true;
      
      dispatch(fetchOrderListRequest({ size: 100, page: 1 })); // 取得較多筆數以涵蓋所有訂單
    }, [dispatch, orderListLoading])
  );

  // 當載入完成時，重置 isFetching 標記
  useEffect(() => {
    if (!orderListLoading && isFetching.current) {
      isFetching.current = false;
    }
  }, [orderListLoading]);

  const handleOrderPress = (orderId: string) => {
    logger.info('OrderListScreen - 查看訂單詳情', { orderId });
    (navigation as any).navigate('OrderDetail', { orderId });
  };

  // 根據分類和 tab 顯示對應的訂單
  const displayOrders = useMemo(() => {
    if (category === 'ongoing') {
      switch (ongoingTab) {
        case 'pending_payment':
          return orderList.filter((order) => order.status === 0);
        case 'pending_release':
          return orderList.filter((order) => order.status === 1);
        default:
          return [];
      }
    } else {
      switch (completedTab) {
        case 'completed':
          // 狀態 4: 已放行
          return orderList.filter((order) => order.status === 4);
        case 'cancelled':
          // 狀態 2: 買家取消, 狀態 3: 賣家取消
          return orderList.filter((order) => order.status === 2 || order.status === 3);
        default:
          return [];
      }
    }
  }, [category, ongoingTab, completedTab, orderList]);

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />
      
      {/* 頂部導航 */}
      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>訂單</Text>
      </View>

      {/* 分類按鈕：進行中 / 已完成 */}
      <View style={styles.categoryButtons}>
        <Pressable 
          style={[
            styles.categoryBtn,
            category === 'ongoing' && styles.categoryBtnActive
          ]}
          onPress={() => setCategory('ongoing')}
        >
          <Text style={[
            styles.categoryBtnText,
            category === 'ongoing' && styles.categoryBtnTextActive
          ]}>
            進行中
          </Text>
        </Pressable>
        <Pressable 
          style={[
            styles.categoryBtn,
            category === 'completed' && styles.categoryBtnActive
          ]}
          onPress={() => setCategory('completed')}
        >
          <Text style={[
            styles.categoryBtnText,
            category === 'completed' && styles.categoryBtnTextActive
          ]}>
            已完成
          </Text>
        </Pressable>
      </View>

      {/* Tabs：根據分類顯示不同的 tabs */}
      <View style={styles.tabs}>
        {category === 'ongoing' ? (
          <>
            <Pressable 
              style={[styles.tab, ongoingTab === 'pending_payment' && styles.tabActive]}
              onPress={() => setOngoingTab('pending_payment')}
            >
              <Text style={[
                styles.tabText,
                ongoingTab === 'pending_payment' && styles.tabTextActive
              ]}>
                待付款
              </Text>
            </Pressable>
            <Pressable 
              style={[styles.tab, ongoingTab === 'pending_release' && styles.tabActive]}
              onPress={() => setOngoingTab('pending_release')}
            >
              <Text style={[
                styles.tabText,
                ongoingTab === 'pending_release' && styles.tabTextActive
              ]}>
                待放行
              </Text>
            </Pressable>
          </>
        ) : (
          <>
            <Pressable 
              style={[styles.tab, completedTab === 'completed' && styles.tabActive]}
              onPress={() => setCompletedTab('completed')}
            >
              <Text style={[
                styles.tabText,
                completedTab === 'completed' && styles.tabTextActive
              ]}>
                已完成
              </Text>
            </Pressable>
            <Pressable 
              style={[styles.tab, completedTab === 'cancelled' && styles.tabActive]}
              onPress={() => setCompletedTab('cancelled')}
            >
              <Text style={[
                styles.tabText,
                completedTab === 'cancelled' && styles.tabTextActive
              ]}>
                已取消
              </Text>
            </Pressable>
          </>
        )}
      </View>

      {/* 訂單列表 */}
      <FlatList
        data={displayOrders}
        keyExtractor={(item) => item.id}
        renderItem={({ item }) => (
          <OrderItem
            order={item}
            onPress={handleOrderPress}
          />
        )}
        contentContainerStyle={styles.orderList}
        showsVerticalScrollIndicator={false}
        onRefresh={() => dispatch(fetchOrderListRequest({ size: 100, page: 1 }))}
        refreshing={orderListLoading}
        ListEmptyComponent={
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>😔</Text>
              <Text style={styles.emptyTitle}>暫無訂單</Text>
              <Text style={styles.emptySubtitle}>
                {category === 'ongoing'
                  ? `目前沒有${ongoingTab === 'pending_payment' ? '待付款' : '待放行'}的訂單`
                  : `目前沒有${completedTab === 'completed' ? '已完成' : '已取消'}的訂單`}
              </Text>
            </View>
        }
      />
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
  },
  topBarTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#000',
  },
  categoryButtons: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 12,
    gap: 12,
  },
  categoryBtn: {
    flex: 1,
    paddingVertical: 10,
    alignItems: 'center',
    borderRadius: 8,
    backgroundColor: '#F5F5F5',
  },
  categoryBtnActive: {
    backgroundColor: '#007AFF',
  },
  categoryBtnText: {
    fontSize: 15,
    fontWeight: '500',
    color: '#666',
  },
  categoryBtnTextActive: {
    color: '#fff',
  },
  tabs: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5E5',
  },
  tab: {
    flex: 1,
    paddingVertical: 14,
    alignItems: 'center',
    borderBottomWidth: 2,
    borderBottomColor: 'transparent',
  },
  tabActive: {
    borderBottomColor: '#007AFF',
  },
  tabText: {
    fontSize: 14,
    fontWeight: '500',
    color: '#666',
  },
  tabTextActive: {
    color: '#007AFF',
    fontWeight: '600',
  },
  orderList: {
    padding: 12,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 60,
  },
  emptyText: {
    fontSize: 48,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
    marginBottom: 8,
  },
  emptySubtitle: {
    fontSize: 14,
    color: '#666',
    textAlign: 'center',
  },
});

