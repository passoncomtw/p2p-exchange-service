import * as React from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import {
  validateCreateOrder,
  calcTotalAmount,
  ASSETS,
  FIATS,
  type OrderType,
} from '@shared';
import * as tokens from '@/theme';
import { listingsApi } from '@/apis/listingsApi';
import { paymentMethodsApi, type PaymentMethodItem } from '@/apis/paymentMethodsApi';

const { colors } = tokens;

export default function V1CreateOrderScreen() {
  const { t } = useTranslation();
  const navigation = useNavigation<any>();

  const [type, setType] = React.useState<OrderType>('buy');
  const [price, setPrice] = React.useState('');
  const [quantity, setQuantity] = React.useState('');
  const [paymentMethods, setPaymentMethods] = React.useState<PaymentMethodItem[]>([]);
  const [selectedPaymentId, setSelectedPaymentId] = React.useState<number | null>(null);
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const [submitting, setSubmitting] = React.useState(false);

  const asset = ASSETS[0];
  const fiat = FIATS[0];
  const total = calcTotalAmount(Number(price), Number(quantity));
  const isBuy = type === 'buy';

  React.useEffect(() => {
    paymentMethodsApi.list().then((list) => {
      setPaymentMethods(list);
      if (list.length > 0) setSelectedPaymentId(list[0].id);
    }).catch(() => {});
  }, []);

  const handleSubmit = async () => {
    const input = { type, asset, fiat, price: Number(price), quantity: Number(quantity), paymentMethod: 'bank_transfer' as const };
    const fieldErrors = validateCreateOrder(input);
    if (fieldErrors.length > 0) {
      const mapped: Record<string, string> = {};
      fieldErrors.forEach((e) => {
        mapped[e.field] = t(e.messageKey);
      });
      setErrors(mapped);
      return;
    }
    if (!isBuy && !selectedPaymentId) {
      Alert.alert('', t('order.message.noPaymentMethod'), [
        { text: t('order.action.dismiss'), style: 'cancel' },
        { text: t('order.action.addPayment'), onPress: () => navigation.navigate('AddPaymentMethod') },
      ]);
      return;
    }
    setErrors({});
    setSubmitting(true);
    try {
      const priceNum = Number(price);
      const quantityNum = Number(quantity);
      await listingsApi.create({
        type,
        price: priceNum,
        totalAmount: quantityNum,
        minOrderFiat: 0,
        maxOrderFiat: priceNum * quantityNum,
        paymentMethodId: isBuy ? null : selectedPaymentId,
      });
      Alert.alert('', t('order.message.createSuccess'));
      setPrice('');
      setQuantity('');
      navigation.navigate('MyOrders');
    } catch {
      Alert.alert('', t('order.message.submitFailed'));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <ScrollView style={styles.screen} contentContainerStyle={styles.container}>
      {/* 買/賣切換 */}
      <Text style={styles.sectionLabel}>{t('order.create.sectionLabel')}</Text>
      <View style={styles.typeRow}>
        {(['buy', 'sell'] as OrderType[]).map((ty) => {
          const active = type === ty;
          const activeColor = tokens.typeColor(ty);
          return (
            <TouchableOpacity
              key={ty}
              onPress={() => setType(ty)}
              style={[styles.typeBtn, active && { backgroundColor: activeColor, borderColor: activeColor }]}
              accessibilityRole="button"
              accessibilityState={{ selected: active }}
            >
              <Text style={[styles.typeBtnText, active && { color: '#fff' }]}>
                {t(`order.create.${ty}Label`)}
              </Text>
            </TouchableOpacity>
          );
        })}
      </View>

      {/* 類型說明 */}
      <View style={styles.hintBox}>
        <Text style={styles.hintText}>
          {isBuy ? t('order.create.buyHint') : t('order.create.sellHint')}
        </Text>
      </View>

      {/* 幣種 / 法幣 */}
      <View style={styles.fieldRow}>
        <View style={styles.fieldCol}>
          <Text style={styles.label}>{t('order.field.asset')}</Text>
          <View style={styles.readonlyBox}><Text style={styles.readonlyText}>{asset}</Text></View>
        </View>
        <View style={styles.fieldCol}>
          <Text style={styles.label}>{t('order.field.fiat')}</Text>
          <View style={styles.readonlyBox}><Text style={styles.readonlyText}>{fiat}</Text></View>
        </View>
      </View>

      {/* 單價 */}
      <Text style={styles.label}>
        {t('order.create.priceLabel')}({t('order.create.priceBuyUnit')}) <Text style={styles.req}>*</Text>
      </Text>
      <TextInput
        style={[styles.input, errors.price && styles.inputError]}
        value={price}
        onChangeText={setPrice}
        keyboardType="decimal-pad"
        placeholder={isBuy ? t('order.create.priceBuyPlaceholder') : t('order.create.priceSellPlaceholder')}
        placeholderTextColor={colors.textPlaceholder}
      />
      {errors.price ? <Text style={styles.errorText}>{errors.price}</Text> : null}

      {/* 數量 */}
      <Text style={[styles.label, { marginTop: 16 }]}>
        {t('order.create.quantityLabel')}({t('order.create.quantityUnit')}) <Text style={styles.req}>*</Text>
      </Text>
      <TextInput
        style={[styles.input, errors.quantity && styles.inputError]}
        value={quantity}
        onChangeText={setQuantity}
        keyboardType="decimal-pad"
        placeholder={isBuy ? t('order.create.quantityBuyPlaceholder') : t('order.create.quantitySellPlaceholder')}
        placeholderTextColor={colors.textPlaceholder}
      />
      {errors.quantity ? <Text style={styles.errorText}>{errors.quantity}</Text> : null}

      {/* 收款帳戶(sell 掛單必選) */}
      {!isBuy && (
        <>
          <Text style={[styles.label, { marginTop: 16 }]}>
            {t('order.create.paymentMethodLabel')} <Text style={styles.req}>*</Text>
          </Text>
          <Text style={styles.payHint}>{t('order.create.paymentMethodHint')}</Text>
          {paymentMethods.length === 0 ? (
            <TouchableOpacity
              style={styles.payBtn}
              onPress={() => navigation.navigate('AddPaymentMethod')}
              accessibilityRole="button"
            >
              <Text style={styles.payText}>{t('order.action.addPayment')}</Text>
            </TouchableOpacity>
          ) : (
            <View style={styles.payCol}>
              {paymentMethods.map((pm) => {
                const active = selectedPaymentId === pm.id;
                return (
                  <TouchableOpacity
                    key={pm.id}
                    onPress={() => setSelectedPaymentId(pm.id)}
                    style={[styles.payBtn, active && styles.payBtnActive]}
                    accessibilityRole="button"
                    accessibilityState={{ selected: active }}
                  >
                    <View style={[styles.payDot, active && styles.payDotActive]} />
                    <Text style={[styles.payText, active && styles.payTextActive]}>
                      {pm.bankName} - {pm.accountNumber}
                    </Text>
                  </TouchableOpacity>
                );
              })}
            </View>
          )}
        </>
      )}

      {/* 總額(琥珀高亮) */}
      <View style={styles.totalBox}>
        <Text style={styles.totalLabel}>
          {isBuy ? t('order.create.totalBuy') : t('order.create.totalSell')}
        </Text>
        <Text style={styles.totalValue}>
          {total.toLocaleString()} <Text style={styles.totalUnit}>{fiat}</Text>
        </Text>
      </View>

      <TouchableOpacity
        style={[
          styles.submitBtn,
          { backgroundColor: tokens.typeColor(type) },
          submitting && styles.submitBtnDisabled,
        ]}
        onPress={handleSubmit}
        disabled={submitting}
        accessibilityRole="button"
      >
        {submitting ? <ActivityIndicator size="small" color="#fff" style={{ marginRight: 8 }} /> : null}
        <Text style={styles.submitBtnText}>
          {isBuy ? t('order.create.submitBuy') : t('order.create.submitSell')}
        </Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bgContent },
  container: { padding: 16, paddingBottom: 28 },
  sectionLabel: { fontSize: 13, fontWeight: '600', color: colors.textSecondary, marginBottom: 8 },
  typeRow: { flexDirection: 'row', gap: 10, marginBottom: 12 },
  typeBtn: {
    flex: 1,
    height: 48,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: colors.borderInput,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.bgCard,
  },
  typeBtnText: { fontSize: 15, color: colors.textSecondary, fontWeight: '600' },
  hintBox: {
    backgroundColor: colors.bgCard,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: colors.borderCard,
    paddingHorizontal: 12,
    paddingVertical: 10,
    marginBottom: 20,
  },
  hintText: { fontSize: 12, color: colors.textTertiary, lineHeight: 18 },
  fieldRow: { flexDirection: 'row', gap: 12, marginBottom: 16 },
  fieldCol: { flex: 1 },
  label: { fontSize: 12, color: colors.textSecondary, marginBottom: 6 },
  req: { color: colors.danger },
  readonlyBox: {
    height: 40,
    borderWidth: 1,
    borderColor: colors.borderInput,
    borderRadius: 4,
    paddingHorizontal: 12,
    justifyContent: 'center',
    backgroundColor: colors.bgContent,
  },
  readonlyText: { fontSize: 13, color: colors.textPrimary },
  input: {
    height: 40,
    borderWidth: 1,
    borderColor: colors.borderInput,
    borderRadius: 4,
    paddingHorizontal: 12,
    fontSize: 13,
    color: colors.textPrimary,
    backgroundColor: colors.bgCard,
  },
  inputError: { borderColor: colors.danger },
  errorText: { fontSize: 11, color: colors.danger, marginTop: 5 },
  payHint: { fontSize: 11, color: colors.textTertiary, marginBottom: 8, lineHeight: 16 },
  payCol: { gap: 8 },
  payBtn: {
    height: 44,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: colors.borderInput,
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 14,
    gap: 10,
    backgroundColor: colors.bgCard,
  },
  payBtnActive: { borderColor: colors.primary, backgroundColor: colors.amberBg },
  payDot: { width: 14, height: 14, borderRadius: 7, borderWidth: 2, borderColor: colors.borderInput },
  payDotActive: { borderColor: colors.primary, backgroundColor: colors.primary },
  payText: { fontSize: 14, color: colors.textSecondary },
  payTextActive: { color: colors.textPrimary, fontWeight: '600' },
  totalBox: {
    marginTop: 20,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: colors.amberBg,
    borderWidth: 1,
    borderColor: colors.amberBorder,
    borderRadius: 4,
    paddingHorizontal: 14,
    paddingVertical: 12,
  },
  totalLabel: { fontSize: 12, color: colors.textSecondary },
  totalValue: { fontSize: 18, fontWeight: '700', color: colors.amberText },
  totalUnit: { fontSize: 12, fontWeight: '400', color: colors.textTertiary },
  submitBtn: {
    marginTop: 20,
    height: 48,
    borderRadius: 4,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
  submitBtnDisabled: { opacity: 0.6 },
  submitBtnText: { fontSize: 16, fontWeight: '600', color: '#fff' },
});
