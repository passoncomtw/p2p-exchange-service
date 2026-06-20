import * as React from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  Alert,
  SafeAreaView,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import {
  validateCreateOrder,
  calcTotalAmount,
  ASSETS,
  FIATS,
  PAYMENT_METHODS,
  type OrderType,
  type PaymentMethod,
} from '@shared';
import * as tokens from '@/theme';
import { ordersApi } from '../../../apis/v1Orders';

const { colors } = tokens;

export default function V1CreateOrderScreen() {
  const { t } = useTranslation();
  const navigation = useNavigation<any>();

  const [type, setType] = React.useState<OrderType>('buy');
  const [price, setPrice] = React.useState('');
  const [quantity, setQuantity] = React.useState('');
  const [paymentMethod, setPaymentMethod] = React.useState<PaymentMethod>('bank_transfer');
  const [errors, setErrors] = React.useState<Record<string, string>>({});
  const [submitting, setSubmitting] = React.useState(false);

  const asset = ASSETS[0];
  const fiat = FIATS[0];
  const total = calcTotalAmount(Number(price), Number(quantity));

  const handleSubmit = async () => {
    const input = {
      type,
      asset,
      fiat,
      price: Number(price),
      quantity: Number(quantity),
      paymentMethod,
    };
    const fieldErrors = validateCreateOrder(input);
    if (fieldErrors.length > 0) {
      const mapped: Record<string, string> = {};
      fieldErrors.forEach((e) => {
        mapped[e.field] = t(e.messageKey);
      });
      setErrors(mapped);
      return;
    }
    setErrors({});
    setSubmitting(true);
    try {
      await ordersApi.create(input);
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
    <SafeAreaView style={styles.safe}>
      <ScrollView contentContainerStyle={styles.container}>
        <Text style={styles.title}>{t('order.pageTitle.create')}</Text>

        {/* 買 / 賣 切換 */}
        <View style={styles.typeRow}>
          {(['buy', 'sell'] as OrderType[]).map((ty) => {
            const active = type === ty;
            const activeColor = tokens.typeColor(ty);
            return (
              <TouchableOpacity
                key={ty}
                onPress={() => setType(ty)}
                style={[
                  styles.typeBtn,
                  active && { backgroundColor: activeColor, borderColor: activeColor },
                ]}
                accessibilityRole="button"
                accessibilityState={{ selected: active }}
              >
                <Text style={[styles.typeBtnText, active && { color: '#fff' }]}>
                  {t(`order.type.${ty}`)}
                </Text>
              </TouchableOpacity>
            );
          })}
        </View>

        {/* 幣種 / 法幣（v1 各僅一種，下拉結構保留） */}
        <View style={styles.fieldRow}>
          <View style={styles.fieldCol}>
            <Text style={styles.label}>{t('order.field.asset')}</Text>
            <View style={styles.readonlyBox}>
              <Text style={styles.readonlyText}>{asset}</Text>
            </View>
          </View>
          <View style={styles.fieldCol}>
            <Text style={styles.label}>{t('order.field.fiat')}</Text>
            <View style={styles.readonlyBox}>
              <Text style={styles.readonlyText}>{fiat}</Text>
            </View>
          </View>
        </View>

        {/* 單價 */}
        <Text style={styles.label}>{t('order.field.price')}</Text>
        <TextInput
          style={[styles.input, errors.price && styles.inputError]}
          value={price}
          onChangeText={setPrice}
          keyboardType="decimal-pad"
          placeholder="0"
          placeholderTextColor={colors.textPlaceholder}
        />
        {errors.price ? <Text style={styles.errorText}>{errors.price}</Text> : null}

        {/* 數量 */}
        <Text style={styles.label}>{t('order.field.quantity')}</Text>
        <TextInput
          style={[styles.input, errors.quantity && styles.inputError]}
          value={quantity}
          onChangeText={setQuantity}
          keyboardType="decimal-pad"
          placeholder="0"
          placeholderTextColor={colors.textPlaceholder}
        />
        {errors.quantity ? <Text style={styles.errorText}>{errors.quantity}</Text> : null}

        {/* 付款方式 */}
        <Text style={styles.label}>{t('order.field.paymentMethod')}</Text>
        <View style={styles.typeRow}>
          {PAYMENT_METHODS.map((pm) => {
            const active = paymentMethod === pm;
            return (
              <TouchableOpacity
                key={pm}
                onPress={() => setPaymentMethod(pm)}
                style={[styles.pmBtn, active && styles.pmBtnActive]}
                accessibilityRole="button"
                accessibilityState={{ selected: active }}
              >
                <Text style={[styles.pmBtnText, active && styles.pmBtnTextActive]}>
                  {t(`order.paymentMethod.${pm}`)}
                </Text>
              </TouchableOpacity>
            );
          })}
        </View>

        {/* 總額 */}
        <View style={styles.totalBox}>
          <Text style={styles.totalLabel}>{t('order.field.totalAmount')}</Text>
          <Text style={styles.totalValue}>
            {total.toLocaleString()} {fiat}
          </Text>
        </View>

        <TouchableOpacity
          style={[styles.submitBtn, submitting && styles.submitBtnDisabled]}
          onPress={handleSubmit}
          disabled={submitting}
          accessibilityRole="button"
        >
          <Text style={styles.submitBtnText}>{t('order.action.submit')}</Text>
        </TouchableOpacity>
      </ScrollView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bgContent },
  container: { padding: 16, gap: 8 },
  title: { fontSize: 16, fontWeight: '600', color: colors.textPrimary, marginBottom: 8 },
  typeRow: { flexDirection: 'row', gap: 12, marginBottom: 4 },
  typeBtn: {
    flex: 1,
    height: 40,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: colors.borderInput,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.bgCard,
  },
  typeBtnText: { fontSize: 14, color: colors.textSecondary, fontWeight: '600' },
  fieldRow: { flexDirection: 'row', gap: 12 },
  fieldCol: { flex: 1 },
  label: { fontSize: 12, color: colors.textSecondary, marginTop: 8, marginBottom: 4 },
  readonlyBox: {
    height: 40,
    borderWidth: 1,
    borderColor: colors.borderInput,
    borderRadius: 4,
    paddingHorizontal: 12,
    justifyContent: 'center',
    backgroundColor: colors.bgCard,
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
  errorText: { fontSize: 12, color: colors.danger, marginTop: 4 },
  pmBtn: {
    flex: 1,
    height: 40,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: colors.borderInput,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.bgCard,
  },
  pmBtnActive: { borderColor: colors.primary, backgroundColor: colors.primaryDisabled },
  pmBtnText: { fontSize: 13, color: colors.textSecondary },
  pmBtnTextActive: { color: colors.textPrimary, fontWeight: '600' },
  totalBox: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: colors.bgCard,
    borderRadius: 4,
    paddingHorizontal: 12,
    paddingVertical: 12,
    marginTop: 12,
  },
  totalLabel: { fontSize: 13, color: colors.textSecondary },
  totalValue: { fontSize: 16, fontWeight: '600', color: colors.textPrimary },
  submitBtn: {
    height: 44,
    borderRadius: 4,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: 16,
  },
  submitBtnDisabled: { backgroundColor: colors.primaryDisabled },
  submitBtnText: { fontSize: 14, fontWeight: '600', color: '#fff' },
});
