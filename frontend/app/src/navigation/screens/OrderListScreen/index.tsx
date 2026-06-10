import React, { useState, useMemo, useCallback } from 'react';
import { View, Text, StyleSheet, FlatList, Pressable, StatusBar } from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { p2pOrdersApi } from '@/apis';
import type { Order } from '@/interfaces';
import OrderItem from './components/OrderItem';
import logger from '@pkg/logger';

type OrderCategory = 'ongoing' | 'completed';
type OngoingTab = 'pending_payment' | 'pending_release';
type CompletedTab = 'completed' | 'cancelled';

const ONGOING_STATUSES = new Set(['matched', 'paid', 'releasing', 'disputed']);
const PENDING_PAYMENT = new Set(['matched']);
const PENDING_RELEASE = new Set(['paid', 'releasing']);
const COMPLETED = new Set(['completed']);
const CANCELLED = new Set(['cancelled', 'timeout']);

export default function OrderListScreen() {
  const navigation = useNavigation();
  const [orders, setOrders] = useState<Order[]>([]);
  const [loading, setLoading] = useState(false);

  const [category, setCategory] = useState<OrderCategory>('ongoing');
  const [ongoingTab, setOngoingTab] = useState<OngoingTab>('pending_payment');
  const [completedTab, setCompletedTab] = useState<CompletedTab>('completed');

  const fetchOrders = useCallback(async () => {
    if (loading) return;
    setLoading(true);
    try {
      const list = await p2pOrdersApi.list({ limit: 100 });
      setOrders(list);
    } catch (err) {
      logger.error('OrderListScreen - 取得訂單失敗', { err });
    } finally {
      setLoading(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      fetchOrders();
    }, [fetchOrders])
  );

  const handleOrderPress = (orderId: string) => {
    (navigation as any).navigate('OrderDetail', { orderId });
  };

  const displayOrders = useMemo(() => {
    if (category === 'ongoing') {
      const statuses = ongoingTab === 'pending_payment' ? PENDING_PAYMENT : PENDING_RELEASE;
      return orders.filter((o) => statuses.has(o.status));
    }
    const statuses = completedTab === 'completed' ? COMPLETED : CANCELLED;
    return orders.filter((o) => statuses.has(o.status));
  }, [category, ongoingTab, completedTab, orders]);

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />

      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>訂單</Text>
      </View>

      <View style={styles.categoryButtons}>
        {(['ongoing', 'completed'] as OrderCategory[]).map((cat) => (
          <Pressable
            key={cat}
            style={[styles.categoryBtn, category === cat && styles.categoryBtnActive]}
            onPress={() => setCategory(cat)}
          >
            <Text style={[styles.categoryBtnText, category === cat && styles.categoryBtnTextActive]}>
              {cat === 'ongoing' ? '進行中' : '已完成'}
            </Text>
          </Pressable>
        ))}
      </View>

      <View style={styles.tabs}>
        {category === 'ongoing' ? (
          <>
            {(['pending_payment', 'pending_release'] as OngoingTab[]).map((tab) => (
              <Pressable
                key={tab}
                style={[styles.tab, ongoingTab === tab && styles.tabActive]}
                onPress={() => setOngoingTab(tab)}
              >
                <Text style={[styles.tabText, ongoingTab === tab && styles.tabTextActive]}>
                  {tab === 'pending_payment' ? '待付款' : '待放行'}
                </Text>
              </Pressable>
            ))}
          </>
        ) : (
          <>
            {(['completed', 'cancelled'] as CompletedTab[]).map((tab) => (
              <Pressable
                key={tab}
                style={[styles.tab, completedTab === tab && styles.tabActive]}
                onPress={() => setCompletedTab(tab)}
              >
                <Text style={[styles.tabText, completedTab === tab && styles.tabTextActive]}>
                  {tab === 'completed' ? '已完成' : '已取消'}
                </Text>
              </Pressable>
            ))}
          </>
        )}
      </View>

      <FlatList
        data={displayOrders}
        keyExtractor={(item) => item.id.toString()}
        renderItem={({ item }) => <OrderItem order={item} onPress={handleOrderPress} />}
        contentContainerStyle={styles.orderList}
        showsVerticalScrollIndicator={false}
        onRefresh={fetchOrders}
        refreshing={loading}
        ListEmptyComponent={
          <View style={styles.emptyContainer}>
            <Text style={styles.emptyText}>暫無訂單</Text>
          </View>
        }
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: '#F5F5F5' },
  topBar: {
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 16,
    paddingTop: 60,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5E5',
  },
  topBarTitle: { fontSize: 20, fontWeight: 'bold', color: '#000' },
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
  categoryBtnActive: { backgroundColor: '#007AFF' },
  categoryBtnText: { fontSize: 15, fontWeight: '500', color: '#666' },
  categoryBtnTextActive: { color: '#fff' },
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
  tabActive: { borderBottomColor: '#007AFF' },
  tabText: { fontSize: 14, fontWeight: '500', color: '#666' },
  tabTextActive: { color: '#007AFF', fontWeight: '600' },
  orderList: { padding: 12 },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 60,
  },
  emptyText: { fontSize: 16, color: '#999' },
});
