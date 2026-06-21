import * as React from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { useRoute } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import * as tokens from '@/theme';
import { p2pOrdersApi } from '@/apis/p2pOrdersApi';
import { useAppSelector } from '@/navigation/store/hooks';
import type { Order } from '@/interfaces/order';

const { colors, orderStatusColors } = tokens;
const formatDateTime = (iso: string) => new Date(iso).toLocaleString();

export default function V1OrderDetailScreen() {
  const { t } = useTranslation();
  const route = useRoute<any>();
  const orderId: number = route.params?.id;
  const currentUserId = useAppSelector((s) => s.auth.user?.id);

  const [order, setOrder] = React.useState<Order | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);
  const [acting, setActing] = React.useState(false);

  const load = React.useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      const data = await p2pOrdersApi.getById(orderId);
      setOrder(data);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [orderId]);

  React.useEffect(() => { load(); }, [load]);

  const doAction = (action: () => Promise<void>, successMsg: string) => {
    return async () => {
      setActing(true);
      try {
        await action();
        Alert.alert('', t(successMsg));
        await load();
      } catch {
        Alert.alert('', t('order.message.submitFailed'));
      } finally {
        setActing(false);
      }
    };
  };

  const confirmAction = (titleKey: string, action: () => Promise<void>, successKey: string) => {
    Alert.alert('', t(titleKey), [
      { text: t('order.action.dismiss'), style: 'cancel' },
      { text: t('order.action.confirm'), style: 'destructive', onPress: doAction(action, successKey) },
    ]);
  };

  if (loading) {
    return (
      <ScrollView style={styles.screen} contentContainerStyle={styles.container}>
        <View style={{ gap: 12 }}>
          {[0.6, 0.4, 0.25].map((o, i) => (
            <View key={i} style={[styles.skeleton, { opacity: o }]} />
          ))}
        </View>
      </ScrollView>
    );
  }

  if (error || !order) {
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

  const isBuyer = order.buyerId === currentUserId;
  const isSeller = order.sellerId === currentUserId;
  const sColor = orderStatusColors[order.status] ?? colors.statusCancelled;
  const roleKey = isBuyer ? 'asBuyer' : 'asSeller';

  return (
    <ScrollView style={styles.screen} contentContainerStyle={styles.container}>
      {/* Status header */}
      <View style={styles.statusHeader}>
        <View style={[styles.statusBadge, { backgroundColor: sColor }]}>
          <Text style={styles.statusBadgeText}>{t(`order.orderStatus.${order.status}`)}</Text>
        </View>
        <Text style={styles.roleLabel}>{t(`order.role.${roleKey}`)}</Text>
      </View>

      {/* Order info */}
      <View style={styles.section}>
        <Row label={t('order.field.orderNo')} value={order.orderNo} />
        <Row label={t('order.field.price')} value={`${order.price.toLocaleString()} ${order.fiatCurrency}`} />
        <Row label={t('order.field.quantity')} value={`${order.cryptoAmount} ${order.cryptoCurrency}`} />
        <Row label={t('order.field.totalAmount')} value={`${order.totalAmount.toLocaleString()} ${order.fiatCurrency}`} amber />
        <Row label={t('order.field.paymentDeadline')} value={formatDateTime(order.paymentDeadline)} muted />
        <Row label={t('order.field.createdAt')} value={formatDateTime(order.createdAt)} muted />
        {order.paidAt && <Row label={t('order.orderStatus.paid')} value={formatDateTime(order.paidAt)} muted />}
        {order.completedAt && <Row label={t('order.orderStatus.completed')} value={formatDateTime(order.completedAt)} muted />}
        {order.cancelledAt && <Row label={t('order.orderStatus.cancelled')} value={formatDateTime(order.cancelledAt)} muted />}
        {order.cancelReason && <Row label={t('order.action.cancelOrder')} value={order.cancelReason} />}
      </View>

      {/* Actions */}
      <View style={styles.actions}>
        {/* Buyer matched: markPaid + cancel */}
        {isBuyer && order.status === 'matched' && (
          <>
            <ActionButton
              label={t('order.action.markPaid')}
              color="#2196F3"
              loading={acting}
              onPress={() => confirmAction('order.message.markPaidConfirm', () => p2pOrdersApi.markPaid(order.id), 'order.message.markPaidSuccess')}
            />
            <ActionButton
              label={t('order.action.cancelOrder')}
              color={colors.danger}
              outline
              loading={acting}
              onPress={() => confirmAction('order.message.cancelOrderConfirm', () => p2pOrdersApi.cancel(order.id, ''), 'order.message.cancelOrderSuccess')}
            />
          </>
        )}

        {/* Seller matched: cancel */}
        {isSeller && order.status === 'matched' && (
          <ActionButton
            label={t('order.action.cancelOrder')}
            color={colors.danger}
            outline
            loading={acting}
            onPress={() => confirmAction('order.message.cancelOrderConfirm', () => p2pOrdersApi.cancel(order.id, ''), 'order.message.cancelOrderSuccess')}
          />
        )}

        {/* Seller paid: confirm */}
        {isSeller && order.status === 'paid' && (
          <ActionButton
            label={t('order.action.confirmReceipt')}
            color="#4CAF50"
            loading={acting}
            onPress={() => confirmAction('order.message.confirmReceiptConfirm', () => p2pOrdersApi.confirm(order.id), 'order.message.confirmReceiptSuccess')}
          />
        )}

        {/* Dispute: buyer or seller, paid status */}
        {(isBuyer || isSeller) && order.status === 'paid' && (
          <ActionButton
            label={t('order.action.dispute')}
            color={colors.danger}
            outline
            loading={acting}
            onPress={() => confirmAction('order.message.disputeConfirm', () => p2pOrdersApi.dispute(order.id, ''), 'order.message.disputeSuccess')}
          />
        )}
      </View>
    </ScrollView>
  );
}

