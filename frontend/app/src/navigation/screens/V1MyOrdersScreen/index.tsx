import * as React from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  Alert,
  ScrollView,
  RefreshControl,
} from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import * as tokens from '@/theme';
import { listingsApi } from '@/apis/listingsApi';
import { p2pOrdersApi } from '@/apis/p2pOrdersApi';
import { useAppSelector } from '@/navigation/store/hooks';
import type { ListingItem } from '@/interfaces/listing';
import type { Order } from '@/interfaces/order';

const { colors, orderStatusColors } = tokens;
const formatDateTime = (iso: string) => new Date(iso).toLocaleString();

export default function V1MyOrdersScreen() {
  const { t } = useTranslation();
  const navigation = useNavigation<any>();
  const currentUserId = useAppSelector((s) => s.auth.user?.id);
  const [listings, setListings] = React.useState<ListingItem[]>([]);
  const [ordersByListing, setOrdersByListing] = React.useState<Record<number, Order[]>>({});
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);

  const load = React.useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      const [listingList, orderList] = await Promise.all([
        listingsApi.mine(),
        p2pOrdersApi.list(),
      ]);
      setListings(listingList);
      const grouped: Record<number, Order[]> = {};
      for (const order of orderList) {
        if (!grouped[order.listingId]) grouped[order.listingId] = [];
        grouped[order.listingId].push(order);
      }
      setOrdersByListing(grouped);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, []);

  useFocusEffect(
    React.useCallback(() => {
      load();
    }, [load]),
  );

  const confirmCancelListing = (listing: ListingItem) => {
    Alert.alert('', t('order.message.cancelConfirm'), [
      { text: t('order.action.dismiss'), style: 'cancel' },
      {
        text: t('order.action.confirm'),
        style: 'destructive',
        onPress: async () => {
          try {
            await listingsApi.cancel(listing.id);
            await load();
          } catch {
            Alert.alert('', t('order.message.submitFailed'));
          }
        },
      },
    ]);
  };

  const Title = () => <Text style={styles.title}>{t('order.pageTitle.myOrders')}</Text>;

  if (loading) {
    return (
      <ScrollView style={styles.screen} contentContainerStyle={styles.pad}>
        <Title />
        <View style={{ gap: 12 }}>
          {[0.6, 0.4, 0.25].map((o, i) => (
            <View key={i} style={[styles.skeleton, { opacity: o }]} />
          ))}
        </View>
      </ScrollView>
    );
  }

  if (error) {
    return (
      <ScrollView style={styles.screen} contentContainerStyle={styles.pad}>
        <Title />
        <View style={styles.centerBox}>
          <Text style={styles.errorMark}>!</Text>
          <Text style={styles.errorTitle}>{t('order.message.errorTitle')}</Text>
          <Text style={styles.errorBody}>{t('order.message.errorBody')}</Text>
          <TouchableOpacity style={styles.outlineBtn} onPress={load} accessibilityRole="button">
            <Text style={styles.outlineBtnText}>{t('order.action.retry')}</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    );
  }

  if (listings.length === 0) {
    return (
      <ScrollView style={styles.screen} contentContainerStyle={styles.pad}>
        <Title />
        <View style={styles.centerBox}>
          <View style={styles.emptyIcon}><Text style={styles.emptyIconText}>+</Text></View>
          <Text style={styles.emptyText}>{t('order.message.emptyMine')}</Text>
          <TouchableOpacity
            style={styles.primaryBtn}
            onPress={() => navigation.navigate('CreateOrder')}
            accessibilityRole="button"
          >
            <Text style={styles.primaryBtnText}>{t('order.action.goCreate')}</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    );
  }

  const renderItem = ({ item }: { item: ListingItem }) => {
    const tColor = tokens.typeColor(item.type);
    const fiatTotal = item.price * item.totalAmount;
    const relatedOrders = ordersByListing[item.id] ?? [];

    const allOrdersCompleted = relatedOrders.length > 0 && relatedOrders.every((o) => o.status === 'completed');
    const displayStatus = allOrdersCompleted ? 'completed' : item.status;
    const sColor = tokens.statusColor(displayStatus);
    const canCancelListing = item.status === 'active' && !allOrdersCompleted;

    return (
      <View style={styles.card}>
        <View style={styles.cardHeader}>
          <View style={[styles.typeTag, { backgroundColor: tColor }]}>
            <Text style={styles.typeTagText}>{t(`order.type.${item.type}`)}</Text>
          </View>
          <View style={[styles.statusTag, { borderColor: sColor }]}>
            <View style={[styles.dot, { backgroundColor: sColor }]} />
            <Text style={[styles.statusTagText, { color: sColor }]}>{t(`order.status.${displayStatus}`)}</Text>
          </View>
        </View>

        <Row label={t('order.field.price')} value={`${item.price.toLocaleString()} ${item.fiatCurrency}`} />
        <Row label={t('order.field.quantity')} value={`${item.totalAmount} ${item.cryptoCurrency}`} />
        <Row label={t('order.field.remainingAmount')} value={`${item.remainingAmount} ${item.cryptoCurrency}`} />
        <Row label={t('order.field.totalAmount')} value={`${fiatTotal.toLocaleString()} ${item.fiatCurrency}`} amber />
        <Row label={t('order.field.createdAt')} value={formatDateTime(item.createdAt)} muted />

        {canCancelListing && (
          <TouchableOpacity style={styles.cancelBtn} onPress={() => confirmCancelListing(item)} accessibilityRole="button">
            <Text style={styles.cancelBtnText}>{t('order.action.cancel')}</Text>
          </TouchableOpacity>
        )}

        {relatedOrders.length > 0 && (
          <View style={styles.ordersSection}>
            <Text style={styles.ordersSectionTitle}>{t('order.nav.orders')} ({relatedOrders.length})</Text>
            {relatedOrders.map((order) => {
              const oColor = orderStatusColors[order.status] ?? colors.statusCancelled;
              const isBuyer = order.buyerId === currentUserId;
              const roleKey = isBuyer ? 'asBuyer' : 'asSeller';

              return (
                <TouchableOpacity
                  key={order.id}
                  style={styles.orderCard}
                  onPress={() => navigation.navigate('OrderDetail', { id: order.id })}
                  accessibilityRole="button"
                >
                  <View style={styles.orderCardHeader}>
                    <Text style={styles.orderNo}>{order.orderNo}</Text>
                    <View style={[styles.orderStatusTag, { borderColor: oColor }]}>
                      <View style={[styles.dot, { backgroundColor: oColor }]} />
                      <Text style={[styles.orderStatusText, { color: oColor }]}>{t(`order.orderStatus.${order.status}`)}</Text>
                    </View>
                  </View>
                  <View style={styles.orderMeta}>
                    <Text style={styles.orderRole}>{t(`order.role.${roleKey}`)}</Text>
                    <Text style={styles.orderAmount}>
                      {order.cryptoAmount} {order.cryptoCurrency} / {order.totalAmount.toLocaleString()} {order.fiatCurrency}
                    </Text>
                  </View>
                </TouchableOpacity>
              );
            })}
          </View>
        )}
      </View>
    );
  };

  return (
    <FlatList
      style={styles.screen}
      data={listings}
      keyExtractor={(o) => String(o.id)}
      renderItem={renderItem}
      ListHeaderComponent={<Title />}
      contentContainerStyle={styles.pad}
      ItemSeparatorComponent={() => <View style={{ height: 12 }} />}
      refreshControl={<RefreshControl refreshing={false} onRefresh={load} />}
    />
  );
}

