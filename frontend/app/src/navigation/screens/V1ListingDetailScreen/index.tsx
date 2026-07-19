import * as React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  ActivityIndicator,
  Alert,
  TextInput,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useNavigation, useRoute } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import * as tokens from '@/theme';
import { listingsApi } from '@/apis/listingsApi';
import { p2pOrdersApi } from '@/apis/p2pOrdersApi';
import { paymentMethodsApi } from '@/apis/paymentMethodsApi';
import { useAppDispatch, useAppSelector } from '@/navigation/store/hooks';
import { pushNotification } from '@/navigation/store/slices/notificationSlice';
import SkeletonList from '@/components/SkeletonList';
import type { ListingItem } from '@/interfaces/listing';
import type { Order } from '@/interfaces/order';

const { colors, orderStatusColors } = tokens;
const formatDateTime = (iso: string) => new Date(iso).toLocaleString();

export default function V1ListingDetailScreen() {
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const navigation = useNavigation<any>();
  const route = useRoute<any>();
  const listingId: number = route.params?.id;
  const currentUserId = useAppSelector((s) => s.auth.user?.id);

  const [listing, setListing] = React.useState<ListingItem | null>(null);
  const [relatedOrders, setRelatedOrders] = React.useState<Order[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);
  const [taking, setTaking] = React.useState(false);
  const [tradeAmount, setTradeAmount] = React.useState('');

  const load = React.useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      const [data, orders] = await Promise.all([
        listingsApi.getById(listingId),
        p2pOrdersApi.list(),
      ]);
      setListing(data);
      setTradeAmount(String(data.remainingAmount));
      setRelatedOrders(orders.filter((o) => o.listingId === listingId));
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [listingId]);

  React.useEffect(() => { load(); }, [load]);

  const handleTakeOrder = async () => {
    if (!listing) return;

    const amount = parseFloat(tradeAmount);
    if (isNaN(amount) || amount <= 0 || amount > listing.remainingAmount) {
      Alert.alert('', t('order.message.invalidAmount'));
      return;
    }

    if (listing.type === 'buy') {
      try {
        const methods = await paymentMethodsApi.list();
        if (!methods || methods.length === 0) {
          Alert.alert('', t('order.message.noPaymentMethod'), [
            { text: t('order.action.dismiss'), style: 'cancel' },
            {
              text: t('order.action.addPayment'),
              onPress: () => navigation.navigate('AddPaymentMethod'),
            },
          ]);
          return;
        }
      } catch {
        dispatch(pushNotification({ type: 'error', message: t('order.message.submitFailed') }));
        return;
      }
    }

    setTaking(true);
    try {
      const result = await p2pOrdersApi.create({
        listingId: listing.id,
        cryptoAmount: amount,
      });
      dispatch(pushNotification({ type: 'success', message: t('order.message.takeOrderSuccess') }));
      navigation.replace('OrderDetail', { id: result.id });
    } catch {
      dispatch(pushNotification({ type: 'error', message: t('order.message.takeOrderFailed') }));
    } finally {
      setTaking(false);
    }
  };

  if (loading) {
    return (
      <ScrollView style={styles.screen} contentContainerStyle={styles.container}>
        <SkeletonList height={48} />
      </ScrollView>
    );
  }

  if (error || !listing) {
    return (
      <ScrollView style={styles.screen} contentContainerStyle={styles.container}>
        <View style={styles.centerBox}>
          <Text style={styles.errorMark}>!</Text>
          <Text style={styles.errorTitle}>{t('order.message.errorTitle')}</Text>
          <TouchableOpacity style={styles.outlineBtn} onPress={load} accessibilityRole="button">
            <Text style={styles.outlineBtnText}>{t('order.action.retry')}</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    );
  }

  const isActive = listing.status === 'active' && listing.remainingAmount > 0;
  const parsedAmount = parseFloat(tradeAmount) || 0;
  const fiatEstimate = listing.price * parsedAmount;
  const tColor = tokens.typeColor(listing.type);
  const actionLabel = listing.type === 'sell' ? t('order.action.takeBuy') : t('order.action.takeSell');

  return (
    <KeyboardAvoidingView style={{ flex: 1 }} behavior={Platform.OS === 'ios' ? 'padding' : undefined}>
    <ScrollView style={styles.screen} contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
      <View style={styles.headerRow}>
        <View style={[styles.typeTag, { backgroundColor: tColor }]}>
          <Text style={styles.typeTagText}>{t(`order.type.${listing.type}`)}</Text>
        </View>
        <Text style={styles.cryptoLabel}>{listing.cryptoCurrency}/{listing.fiatCurrency}</Text>
      </View>

      <View style={styles.section}>
        <Row label={t('order.field.price')} value={`${listing.price.toLocaleString()} ${listing.fiatCurrency}`} />
        <Row label={t('order.field.quantity')} value={`${listing.totalAmount} ${listing.cryptoCurrency}`} />
        <Row label={t('order.field.remainingAmount')} value={`${listing.remainingAmount} ${listing.cryptoCurrency}`} />
        <Row label={t('order.field.createdAt')} value={formatDateTime(listing.createdAt)} muted />
      </View>

      {isActive ? (
        <>
          <View style={styles.amountSection}>
            <Text style={styles.amountLabel}>{t('order.field.tradeAmount')} ({listing.cryptoCurrency})</Text>
            <View style={styles.amountInputRow}>
              <TextInput
                style={styles.amountInput}
                value={tradeAmount}
                onChangeText={setTradeAmount}
                keyboardType="decimal-pad"
                placeholder="0"
                placeholderTextColor={colors.textTertiary}
              />
              <TouchableOpacity
                style={styles.maxBtn}
                onPress={() => setTradeAmount(String(listing.remainingAmount))}
                accessibilityRole="button"
              >
                <Text style={styles.maxBtnText}>{t('order.action.max')}</Text>
              </TouchableOpacity>
            </View>
            <Text style={styles.fiatEstimate}>
              {t('order.field.fiatEstimate')}: {fiatEstimate.toLocaleString()} {listing.fiatCurrency}
            </Text>
          </View>
          <TouchableOpacity
            style={[styles.actionBtn, { backgroundColor: tColor }, taking && styles.actionBtnDisabled]}
            onPress={handleTakeOrder}
            disabled={taking}
            accessibilityRole="button"
          >
            {taking ? <ActivityIndicator size="small" color="#fff" style={{ marginRight: 8 }} /> : null}
            <Text style={styles.actionBtnText}>{actionLabel}</Text>
          </TouchableOpacity>
        </>
      ) : relatedOrders.length > 0 ? (
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
                <Text style={styles.viewDetail}>{t('order.action.viewDetail')} &gt;</Text>
              </TouchableOpacity>
            );
          })}
        </View>
      ) : (
        <View style={styles.unavailableBox}>
          <Text style={styles.unavailableText}>{t('order.message.listingUnavailable')}</Text>
        </View>
      )}
    </ScrollView>
    </KeyboardAvoidingView>
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
  container: { padding: 16, paddingBottom: 28 },
  centerBox: { alignItems: 'center', paddingVertical: 56, paddingHorizontal: 20 },
  errorMark: { fontSize: 32, color: colors.danger, marginBottom: 8, fontWeight: '700' },
  errorTitle: { fontSize: 14, color: colors.textPrimary, fontWeight: '500', marginBottom: 12 },
  outlineBtn: { height: 36, paddingHorizontal: 20, borderWidth: 1, borderColor: colors.borderInput, borderRadius: 4, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.bgCard },
  outlineBtnText: { fontSize: 14, fontWeight: '500', color: colors.textPrimary },
  headerRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 },
  typeTag: { paddingHorizontal: 10, paddingVertical: 4, borderRadius: 4 },
  typeTagText: { fontSize: 13, fontWeight: '600', color: '#fff' },
  cryptoLabel: { fontSize: 13, color: colors.textTertiary, fontWeight: '500' },
  section: { backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard, padding: 14, marginBottom: 20 },
  row: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 5 },
  rowLabel: { fontSize: 13, color: colors.textSecondary },
  rowValue: { fontSize: 13, color: colors.textPrimary, fontWeight: '500' },
  rowAmber: { color: colors.amberText, fontWeight: '700' },
  rowMuted: { color: colors.textTertiary, fontWeight: '400' },
  unavailableBox: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.borderCard,
    padding: 16,
    alignItems: 'center',
  },
  unavailableText: { fontSize: 14, color: colors.textTertiary },
  amountSection: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.borderCard,
    padding: 14,
    marginBottom: 16,
  },
  amountLabel: { fontSize: 13, color: colors.textSecondary, marginBottom: 8 },
  amountInputRow: { flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 6 },
  amountInput: {
    flex: 1,
    height: 44,
    borderWidth: 1,
    borderColor: colors.borderInput,
    borderRadius: 6,
    paddingHorizontal: 12,
    fontSize: 16,
    color: colors.textPrimary,
    backgroundColor: colors.bgContent,
  },
  maxBtn: {
    height: 44,
    paddingHorizontal: 14,
    borderWidth: 1,
    borderColor: colors.borderInput,
    borderRadius: 6,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.bgCard,
  },
  maxBtnText: { fontSize: 12, fontWeight: '600', color: colors.textSecondary },
  fiatEstimate: { fontSize: 12, color: colors.textTertiary },
  actionBtn: {
    height: 48,
    borderRadius: 4,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 8,
  },
  actionBtnDisabled: { opacity: 0.6 },
  actionBtnText: { fontSize: 16, fontWeight: '600', color: '#fff' },
  ordersSection: { gap: 8 },
  ordersSectionTitle: { fontSize: 14, fontWeight: '600', color: colors.textPrimary, marginBottom: 4 },
  orderCard: { backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard, padding: 12 },
  orderCardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 },
  orderNo: { fontSize: 13, fontWeight: '600', color: colors.textPrimary },
  orderStatusTag: { flexDirection: 'row', alignItems: 'center', gap: 4, paddingHorizontal: 6, paddingVertical: 2, borderRadius: 4, borderWidth: 1 },
  dot: { width: 6, height: 6, borderRadius: 3 },
  orderStatusText: { fontSize: 11, fontWeight: '500' },
  orderMeta: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 },
  orderRole: { fontSize: 11, color: colors.textTertiary },
  orderAmount: { fontSize: 11, color: colors.textSecondary },
  viewDetail: { fontSize: 12, color: colors.primary, fontWeight: '500', textAlign: 'right' },
});
