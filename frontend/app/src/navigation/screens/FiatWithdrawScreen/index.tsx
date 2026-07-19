import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAppDispatch } from '@/navigation/store/hooks';
import { pushNotification } from '@/navigation/store/slices/notificationSlice';
import { walletApi } from '@/apis/walletApi';
import { colors } from '@/theme';

const MIN_AMOUNT = 100;

export default function FiatWithdrawScreen() {
  const dispatch = useAppDispatch();
  const navigation = useNavigation<any>();

  const [amount, setAmount] = useState('');
  const [bankCode, setBankCode] = useState('');
  const [bankAccount, setBankAccount] = useState('');
  const [accountName, setAccountName] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validate = (): boolean => {
    const errs: Record<string, string> = {};
    const num = parseFloat(amount);
    if (!amount.trim() || isNaN(num)) {
      errs.amount = '請輸入有效金額';
    } else if (num < MIN_AMOUNT) {
      errs.amount = `最低提領金額為 ${MIN_AMOUNT} TWD`;
    }
    if (!bankCode.trim()) errs.bankCode = '請輸入銀行代碼';
    if (!bankAccount.trim()) errs.bankAccount = '請輸入銀行帳號';
    if (!accountName.trim()) errs.accountName = '請輸入戶名';
    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = () => {
    if (!validate()) return;

    Alert.alert(
      '確認提領',
      `提領 ${amount} TWD\n至 ${bankCode} ${bankAccount}（${accountName}）`,
      [
        { text: '取消', style: 'cancel' },
        {
          text: '確認',
          onPress: async () => {
            setSubmitting(true);
            try {
              await walletApi.fiatWithdraw({
                amount: amount.trim(),
                bank_code: bankCode.trim(),
                bank_account: bankAccount.trim(),
                account_name: accountName.trim(),
              });
              dispatch(pushNotification({ type: 'success', message: 'TWD 提領申請已送出，等待人工審核' }));
              navigation.goBack();
            } catch (err: any) {
              const msg = err?.response?.data?.message || '提領申請失敗，請稍後再試';
              dispatch(pushNotification({ type: 'error', message: msg }));
            } finally {
              setSubmitting(false);
            }
          },
        },
      ]
    );
  };

  return (
    <ScrollView style={styles.screen} contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
      <View style={styles.noticeCard}>
        <Text style={styles.noticeText}>• 最低提領金額：{MIN_AMOUNT} TWD</Text>
        <Text style={styles.noticeText}>• 提領申請需人工審核，審核完成後撥款</Text>
        <Text style={styles.noticeText}>• 申請送出後餘額將凍結，審核結果以通知為準</Text>
      </View>

      <Text style={styles.label}>提領金額 (TWD) <Text style={styles.req}>*</Text></Text>
      <TextInput
        style={[styles.input, !!errors.amount && styles.inputError]}
        value={amount}
        onChangeText={setAmount}
        placeholder={`最低 ${MIN_AMOUNT} TWD`}
        placeholderTextColor={colors.textPlaceholder}
        keyboardType="decimal-pad"
      />
      {!!errors.amount && <Text style={styles.errorText}>{errors.amount}</Text>}

      <Text style={[styles.label, { marginTop: 16 }]}>銀行代碼 <Text style={styles.req}>*</Text></Text>
      <TextInput
        style={[styles.input, !!errors.bankCode && styles.inputError]}
        value={bankCode}
        onChangeText={setBankCode}
        placeholder="例：004（台灣銀行）"
        placeholderTextColor={colors.textPlaceholder}
        keyboardType="number-pad"
        maxLength={10}
      />
      {!!errors.bankCode && <Text style={styles.errorText}>{errors.bankCode}</Text>}

      <Text style={[styles.label, { marginTop: 16 }]}>銀行帳號 <Text style={styles.req}>*</Text></Text>
      <TextInput
        style={[styles.input, !!errors.bankAccount && styles.inputError]}
        value={bankAccount}
        onChangeText={setBankAccount}
        placeholder="請輸入完整帳號"
        placeholderTextColor={colors.textPlaceholder}
        keyboardType="number-pad"
      />
      {!!errors.bankAccount && <Text style={styles.errorText}>{errors.bankAccount}</Text>}

      <Text style={[styles.label, { marginTop: 16 }]}>戶名 <Text style={styles.req}>*</Text></Text>
      <TextInput
        style={[styles.input, !!errors.accountName && styles.inputError]}
        value={accountName}
        onChangeText={setAccountName}
        placeholder="請輸入銀行戶名"
        placeholderTextColor={colors.textPlaceholder}
      />
      {!!errors.accountName && <Text style={styles.errorText}>{errors.accountName}</Text>}

      <TouchableOpacity
        style={[styles.submitBtn, submitting && styles.submitBtnDisabled]}
        onPress={handleSubmit}
        disabled={submitting}
        accessibilityRole="button"
      >
        {submitting && <ActivityIndicator size="small" color="#1F2327" style={{ marginRight: 8 }} />}
        <Text style={styles.submitBtnText}>送出提領申請</Text>
      </TouchableOpacity>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bgContent },
  container: { padding: 16, paddingBottom: 32 },

  noticeCard: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    padding: 16,
    marginBottom: 20,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
  noticeText: {
    fontSize: 13,
    color: colors.textSecondary,
    lineHeight: 22,
  },

  label: { fontSize: 12, color: colors.textSecondary, marginBottom: 6 },
  req: { color: colors.danger },

  input: {
    height: 48,
    borderWidth: 1,
    borderColor: colors.borderInput,
    borderRadius: 4,
    paddingHorizontal: 12,
    fontSize: 14,
    color: colors.textPrimary,
    backgroundColor: colors.bgCard,
  },
  inputError: { borderColor: colors.danger },
  errorText: { fontSize: 11, color: colors.danger, marginTop: 4 },

  submitBtn: {
    marginTop: 28,
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
