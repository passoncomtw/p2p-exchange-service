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
  Linking,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAppDispatch } from '@/navigation/store/hooks';
import { pushNotification } from '@/navigation/store/slices/notificationSlice';
import { walletApi } from '@/apis/walletApi';
import { colors } from '@/theme';
import type { FiatDepositResponse } from '@/interfaces/wallet';

const MIN_AMOUNT = 100;
const MAX_AMOUNT = 50000;

export default function FiatDepositScreen() {
  const dispatch = useAppDispatch();
  const navigation = useNavigation<any>();

  const [amount, setAmount] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [amountError, setAmountError] = useState('');
  const [paymentResult, setPaymentResult] = useState<FiatDepositResponse | null>(null);

  const validate = (): boolean => {
    const num = parseInt(amount, 10);
    if (!amount.trim() || isNaN(num)) {
      setAmountError('請輸入有效金額');
      return false;
    }
    if (num < MIN_AMOUNT) {
      setAmountError(`最低入金金額為 ${MIN_AMOUNT} TWD`);
      return false;
    }
    if (num > MAX_AMOUNT) {
      setAmountError(`最高入金金額為 ${MAX_AMOUNT} TWD`);
      return false;
    }
    setAmountError('');
    return true;
  };

  const handleSubmit = async () => {
    if (!validate()) return;
    setSubmitting(true);
    try {
      const result = await walletApi.fiatDeposit({ amount: parseInt(amount, 10) });
      setPaymentResult(result);
    } catch (err: any) {
      const msg = err?.response?.data?.message || '建立付款訂單失敗，請稍後再試';
      dispatch(pushNotification({ type: 'error', message: msg }));
    } finally {
      setSubmitting(false);
    }
  };

  const handleOpenPayment = async () => {
    if (!paymentResult) return;
    const url = paymentResult.payment_url;
    const canOpen = await Linking.canOpenURL(url);
    if (canOpen) {
      await Linking.openURL(url);
    } else {
      Alert.alert('無法開啟', '請複製以下連結在瀏覽器中開啟:\n' + url);
    }
  };

  if (paymentResult) {
    return (
      <View style={styles.screen}>
        <ScrollView contentContainerStyle={styles.container}>
          <View style={styles.successCard}>
            <Text style={styles.successTitle}>付款訂單已建立</Text>
            <Text style={styles.successSub}>交易編號：{paymentResult.merchant_trade_no}</Text>
            <Text style={styles.successHint}>
              請點擊下方按鈕前往 ECPay 完成付款，付款後餘額將自動入帳
            </Text>
          </View>

          <TouchableOpacity style={styles.payBtn} onPress={handleOpenPayment} accessibilityRole="button">
            <Text style={styles.payBtnText}>前往付款</Text>
          </TouchableOpacity>

          <TouchableOpacity style={styles.cancelBtn} onPress={() => navigation.goBack()} accessibilityRole="button">
            <Text style={styles.cancelBtnText}>稍後付款</Text>
          </TouchableOpacity>
        </ScrollView>
      </View>
    );
  }

  return (
    <ScrollView style={styles.screen} contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
      <View style={styles.noticeCard}>
        <Text style={styles.noticeText}>• 入金幣種：TWD（新台幣）</Text>
        <Text style={styles.noticeText}>• 入金範圍：{MIN_AMOUNT} ~ {MAX_AMOUNT} TWD</Text>
        <Text style={styles.noticeText}>• 透過 ECPay 信用卡或網路ATM付款</Text>
        <Text style={styles.noticeText}>• 付款後餘額自動入帳，無需額外操作</Text>
      </View>

      <Text style={styles.label}>入金金額 (TWD) <Text style={styles.req}>*</Text></Text>
      <TextInput
        style={[styles.input, !!amountError && styles.inputError]}
        value={amount}
        onChangeText={setAmount}
        placeholder={`${MIN_AMOUNT} ~ ${MAX_AMOUNT}`}
        placeholderTextColor={colors.textPlaceholder}
        keyboardType="number-pad"
      />
      {!!amountError && <Text style={styles.errorText}>{amountError}</Text>}

      <TouchableOpacity
        style={[styles.submitBtn, submitting && styles.submitBtnDisabled]}
        onPress={handleSubmit}
        disabled={submitting}
        accessibilityRole="button"
      >
        {submitting && <ActivityIndicator size="small" color="#1F2327" style={{ marginRight: 8 }} />}
        <Text style={styles.submitBtnText}>建立付款訂單</Text>
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

  successCard: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    padding: 20,
    marginBottom: 24,
    borderWidth: 1,
    borderColor: colors.borderCard,
    alignItems: 'center',
  },
  successTitle: {
    fontSize: 18,
    fontWeight: '700',
    color: colors.textPrimary,
    marginBottom: 8,
  },
  successSub: {
    fontSize: 12,
    color: colors.textTertiary,
    marginBottom: 12,
  },
  successHint: {
    fontSize: 13,
    color: colors.textSecondary,
    textAlign: 'center',
    lineHeight: 20,
  },

  payBtn: {
    height: 48,
    borderRadius: 4,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 12,
  },
  payBtnText: { fontSize: 16, fontWeight: '600', color: '#1F2327' },

  cancelBtn: {
    height: 48,
    borderRadius: 4,
    borderWidth: 1,
    borderColor: colors.borderInput,
    alignItems: 'center',
    justifyContent: 'center',
  },
  cancelBtnText: { fontSize: 15, color: colors.textSecondary },
});
