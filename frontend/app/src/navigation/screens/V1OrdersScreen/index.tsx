import * as React from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  RefreshControl,
} from 'react-native';
import { useFocusEffect, useNavigation } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import * as tokens from '@/theme';
import { p2pOrdersApi } from '@/apis/p2pOrdersApi';
import { useAppSelector } from '@/navigation/store/hooks';
import type { Order } from '@/interfaces/order';

const { colors, orderStatusColors } = tokens;
const formatDateTime = (iso: string) => new Date(iso).toLocaleString();

const STATUS_TABS = ['all', 'matched', 'paid', 'completed', 'cancelled'] as const;

export default function V1OrdersScreen() {
  const { t } = useTranslation();
  const navigation = useNavigation<any>();
  const currentUserId = useAppSelector((s) => s.auth.user?.id);
  const [orders, setOrders] = React.useState<Order[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);
  const [tab, setTab] = React.useState<string>('all');

  const load = React.useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      const params = tab === 'all' ? {} : { status: tab };
      const list = await p2pOrdersApi.list(params);
      setOrders(list);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [tab]);

  useFocusEffect(
    React.useCallback(() => {
      load();
    }, [load]),
  );

  const Title = () => <Text style={styles.title}>{t('order.nav.orders')}</Text>;

  const Tabs = () => (
    <ScrollView horizontal showsHorizontalScrollIndicator={false} style={styles.tabsRow} contentContainerStyle={styles.tabsContent}>
      {STATUS_TABS.map((s) => {
        const active = tab === s;
        return (
          <TouchableOpacity
            key={s}
            onPress={() => setTab(s)}
            style={[styles.tab, active && styles.tabActive]}
            accessibilityRole="button"
            accessibilityState={{ selected: active }}
          >
            <Text style={[styles.tabText, active && styles.tabTextActive]}>
              {t(`order.orderStatus.${s}`)}
            </Text>
          </TouchableOpacity>
        );
      })}
    </ScrollView>
  );

  if (loading) {
    return (
      <ScrollView style={styles.screen} contentContainerStyle={styles.pad}>
        <Title />
        <Tabs />
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
          <TouchableOpacity style={styles.outlineBtn} onPress={load} accessibilityRole="button">
            <Text style={styles.outlineBtnText}>{t('order.action.retry')}</Text>
          </TouchableOpacity>
        </View>
      </ScrollView>
    );
  }

  if (orders.length === 0) {
    return (
      <ScrollView style={styles.screen} contentContainerStyle={styles.pad}>
        <Title />
        <Tabs />
        <View style={styles.centerBox}>
          <View style={styles.emptyIcon}><Text style={styles.emptyIconText}>~</Text></View>
          <Text style={styles.emptyText}>{t('order.message.emptyOrders')}</Text>
        </View>
      </ScrollView>
    );
  }

  const renderItem = ({ item }: { item: Order }) => {
    const sColor = orderStatusColors[item.status] ?? colors.statusCancelled;
    const isBuyer = item.buyerId === currentUserId;
    const roleKey = isBuyer ? 'asBuyer' : 'asSeller';
    return (
      <TouchableOpacity
        style={styles.card}
        onPress={() => navigation.navigate('OrderDetail', { id: item.id })}
        accessibilityRole="button"
      >
        <View style={styles.cardHeader}>
          <Text style={styles.orderNo}>{item.orderNo}</Text>
          <View style={[styles.statusTag, { borderColor: sColor }]}>
            <View style={[styles.dot, { backgroundColor: sColor }]} />
            <Text style={[styles.statusText, { color: sColor }]}>{t(`order.orderStatus.${item.status}`)}</Text>
          </View>
        </View>
        <View style={styles.roleRow}>
          <Text style={styles.roleText}>{t(`order.role.${roleKey}`)}</Text>
        </View>
        <Row label={t('order.field.price')} value={`${item.price.toLocaleString()} ${item.fiatCurrency}`} />
        <Row label={t('order.field.quantity')} value={`${item.cryptoAmount} ${item.cryptoCurrency}`} />
        <Row label={t('order.field.totalAmount')} value={`${item.totalAmount.toLocaleString()} ${item.fiatCurrency}`} amber />
        <Row label={t('order.field.createdAt')} value={formatDateTime(item.createdAt)} muted />
      </TouchableOpacity>
    );
  };

  return (
    <FlatList
      style={styles.screen}
      data={orders}
      keyExtractor={(o) => String(o.id)}
      renderItem={renderItem}
      ListHeaderComponent={<><Title /><Tabs /></>}
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
  title: { fontSize: 16, fontWeight: '600', color: colors.textPrimary, marginBottom: 10 },
  tabsRow: { marginBottom: 14 },
  tabsContent: { gap: 8 },
  tab: { paddingHorizontal: 14, paddingVertical: 6, borderRadius: 16, backgroundColor: colors.bgCard, borderWidth: 1, borderColor: colors.borderCard },
  tabActive: { backgroundColor: colors.primary, borderColor: colors.primary },
  tabText: { fontSize: 12, color: colors.textSecondary },
  tabTextActive: { color: '#1F2327', fontWeight: '600' },
  skeleton: { height: 96, backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard },
  centerBox: { alignItems: 'center', paddingVertical: 56, paddingHorizontal: 20 },
  errorMark: { fontSize: 32, color: colors.danger, marginBottom: 8, fontWeight: '700' },
  errorTitle: { fontSize: 14, color: colors.textPrimary, fontWeight: '500', marginBottom: 12 },
  outlineBtn: { height: 36, paddingHorizontal: 20, borderWidth: 1, borderColor: colors.borderInput, borderRadius: 4, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.bgCard },
  outlineBtnText: { fontSize: 14, fontWeight: '500', color: colors.textPrimary },
  emptyIcon: { width: 56, height: 56, borderRadius: 28, backgroundColor: '#EFEFF2', alignItems: 'center', justifyContent: 'center', marginBottom: 14 },
  emptyIconText: { fontSize: 24, color: '#BDBDBD' },
  emptyText: { fontSize: 14, color: colors.textSecondary, marginBottom: 16 },
  card: { backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard, padding: 14 },
  cardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 },
  orderNo: { fontSize: 13, fontWeight: '600', color: colors.textPrimary },
  statusTag: { flexDirection: 'row', alignItems: 'center', gap: 6, paddingHorizontal: 8, paddingVertical: 2, borderRadius: 4, borderWidth: 1 },
  dot: { width: 6, height: 6, borderRadius: 3 },
  statusText: { fontSize: 12, fontWeight: '500' },
  roleRow: { marginBottom: 8 },
  roleText: { fontSize: 11, color: colors.textTertiary },
  row: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 3 },
  rowLabel: { fontSize: 12, color: colors.textSecondary },
  rowValue: { fontSize: 12, color: colors.textPrimary, fontWeight: '500' },
  rowAmber: { color: colors.amberText, fontWeight: '700' },
  rowMuted: { color: colors.textTertiary, fontWeight: '400' },
});
