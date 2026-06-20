import * as React from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  Alert,
  SafeAreaView,
  RefreshControl,
} from 'react-native';
import { useFocusEffect } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import { canCancel, type Order } from '@shared';
import * as tokens from '@/theme';
import { ordersApi } from '../../../apis/v1Orders';

const { colors } = tokens;

const formatDateTime = (iso: string) => new Date(iso).toLocaleString();

export default function V1MyOrdersScreen() {
  const { t } = useTranslation();
  const [orders, setOrders] = React.useState<Order[]>([]);
  const [loading, setLoading] = React.useState(false);

  const load = React.useCallback(async () => {
    setLoading(true);
    try {
      const list = await ordersApi.listMine();
      setOrders(list);
    } catch {
      Alert.alert('', t('order.message.loadFailed'));
    } finally {
      setLoading(false);
    }
  }, [t]);

  useFocusEffect(
    React.useCallback(() => {
      load();
    }, [load]),
  );

  const confirmCancel = (order: Order) => {
    Alert.alert('', t('order.message.cancelConfirm'), [
      { text: t('order.action.dismiss'), style: 'cancel' },
      {
        text: t('order.action.confirm'),
        style: 'destructive',
        onPress: async () => {
          try {
            await ordersApi.cancel(order.id);
            await load();
          } catch {
            Alert.alert('', t('order.message.submitFailed'));
          }
        },
      },
    ]);
  };

  const renderItem = ({ item }: { item: Order }) => {
    const statusColor = tokens.statusColor(item.status);
    const typeColor = tokens.typeColor(item.type);
    return (
      <View style={styles.card}>
        <View style={styles.cardHeader}>
          <View style={[styles.typeTag, { backgroundColor: typeColor }]}>
            <Text style={styles.typeTagText}>{t(`order.type.${item.type}`)}</Text>
          </View>
          <View style={[styles.statusTag, { borderColor: statusColor }]}>
            <View style={[styles.dot, { backgroundColor: statusColor }]} />
            <Text style={[styles.statusTagText, { color: statusColor }]}>
              {t(`order.status.${item.status}`)}
            </Text>
          </View>
        </View>

        <View style={styles.row}>
          <Text style={styles.rowLabel}>{t('order.field.price')}</Text>
          <Text style={styles.rowValue}>{item.price.toLocaleString()} {item.fiat}</Text>
        </View>
        <View style={styles.row}>
          <Text style={styles.rowLabel}>{t('order.field.quantity')}</Text>
          <Text style={styles.rowValue}>{item.quantity.toLocaleString()} {item.asset}</Text>
        </View>
        <View style={styles.row}>
          <Text style={styles.rowLabel}>{t('order.field.totalAmount')}</Text>
          <Text style={styles.rowValueStrong}>{item.totalAmount.toLocaleString()} {item.fiat}</Text>
        </View>
        <View style={styles.row}>
          <Text style={styles.rowLabel}>{t('order.field.paymentMethod')}</Text>
          <Text style={styles.rowValue}>{t(`order.paymentMethod.${item.paymentMethod}`)}</Text>
        </View>
        <View style={styles.row}>
          <Text style={styles.rowLabel}>{t('order.field.createdAt')}</Text>
          <Text style={styles.rowValue}>{formatDateTime(item.createdAt)}</Text>
        </View>

        {canCancel(item.status) && (
          <TouchableOpacity
            style={styles.cancelBtn}
            onPress={() => confirmCancel(item)}
            accessibilityRole="button"
          >
            <Text style={styles.cancelBtnText}>{t('order.action.cancel')}</Text>
          </TouchableOpacity>
        )}
      </View>
    );
  };

  return (
    <SafeAreaView style={styles.safe}>
      <Text style={styles.title}>{t('order.pageTitle.myOrders')}</Text>
      <FlatList
        data={orders}
        keyExtractor={(o) => o.id}
        renderItem={renderItem}
        contentContainerStyle={styles.listContent}
        refreshControl={<RefreshControl refreshing={loading} onRefresh={load} />}
        ListEmptyComponent={
          !loading ? <Text style={styles.empty}>{t('order.message.emptyMine')}</Text> : null
        }
      />
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bgContent },
  title: { fontSize: 16, fontWeight: '600', color: colors.textPrimary, padding: 16, paddingBottom: 8 },
  listContent: { paddingHorizontal: 16, paddingBottom: 24, gap: 12 },
  card: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    borderWidth: 1,
    borderColor: colors.borderCard,
    padding: 16,
  },
  cardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 },
  typeTag: { paddingHorizontal: 8, paddingVertical: 2, borderRadius: 4 },
  typeTagText: { fontSize: 12, fontWeight: '600', color: '#fff' },
  statusTag: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 6,
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
    borderWidth: 1,
  },
  dot: { width: 6, height: 6, borderRadius: 3 },
  statusTagText: { fontSize: 12, fontWeight: '500' },
  row: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 4 },
  rowLabel: { fontSize: 13, color: colors.textSecondary },
  rowValue: { fontSize: 13, color: colors.textPrimary },
  rowValueStrong: { fontSize: 14, color: colors.textPrimary, fontWeight: '600' },
  cancelBtn: {
    marginTop: 12,
    height: 36,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: colors.danger,
    alignItems: 'center',
    justifyContent: 'center',
  },
  cancelBtnText: { fontSize: 13, color: colors.danger, fontWeight: '500' },
  empty: { textAlign: 'center', color: colors.textTertiary, fontSize: 13, paddingVertical: 48 },
});
