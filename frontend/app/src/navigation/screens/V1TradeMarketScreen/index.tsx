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
import { listingsApi } from '@/apis/listingsApi';
import { useAppSelector } from '@/navigation/store/hooks';
import type { ListingItem } from '@/interfaces/listing';

const { colors } = tokens;
const formatDateTime = (iso: string) => new Date(iso).toLocaleString();

export default function V1TradeMarketScreen() {
  const { t } = useTranslation();
  const navigation = useNavigation<any>();
  const currentUserId = useAppSelector((s) => s.auth.user?.id);
  const [listings, setListings] = React.useState<ListingItem[]>([]);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);

  const load = React.useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      const all = await listingsApi.list({ status: 'active' });
      setListings(all.filter((l) => l.userId !== currentUserId));
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [currentUserId]);

  useFocusEffect(
    React.useCallback(() => {
      load();
    }, [load]),
  );

  const Title = () => <Text style={styles.title}>{t('order.pageTitle.trade')}</Text>;

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
          <Text style={styles.errorBody}>{t('order.message.loadFailed')}</Text>
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
          <View style={styles.emptyIcon}><Text style={styles.emptyIconText}>~</Text></View>
          <Text style={styles.emptyText}>{t('order.message.emptyTrade')}</Text>
        </View>
      </ScrollView>
    );
  }

  const renderItem = ({ item }: { item: ListingItem }) => {
    const tColor = tokens.typeColor(item.type);
    const fiatTotal = item.price * item.remainingAmount;
    return (
      <TouchableOpacity
        style={styles.card}
        onPress={() => navigation.navigate('ListingDetail', { id: item.id })}
        accessibilityRole="button"
      >
        <View style={styles.cardHeader}>
          <View style={[styles.typeTag, { backgroundColor: tColor }]}>
            <Text style={styles.typeTagText}>{t(`order.type.${item.type}`)}</Text>
          </View>
          <Text style={styles.cryptoLabel}>{item.cryptoCurrency}/{item.fiatCurrency}</Text>
        </View>
        <Row label={t('order.field.price')} value={`${item.price.toLocaleString()} ${item.fiatCurrency}`} />
        <Row label={t('order.field.remainingAmount')} value={`${item.remainingAmount} ${item.cryptoCurrency}`} />
        <Row label={t('order.field.fiatTotal')} value={`${fiatTotal.toLocaleString()} ${item.fiatCurrency}`} amber />
        <Row label={t('order.field.createdAt')} value={formatDateTime(item.createdAt)} muted />
      </TouchableOpacity>
    );
  };

  return (
    <FlatList
      style={styles.screen}
      data={listings}
      keyExtractor={(l) => String(l.id)}
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
  card: { backgroundColor: colors.bgCard, borderRadius: 8, borderWidth: 1, borderColor: colors.borderCard, padding: 14 },
  cardHeader: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 },
  typeTag: { paddingHorizontal: 8, paddingVertical: 2, borderRadius: 4 },
  typeTagText: { fontSize: 12, fontWeight: '600', color: '#fff' },
  cryptoLabel: { fontSize: 12, color: colors.textTertiary, fontWeight: '500' },
  row: { flexDirection: 'row', justifyContent: 'space-between', paddingVertical: 3 },
  rowLabel: { fontSize: 12, color: colors.textSecondary },
  rowValue: { fontSize: 12, color: colors.textPrimary, fontWeight: '500' },
  rowAmber: { color: colors.amberText, fontWeight: '700' },
  rowMuted: { color: colors.textTertiary, fontWeight: '400' },
});
