/**
 * TradeScreen - 交易頁面
 */

import React, { useState, useEffect } from 'react';
import { View, Text, StyleSheet, ScrollView, Pressable, StatusBar, ActivityIndicator } from 'react-native';
import { useRoute, useNavigation, RouteProp, useFocusEffect } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '../../store/hooks';
import { fetchBuyOrdersRequest, fetchSellOrdersRequest } from '../../store/actions/marketActions';
import type { PendingOrder } from '@/apis/ordersApi';
import logger from '@pkg/logger';

type TradeScreenRouteProp = RouteProp<{ Trade: { initialTab?: 'buy' | 'sell' } }, 'Trade'>;

/**
 * 將 API 掛單資料轉換為 UI 顯示格式
 */
interface TransactionUI {
  id: string;
  userId?: number; // 建立掛單的使用者 ID
  userName: string;
  amount: number;
  balance: number;
  price: number;
  minLimit: number;
  maxLimit: number;
  paymentMethod: string;
  successRate: number;
  transactionCount: number;
  bankName?: string;
  // 銀行卡詳細資訊
  bankCardNumber?: string;
  bankCardHolderName?: string;
  bankBranchName?: string;
}

function mapOrderToTransaction(order: PendingOrder): TransactionUI {
  // 計算成交率
  const totalCount = (order.doneCount || 0) + (order.cancelCount || 0);
  const successRate = totalCount > 0 ? Math.round(((order.doneCount || 0) / totalCount) * 100) : 0;

  return {
    id: order.id,
    userId: order.user?.id, // 從 market reducer 的 user.id 取得建立掛單的使用者 ID
    userName: order.user?.name || '匿名用戶',
    amount: order.amount,
    balance: order.balance,
    price: 1.00, // 固定匯率 1:1
    minLimit: order.minAmount,
    maxLimit: order.balance, // 使用剩餘餘額作為最大限額
    paymentMethod: order.bankcard?.bank?.bankName || '銀行卡',
    successRate,
    transactionCount: order.doneCount || 0,
    bankName: order.bankcard?.bank?.bankName,
    // 銀行卡詳細資訊
    bankCardNumber: order.bankcard?.cardNumber,
    bankCardHolderName: order.bankcard?.name,
    bankBranchName: order.bankcard?.branchName,
  };
}

