import React, { useState, useCallback } from 'react';
import { View, Text, StyleSheet, Pressable, StatusBar, Alert, ActivityIndicator } from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { useDispatch } from 'react-redux';
import { listingsApi } from '@/apis';
import { pushNotification } from '@/navigation/store/slices/notificationSlice';
import type { ListingItem } from '@/interfaces';
import EmptyState from './components/EmptyState';
import OrdersList, { PendingOrder } from './components/OrdersList';
import logger from '@pkg/logger';

function mapListingToOrder(listing: ListingItem): PendingOrder {
  return {
    id: listing.id.toString(),
    type: listing.type,
    status: listing.status === 'active' ? 'active' : 'locked',
    amount: listing.totalAmount,
    totalPrice: listing.totalAmount * listing.price,
    minAmount: listing.minOrderFiat,
    paymentTimeout: listing.paymentTimeLimit,
    createdAt: listing.createdAt,
  };
}

export default function OrdersScreen() {
  const navigation = useNavigation();
  const dispatch = useDispatch();
  const [listings, setListings] = useState<ListingItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [deleting, setDeleting] = useState(false);

  const fetchMyListings = useCallback(async () => {
    if (loading) return;
    setLoading(true);
    setError(null);
    try {
      const list = await listingsApi.mine();
      setListings(list.filter((l) => l.status !== 'cancelled' && l.status !== 'completed'));
    } catch (err) {
      logger.error('OrdersScreen - 取得掛單失敗', { err });
      setError('無法取得掛單，請稍後再試');
    } finally {
      setLoading(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      fetchMyListings();
    }, [fetchMyListings])
  );

  const hasBuy = listings.some((l) => l.type === 'buy');
  const hasSell = listings.some((l) => l.type === 'sell');
  const canAddBuy = !hasBuy;
  const canAddSell = !hasSell;

  const handleCreateOrder = () => {
    if (!canAddBuy && !canAddSell) {
      Alert.alert('提示', '您已有買入和賣出掛單，無法新增更多');
      return;
    }

    const options: any[] = [];
    if (canAddBuy) {
      options.push({
        text: '買幣',
        onPress: () => (navigation as any).push('CreateOrderBuy', { type: 'buy' }),
      });
    }
    if (canAddSell) {
      options.push({
        text: '賣幣',
        onPress: () => (navigation as any).push('CreateOrderSell', { type: 'sell' }),
      });
    }
    options.push({ text: '取消', style: 'cancel' });
    Alert.alert('選擇掛單類型', '請選擇要建立的掛單類型', options);
  };

  const handleDelete = (orderId: string) => {
    if (deleting) return;
    Alert.alert('確認取消', '確定要取消此掛單嗎？', [
      { text: '返回', style: 'cancel' },
      {
        text: '確認取消',
        style: 'destructive',
        onPress: async () => {
          setDeleting(true);
          try {
            await listingsApi.cancel(Number(orderId));
            await fetchMyListings();
          } catch (err) {
            logger.error('OrdersScreen - 取消掛單失敗', { err });
            dispatch(pushNotification({ type: 'error', message: '取消掛單失敗，請稍後再試' }));
          } finally {
            setDeleting(false);
          }
        },
      },
    ]);
  };

  const orders = listings.map(mapListingToOrder);
  const isEmpty = !loading && orders.length === 0;

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />

      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>掛單</Text>
        <Pressable onPress={handleCreateOrder} style={styles.addButton}>
          <Text style={styles.addButtonText}>+</Text>
        </Pressable>
      </View>

      {loading && (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#007AFF" />
        </View>
      )}

      {error && !loading && (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
          <Pressable style={styles.retryButton} onPress={fetchMyListings}>
            <Text style={styles.retryButtonText}>重試</Text>
          </Pressable>
        </View>
      )}

      {!loading && !error && (
        isEmpty
          ? <EmptyState onCreateOrder={handleCreateOrder} />
          : <OrdersList orders={orders} showSuccessAlert={false} onDelete={handleDelete} />
      )}
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
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  topBarTitle: { fontSize: 20, fontWeight: 'bold', color: '#000' },
  addButton: { padding: 8 },
  addButtonText: { fontSize: 28, color: '#007AFF', lineHeight: 28 },
  loadingContainer: { flex: 1, justifyContent: 'center', alignItems: 'center' },
  errorContainer: { padding: 16, alignItems: 'center' },
  errorText: { fontSize: 14, color: '#F44336', marginBottom: 12, textAlign: 'center' },
  retryButton: { paddingHorizontal: 24, paddingVertical: 8, backgroundColor: '#007AFF', borderRadius: 6 },
  retryButtonText: { color: '#fff', fontSize: 14, fontWeight: '500' },
});
