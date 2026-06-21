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
import { useNavigation, useRoute } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import * as tokens from '@/theme';
import { listingsApi } from '@/apis/listingsApi';
import { p2pOrdersApi } from '@/apis/p2pOrdersApi';
import { paymentMethodsApi } from '@/apis/paymentMethodsApi';
import type { ListingItem } from '@/interfaces/listing';

const { colors } = tokens;
const formatDateTime = (iso: string) => new Date(iso).toLocaleString();

export default function V1ListingDetailScreen() {
  const { t } = useTranslation();
  const navigation = useNavigation<any>();
  const route = useRoute<any>();
  const listingId: number = route.params?.id;

  const [listing, setListing] = React.useState<ListingItem | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState(false);
  const [taking, setTaking] = React.useState(false);

  const load = React.useCallback(async () => {
    setLoading(true);
    setError(false);
    try {
      const data = await listingsApi.getById(listingId);
      setListing(data);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, [listingId]);

  React.useEffect(() => { load(); }, [load]);

  const handleTakeOrder = async () => {
    if (!listing) return;

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
        Alert.alert('', t('order.message.submitFailed'));
        return;
      }
    }

    setTaking(true);
    try {
      const result = await p2pOrdersApi.create({
        listingId: listing.id,
        cryptoAmount: listing.remainingAmount,
      });
      Alert.alert('', t('order.message.takeOrderSuccess'));
      navigation.replace('OrderDetail', { id: result.id });
    } catch {
      Alert.alert('', t('order.message.takeOrderFailed'));
    } finally {
      setTaking(false);
    }
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
  const fiatTotal = listing.price * listing.remainingAmount;
  const tColor = tokens.typeColor(listing.type);
  const actionLabel = listing.type === 'sell' ? t('order.action.takeBuy') : t('order.action.takeSell');

  return (
    <ScrollView style={styles.screen} contentContainerStyle={styles.container}>
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
        <Row label={t('order.field.fiatTotal')} value={`${fiatTotal.toLocaleString()} ${listing.fiatCurrency}`} amber />
        <Row label={t('order.field.createdAt')} value={formatDateTime(listing.createdAt)} muted />
      </View>

      {!isActive ? (
        <View style={styles.unavailableBox}>
          <Text style={styles.unavailableText}>{t('order.message.listingUnavailable')}</Text>
        </View>
      ) : (
        <TouchableOpacity
          style={[styles.actionBtn, { backgroundColor: tColor }, taking && styles.actionBtnDisabled]}
          onPress={handleTakeOrder}
          disabled={taking}
          accessibilityRole="button"
        >
          {taking ? <ActivityIndicator size="small" color="#fff" style={{ marginRight: 8 }} /> : null}
          <Text style={styles.actionBtnText}>{actionLabel}</Text>
        </TouchableOpacity>
      )}
    </ScrollView>
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
  actionBtn: {
    height: 48,
    borderRadius: 4,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
  actionBtnDisabled: { opacity: 0.6 },
  actionBtnText: { fontSize: 16, fontWeight: '600', color: '#fff' },
});