function Row({ label, value, amber, muted }: { label: string; value: string; amber?: boolean; muted?: boolean }) {
  return (
    <View style={styles.row}>
      <Text style={styles.rowLabel}>{label}</Text>
      <Text style={[styles.rowValue, amber && styles.rowAmber, muted && styles.rowMuted]}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bgContent },
  pad: { padding: 16 },
  title: { fontSize: 16, fontWeight: '600', color: colors.textPrimary, marginBottom: 14 },

  skeleton: { height: 96, backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard },

  centerBox: { alignItems: 'center', paddingVertical: 56, paddingHorizontal: 20 },
  errorMark: { fontSize: 32, color: colors.danger, marginBottom: 8, fontWeight: '700' },
  errorTitle: { fontSize: 14, color: colors.textPrimary, fontWeight: '500', marginBottom: 4 },
  errorBody: { fontSize: 12, color: colors.textTertiary, marginBottom: 16, textAlign: 'center' },
  outlineBtn: { height: 36, paddingHorizontal: 20, borderWidth: 1, borderColor: colors.borderInput, borderRadius: 4, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.bgCard },
  outlineBtnText: { fontSize: 14, fontWeight: '500', color: colors.textPrimary },

  emptyIcon: { width: 56, height: 56, borderRadius: 28, backgroundColor: '#EFEFF2', alignItems: 'center', justifyContent: 'center', marginBottom: 14 },
  emptyIconText: { fontSize: 24, color: '#BDBDBD' },
  emptyText: { fontSize: 14, color: colors.textSecondary, marginBottom: 16 },
  primaryBtn: { height: 36, paddingHorizontal: 20, backgroundColor: colors.primary, borderRadius: 4, alignItems: 'center', justifyContent: 'center' },
  primaryBtnText: { fontSize: 14, fontWeight: '500', color: '#1F2327' },

  card: { backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard, padding: 14 },
  cardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 },
  typeTag: { paddingHorizontal: 8, paddingVertical: 2, borderRadius: 4 },
  typeTagText: { fontSize: 12, fontWeight: '600', color: '#fff' },
  statusTag: { flexDirection: 'row', alignItems: 'center', gap: 6, paddingHorizontal: 8, paddingVertical: 2, borderRadius: 4, borderWidth: 1 },
  dot: { width: 6, height: 6, borderRadius: 3 },
  statusTagText: { fontSize: 12, fontWeight: '500' },
  row: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 3 },
  rowLabel: { fontSize: 12, color: colors.textSecondary },
  rowValue: { fontSize: 12, color: colors.textPrimary, fontWeight: '500' },
  rowAmber: { color: colors.amberText, fontWeight: '700' },
  rowMuted: { color: colors.textTertiary, fontWeight: '400' },
  cancelBtn: { marginTop: 12, height: 36, borderRadius: 4, borderWidth: 1, borderColor: colors.danger, alignItems: 'center', justifyContent: 'center' },
  cancelBtnText: { fontSize: 14, color: colors.danger, fontWeight: '500' },

  ordersSection: { marginTop: 14, borderTopWidth: 1, borderTopColor: colors.borderCard, paddingTop: 12 },
  ordersSectionTitle: { fontSize: 13, fontWeight: '600', color: colors.textPrimary, marginBottom: 8 },
  orderCard: { backgroundColor: colors.bgContent, borderRadius: 6, borderWidth: 1, borderColor: colors.borderCard, padding: 10, marginBottom: 8 },
  orderCardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 },
  orderNo: { fontSize: 12, fontWeight: '600', color: colors.textPrimary },
  orderStatusTag: { flexDirection: 'row', alignItems: 'center', gap: 4, paddingHorizontal: 6, paddingVertical: 1, borderRadius: 4, borderWidth: 1 },
  orderStatusText: { fontSize: 11, fontWeight: '500' },
  orderMeta: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center' },
  orderRole: { fontSize: 11, color: colors.textTertiary },
  orderAmount: { fontSize: 11, color: colors.textSecondary },
});
