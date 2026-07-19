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
import * as tokens from '@/theme';
import { paymentMethodsApi } from '@/apis/paymentMethodsApi';
import { useAppDispatch } from '@/navigation/store/hooks';
import { pushNotification } from '@/navigation/store/slices/notificationSlice';

const { colors } = tokens;

export default function V1AddPaymentMethodScreen() {
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const navigation = useNavigation<any>();

  const [bankName, setBankName] = React.useState('');
  const [accountName, setAccountName] = React.useState('');
  const [accountNumber, setAccountNumber] = React.useState('');
  const [errors, setErrors] = React.useState<Record<string, boolean>>({});
  const [submitting, setSubmitting] = React.useState(false);

  const handleSubmit = async () => {
    const errs: Record<string, boolean> = {};
    if (!bankName.trim()) errs.bankName = true;
    if (!accountName.trim()) errs.accountName = true;
    if (!accountNumber.trim()) errs.accountNumber = true;
    if (Object.keys(errs).length > 0) {
      setErrors(errs);
      return;
    }
    setErrors({});
    setSubmitting(true);
    try {
      await paymentMethodsApi.create({
        type: 'bank_transfer',
        bankName: bankName.trim(),
        accountName: accountName.trim(),
        accountNumber: accountNumber.trim(),
      });
      dispatch(pushNotification({ type: 'success', message: t('order.message.addPaymentSuccess') }));
      navigation.goBack();
    } catch {
      dispatch(pushNotification({ type: 'error', message: t('order.message.submitFailed') }));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <ScrollView style={styles.screen} contentContainerStyle={styles.container}>
      <Text style={styles.label}>
        {t('order.field.bankName')} <Text style={styles.req}>*</Text>
      </Text>
      <TextInput
        style={[styles.input, errors.bankName && styles.inputError]}
        value={bankName}
        onChangeText={setBankName}
        placeholder={t('order.field.bankName')}
        placeholderTextColor={colors.textPlaceholder}
      />

      <Text style={[styles.label, { marginTop: 16 }]}>
        {t('order.field.accountName')} <Text style={styles.req}>*</Text>
      </Text>
      <TextInput
        style={[styles.input, errors.accountName && styles.inputError]}
        value={accountName}
        onChangeText={setAccountName}
        placeholder={t('order.field.accountName')}
        placeholderTextColor={colors.textPlaceholder}
      />

      <Text style={[styles.label, { marginTop: 16 }]}>
        {t('order.field.accountNumber')} <Text style={styles.req}>*</Text>
      </Text>
      <TextInput
        style={[styles.input, errors.accountNumber && styles.inputError]}
        value={accountNumber}
        onChangeText={setAccountNumber}
        placeholder={t('order.field.accountNumber')}
        placeholderTextColor={colors.textPlaceholder}
        keyboardType="number-pad"
      />

      <TouchableOpacity
        style={[styles.submitBtn, submitting && styles.submitBtnDisabled]}
        onPress={handleSubmit}
        disabled={submitting}
        accessibilityRole="button"
      >
        {submitting ? <ActivityIndicator size="small" color="#1F2327" style={{ marginRight: 8 }} /> : null}
        <Text style={styles.submitBtnText}>{t('order.action.addPayment')}</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bgContent },
  container: { padding: 16, paddingBottom: 28 },
  label: { fontSize: 12, color: colors.textSecondary, marginBottom: 6 },
  req: { color: colors.danger },
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
  submitBtn: {
    marginTop: 24,
    height: 48,
    borderRadius: 4,
    backgroundColor: colors.primary,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
  submitBtnDisabled: { backgroundColor: colors.primaryDisabled },
  submitBtnText: { fontSize: 16, fontWeight: '600', color: '#1F2327' },
});