function ActionButton({
  label,
  color,
  outline,
  loading,
  onPress,
}: {
  label: string;
  color: string;
  outline?: boolean;
  loading: boolean;
  onPress: () => void;
}) {
  return (
    <TouchableOpacity
      style={[
        styles.actionBtn,
        outline
          ? { borderWidth: 1, borderColor: color, backgroundColor: 'transparent' }
          : { backgroundColor: color },
        loading && styles.actionBtnDisabled,
      ]}
      onPress={onPress}
      disabled={loading}
      accessibilityRole="button"
    >
      {loading ? <ActivityIndicator size="small" color={outline ? color : '#fff'} style={{ marginRight: 8 }} /> : null}
      <Text style={[styles.actionBtnText, { color: outline ? color : '#fff' }]}>{label}</Text>
    </TouchableOpacity>
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
  skeleton: { height: 48, backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard },
  centerBox: { alignItems: 'center', paddingVertical: 56, paddingHorizontal: 20 },
  errorMark: { fontSize: 32, color: colors.danger, marginBottom: 8, fontWeight: '700' },
  errorTitle: { fontSize: 14, color: colors.textPrimary, fontWeight: '500', marginBottom: 12 },
  outlineBtn: { height: 36, paddingHorizontal: 20, borderWidth: 1, borderColor: colors.borderInput, borderRadius: 4, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.bgCard },
  outlineBtnText: { fontSize: 14, fontWeight: '500', color: colors.textPrimary },
  statusHeader: { flexDirection: 'row', alignItems: 'center', gap: 12, marginBottom: 16 },
  statusBadge: { paddingHorizontal: 12, paddingVertical: 6, borderRadius: 4 },
  statusBadgeText: { fontSize: 14, fontWeight: '600', color: '#fff' },
  roleLabel: { fontSize: 13, color: colors.textTertiary },
  section: { backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard, padding: 14, marginBottom: 20 },
  row: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 5 },
  rowLabel: { fontSize: 13, color: colors.textSecondary },
  rowValue: { fontSize: 13, color: colors.textPrimary, fontWeight: '500', flexShrink: 1, textAlign: 'right' },
  rowAmber: { color: colors.amberText, fontWeight: '700' },
  rowMuted: { color: colors.textTertiary, fontWeight: '400' },
  actions: { gap: 12 },
  actionBtn: {
    height: 48,
    borderRadius: 4,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
  actionBtnDisabled: { opacity: 0.6 },
  actionBtnText: { fontSize: 16, fontWeight: '600' },
});