export default function TradeScreen() {
  const route = useRoute<TradeScreenRouteProp>();
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const [activeTab, setActiveTab] = useState<'buy' | 'sell'>('buy');

  // 從 Redux 取得市場資料
  const {
    sellOrders,      // 賣幣掛單（我要買時顯示）
    sellOrdersLoading,
    sellOrdersError,
    buyOrders,       // 買幣掛單（我要賣時顯示）
    buyOrdersLoading,
    buyOrdersError,
  } = useAppSelector((state) => state.market);

  // 接收從其他頁面傳來的 initialTab 參數
  useEffect(() => {
    if (route.params?.initialTab) {
      setActiveTab(route.params.initialTab);
    }
  }, [route.params?.initialTab]);

  // 當頁面聚焦或 tab 切換時，取得對應的掛單列表
  useFocusEffect(
    React.useCallback(() => {
      logger.info('TradeScreen - 取得掛單列表', { activeTab });
      if (activeTab === 'buy') {
        // 我要買 → 取得賣幣掛單 (type=1)
        dispatch(fetchSellOrdersRequest({ size: 20, page: 1 }));
      } else {
        // 我要賣 → 取得買幣掛單 (type=0)
        dispatch(fetchBuyOrdersRequest({ size: 20, page: 1 }));
      }
    }, [dispatch, activeTab])
  );

  // 根據 tab 選擇資料
  const orders = activeTab === 'buy' ? sellOrders : buyOrders;
  const loading = activeTab === 'buy' ? sellOrdersLoading : buyOrdersLoading;
  const error = activeTab === 'buy' ? sellOrdersError : buyOrdersError;

  // 轉換為 UI 格式
  const transactions = orders.map(mapOrderToTransaction);

  const formatNumber = (num: number) => {
    return num.toLocaleString('zh-TW');
  };

  const handleSearch = () => {
    // TODO: 實作搜尋功能
  };

  const handleTransactionPress = (transaction: TransactionUI) => {
    logger.info('TradeScreen - 點擊掛單', {
      orderId: transaction.id,
      activeTab,
      userName: transaction.userName,
      userId: transaction.userId,
    });

    // 導航到統一的確認訂單頁面
    // activeTab 'buy' = 我要買 = 顯示賣幣掛單 = 創建買幣訂單
    // activeTab 'sell' = 我要賣 = 顯示買幣掛單 = 創建賣幣訂單
    (navigation as any).navigate('ConfirmOrder', {
      type: activeTab, // 'buy' | 'sell'
      orderId: transaction.id,
      orderCreatorId: transaction.userId, // 建立掛單的使用者 ID
      userName: transaction.userName,
      availableAmount: transaction.balance, // 使用剩餘數量
      minAmount: transaction.minLimit,
      maxAmount: transaction.maxLimit,
      price: transaction.price,
      bankName: transaction.bankName || transaction.paymentMethod,
      // 銀行卡詳細資訊
      bankCardNumber: transaction.bankCardNumber,
      bankCardHolderName: transaction.bankCardHolderName,
      bankBranchName: transaction.bankBranchName,
    });

    logger.info('TradeScreen - 導航到確認訂單頁面', {
      type: activeTab,
      orderId: transaction.id,
      orderCreatorId: transaction.userId,
      bankCardNumber: transaction.bankCardNumber,
    });
  };

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />
      
      {/* 頂部導航 */}
      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>交易</Text>
        <Pressable onPress={handleSearch} style={styles.searchButton}>
          <Text style={styles.searchIcon}>🔍</Text>
        </Pressable>
      </View>

      {/* Tabs */}
      <View style={styles.tabs}>
        <Pressable 
          style={[styles.tab, activeTab === 'buy' && styles.tabActive]}
          onPress={() => setActiveTab('buy')}
        >
          <Text style={[styles.tabText, activeTab === 'buy' && styles.tabTextActive]}>
            我要買
          </Text>
        </Pressable>
        <Pressable 
          style={[styles.tab, activeTab === 'sell' && styles.tabActive]}
          onPress={() => setActiveTab('sell')}
        >
          <Text style={[styles.tabText, activeTab === 'sell' && styles.tabTextActive]}>
            我要賣
          </Text>
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
            onPress={() => {
              if (activeTab === 'buy') {
                dispatch(fetchSellOrdersRequest({ size: 20, page: 1 }));
              } else {
                dispatch(fetchBuyOrdersRequest({ size: 20, page: 1 }));
              }
            }}
          >
            <Text style={styles.retryButtonText}>重試</Text>
          </Pressable>
        </View>
      )}

      {/* 空狀態 */}
      {!loading && !error && transactions.length === 0 && (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyText}>😔</Text>
          <Text style={styles.emptyTitle}>暫無掛單</Text>
          <Text style={styles.emptySubtitle}>
            {activeTab === 'buy' ? '目前沒有賣幣掛單' : '目前沒有買幣掛單'}
          </Text>
        </View>
      )}

      {/* 交易列表 */}
      {!loading && !error && transactions.length > 0 && (
        <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
          <View style={styles.transactionList}>
            {transactions.map((transaction) => (
              <Pressable
                key={transaction.id}
                style={({ pressed }) => [
                  styles.transactionItem,
                  pressed && styles.transactionItemPressed,
                ]}
                onPress={() => handleTransactionPress(transaction)}
              >
                {/* 頂部：用戶名 + 剩餘數量 */}
                <View style={styles.transactionHeader}>
                  <View style={styles.userInfo}>
                    <Text style={styles.userName}>{transaction.userName}</Text>
                  </View>
                  <Text style={styles.transactionAmount}>
                    剩餘 {formatNumber(transaction.balance)} E幣
                  </Text>
                </View>

                {/* 單價 + 限額 */}
                <View style={styles.transactionInfo}>
                  <Text style={styles.infoText}>單價: ¥{transaction.price.toFixed(2)}</Text>
                  <Text style={styles.infoText}>
                    限額: {formatNumber(transaction.minLimit)}-{formatNumber(transaction.maxLimit)}
                  </Text>
                </View>

                {/* 支付方式 */}
                <View style={styles.transactionInfo}>
                  <Text style={styles.infoText}>💳 {transaction.paymentMethod}</Text>
                </View>

                {/* 成交統計 */}
                {transaction.transactionCount > 0 && (
                  <View style={[styles.transactionInfo, styles.transactionFooter]}>
                    <Text style={styles.statsText}>
                      成交率: {transaction.successRate}% | 成交次數: {transaction.transactionCount}
                    </Text>
                  </View>
                )}
              </Pressable>
            ))}
          </View>
        </ScrollView>
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
  searchButton: {
    padding: 8,
  },
  searchIcon: {
    fontSize: 20,
  },
  tabs: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5E5',
  },
  tab: {
    flex: 1,
    paddingVertical: 16,
    alignItems: 'center',
    borderBottomWidth: 2,
    borderBottomColor: 'transparent',
  },
  tabActive: {
    borderBottomColor: '#007AFF',
  },
  tabText: {
    fontSize: 16,
    fontWeight: '500',
    color: '#666',
  },
  tabTextActive: {
    color: '#007AFF',
    fontWeight: '600',
  },
  content: {
    flex: 1,
  },
  transactionList: {
    padding: 16,
  },
  transactionItem: {
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
  transactionItemPressed: {
    backgroundColor: '#F8F8F8',
  },
  transactionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  userInfo: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  userName: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
    marginRight: 8,
  },
  statusOnline: {
    backgroundColor: '#4CAF50',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  statusOffline: {
    backgroundColor: '#999',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  statusText: {
    fontSize: 11,
    color: '#fff',
    fontWeight: '500',
  },
  transactionAmount: {
    fontSize: 16,
    fontWeight: 'bold',
    color: '#333',
  },
  transactionInfo: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  transactionFooter: {
    marginTop: 4,
    marginBottom: 0,
  },
  infoText: {
    fontSize: 14,
    color: '#666',
  },
  verifiedText: {
    fontSize: 14,
    color: '#4CAF50',
    fontWeight: '500',
  },
  statsText: {
    fontSize: 12,
    color: '#666',
  },
  badgeFast: {
    fontSize: 12,
    color: '#FF9800',
    fontWeight: '500',
  },
  badgeTrusted: {
    fontSize: 12,
    color: '#2196F3',
    fontWeight: '500',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 40,
  },
  loadingText: {
    marginTop: 12,
    fontSize: 14,
    color: '#666',
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
  },
  errorText: {
    fontSize: 14,
    color: '#F44336',
    marginBottom: 16,
    textAlign: 'center',
  },
  retryButton: {
    paddingHorizontal: 24,
    paddingVertical: 10,
    backgroundColor: '#007AFF',
    borderRadius: 6,
  },
  retryButtonText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '500',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 32,
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
