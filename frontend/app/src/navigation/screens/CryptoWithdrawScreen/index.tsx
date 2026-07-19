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

const MIN_AMOUNT = 10;

export default function CryptoWithdrawScreen() {
  const dispatch = useAppDispatch();
  const navigation = useNavigation<any>();

  const [toAddress, setToAddress] = useState('');
  const [amount, setAmount] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  const validate = (): boolean => {
    const errs: Record<string, string> = {};
    if (!toAddress.trim()) {
      errs.toAddress = '請輸入提領地址';
    }
    const num = parseFloat(amount);
    if (!amount.trim() || isNaN(num)) {
      errs.amount = '請輸入有效金額';
    } else if (num < MIN_AMOUNT) {
      errs.amount = `最低提領金額為 ${MIN_AMOUNT} USDT`;
    }
    setErrors(errs);
    return Object.keys(errs).length === 0;
  };

  const handleSubmit = () => {
    if (!validate()) return;

    Alert.alert(
      '確認提領',
      `提領 ${amount} USDT 至\n${toAddress}`,
      [
        { text: '取消', style: 'cancel' },
        {
          text: '確認',
          onPress: async () => {
            setSubmitting(true);
            try {
              await walletApi.cryptoWithdraw({ to_address: toAddress.trim(), amount: amount.trim() });
              dispatch(pushNotification({ type: 'success', message: 'USDT 提領申請已送出' }));
              navigation.goBack();
            } catch (err: any) {
              const msg = err?.response?.data?.message || '提領失敗，請稍後再試';
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
        <Text style={styles.noticeText}>• 最低提領金額：{MIN_AMOUNT} USDT</Text>
        <Text style={styles.noticeText}>• 提領地址須為 Tron（TRC-20）網路</Text>
        <Text style={styles.noticeText}>• 提領申請送出後將凍結餘額，等待鏈上確認</Text>
      </View>

      <Text style={styles.label}>提領地址 <Text style={styles.req}>*</Text></Text>
      <TextInput
        style={[styles.input, !!errors.toAddress && styles.inputError]}
        value={toAddress}
        onChangeText={setToAddress}
        placeholder="請輸入 Tron (TRC-20) 錢包地址"
        placeholderTextColor={colors.textPlaceholder}
        autoCapitalize="none"
        autoCorrect={false}
      />
      {!!errors.toAddress && <Text style={styles.errorText}>{errors.toAddress}</Text>}

      <Text style={[styles.label, { marginTop: 16 }]}>提領金額 (USDT) <Text style={styles.req}>*</Text></Text>
      <TextInput
        style={[styles.input, !!errors.amount && styles.inputError]}
        value={amount}
        onChangeText={setAmount}
        placeholder={`最低 ${MIN_AMOUNT} USDT`}
        placeholderTextColor={colors.textPlaceholder}
        keyboardType="decimal-pad"
      />
      {!!errors.amount && <Text style={styles.errorText}>{errors.amount}</Text>}

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
