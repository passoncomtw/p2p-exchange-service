import React, { useState, useCallback } from 'react';
import {
  View, Text, StyleSheet, ScrollView, Pressable,
  StatusBar, ActivityIndicator,
} from 'react-native';
import { useRoute, useNavigation, RouteProp, useFocusEffect } from '@react-navigation/native';
import { listingsApi } from '@/apis';
import type { ListingItem } from '@/interfaces';
import logger from '@pkg/logger';

type TradeScreenRouteProp = RouteProp<{ Trade: { initialTab?: 'buy' | 'sell' } }, 'Trade'>;

const fmt = (n: number) => n.toLocaleString('zh-TW');

export default function TradeScreen() {
  const route = useRoute<TradeScreenRouteProp>();
  const navigation = useNavigation();
  const [activeTab, setActiveTab] = useState<'buy' | 'sell'>(
    route.params?.initialTab ?? 'buy'
  );
  const [listings, setListings] = useState<ListingItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchListings = useCallback(async (tab: 'buy' | 'sell') => {
    setLoading(true);
    setError(null);
    try {
      // 我要買 → 看別人的賣單 (type='sell')
      // 我要賣 → 看別人的買單 (type='buy')
      const counterType = tab === 'buy' ? 'sell' : 'buy';
      const list = await listingsApi.list({ type: counterType, status: 'active', limit: 50 });
      setListings(list);
    } catch (err) {
      logger.error('TradeScreen - 取得掛單失敗', { err });
      setError('無法取得交易列表，請稍後再試');
    } finally {
      setLoading(false);
    }
  }, []);

  useFocusEffect(
    useCallback(() => {
      fetchListings(activeTab);
    }, [fetchListings, activeTab])
  );

  const handleTabChange = (tab: 'buy' | 'sell') => {
    setActiveTab(tab);
    // useFocusEffect 不會因 state 變更自動重新觸發，手動呼叫
    fetchListings(tab);
  };

  const handleListingPress = (listing: ListingItem) => {
    logger.info('TradeScreen - 點擊掛單', { listingId: listing.id, tab: activeTab });
    (navigation as any).navigate('ConfirmOrder', {
      type: activeTab,
      orderId: listing.id.toString(),
      orderCreatorId: listing.userId,
      userName: `用戶 #${listing.userId}`,
      availableAmount: listing.remainingAmount,
      minAmount: listing.minOrderFiat,
      maxAmount: listing.maxOrderFiat,
      price: listing.price,
      bankName: listing.paymentMethodId ? `付款方式 #${listing.paymentMethodId}` : '未指定',
    });
  };

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />

      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>交易</Text>
      </View>

      <View style={styles.tabs}>
        {(['buy', 'sell'] as const).map((tab) => (
          <Pressable
            key={tab}
            style={[styles.tab, activeTab === tab && styles.tabActive]}
            onPress={() => handleTabChange(tab)}
          >
            <Text style={[styles.tabText, activeTab === tab && styles.tabTextActive]}>
              {tab === 'buy' ? '我要買' : '我要賣'}
            </Text>
          </Pressable>
        ))}
      </View>

      {loading && (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#007AFF" />
          <Text style={styles.loadingText}>載入中...</Text>
        </View>
      )}

      {error && !loading && (
        <View style={styles.errorContainer}>
          <Text style={styles.errorText}>{error}</Text>
          <Pressable style={styles.retryButton} onPress={() => fetchListings(activeTab)}>
            <Text style={styles.retryButtonText}>重試</Text>
          </Pressable>
        </View>
      )}

      {!loading && !error && listings.length === 0 && (
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyTitle}>暫無掛單</Text>
          <Text style={styles.emptySubtitle}>
            {activeTab === 'buy' ? '目前沒有賣幣掛單' : '目前沒有買幣掛單'}
          </Text>
        </View>
      )}

      {!loading && !error && listings.length > 0 && (
        <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
          <View style={styles.list}>
            {listings.map((listing) => (
              <Pressable
                key={listing.id}
                style={({ pressed }) => [styles.card, pressed && styles.cardPressed]}
                onPress={() => handleListingPress(listing)}
              >
                <View style={styles.cardHeader}>
                  <Text style={styles.userName}>用戶 #{listing.userId}</Text>
                  <Text style={styles.remaining}>剩餘 {fmt(listing.remainingAmount)} {listing.cryptoCurrency}</Text>
                </View>
                <View style={styles.cardRow}>
                  <Text style={styles.infoText}>單價: {listing.fiatCurrency} {fmt(listing.price)}</Text>
                  <Text style={styles.infoText}>
                    限額: {fmt(listing.minOrderFiat)}–{fmt(listing.maxOrderFiat)} {listing.fiatCurrency}
                  </Text>
                </View>
                <View style={styles.cardRow}>
                  <Text style={styles.infoText}>支付時限: {listing.paymentTimeLimit} 分鐘</Text>
                </View>
              </Pressable>
            ))}
          </View>
        </ScrollView>
      )}
    </View>
  );
}

const styles = StyleSheet.create({
  container:  { flex: 1, backgroundColor: '#F5F5F5' },
  topBar: {
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 16,
    paddingTop: 60,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5E5',
  },
  topBarTitle: { fontSize: 20, fontWeight: 'bold', color: '#000' },
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
  tabActive: { borderBottomColor: '#007AFF' },
  tabText:       { fontSize: 16, fontWeight: '500', color: '#666' },
  tabTextActive: { color: '#007AFF', fontWeight: '600' },
  content: { flex: 1 },
  list:    { padding: 16 },
  card: {
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
  cardPressed:  { backgroundColor: '#F8F8F8' },
  cardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  userName:  { fontSize: 16, fontWeight: '600', color: '#333' },
  remaining: { fontSize: 16, fontWeight: 'bold', color: '#333' },
  cardRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginBottom: 8,
  },
  infoText: { fontSize: 14, color: '#666' },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingVertical: 40,
  },
  loadingText:  { marginTop: 12, fontSize: 14, color: '#666' },
  errorContainer: { flex: 1, justifyContent: 'center', alignItems: 'center', padding: 32 },
  errorText:   { fontSize: 14, color: '#F44336', marginBottom: 16, textAlign: 'center' },
  retryButton: { paddingHorizontal: 24, paddingVertical: 10, backgroundColor: '#007AFF', borderRadius: 6 },
  retryButtonText: { color: '#fff', fontSize: 14, fontWeight: '500' },
  emptyContainer: { flex: 1, justifyContent: 'center', alignItems: 'center', padding: 32 },
  emptyTitle:    { fontSize: 18, fontWeight: '600', color: '#333', marginBottom: 8 },
  emptySubtitle: { fontSize: 14, color: '#666', textAlign: 'center' },
});
