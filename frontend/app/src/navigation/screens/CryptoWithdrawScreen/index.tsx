import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TextInput,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  RefreshControl,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '@/navigation/store/hooks';
import { pushNotification } from '@/navigation/store/slices/notificationSlice';
import { fetchWalletsRequest } from '@/navigation/store/actions/walletActions';
import { walletApi } from '@/apis/walletApi';
import { colors } from '@/theme';
import type { CryptoWithdrawalItem } from '@/interfaces/wallet';

const MIN_AMOUNT = 10;

const TRON_ADDRESS_RE = /^T[1-9A-HJ-NP-Za-km-z]{33}$/;

const STATUS_LABEL: Record<string, string> = {
  pending: '待廣播',
  broadcasting: '廣播中',
  confirmed: '已確認',
  failed: '失敗',
};

const STATUS_COLOR: Record<string, string> = {
  pending: colors.textSecondary,
  broadcasting: colors.amberText,
  confirmed: colors.statusCompleted,
  failed: colors.danger,
};

function formatDate(iso: string): string {
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function formatBalance(val: string): string {
  const num = parseFloat(val);
  return isNaN(num) ? '0.000000' : num.toLocaleString('en-US', { minimumFractionDigits: 6, maximumFractionDigits: 6 });
}

function WithdrawalRecordRow({ item }: { item: CryptoWithdrawalItem }) {
  const statusColor = STATUS_COLOR[item.status] ?? colors.textSecondary;
  return (
    <View style={styles.recordRow}>
      <View style={styles.recordLeft}>
        <Text style={[styles.recordStatus, { color: statusColor }]}>{STATUS_LABEL[item.status] ?? item.status}</Text>
        <Text style={styles.recordDate}>{formatDate(item.createdAt)}</Text>
        <Text style={styles.recordAddress} numberOfLines={1}>{item.toAddress}</Text>
      </View>
      <Text style={styles.recordAmount}>-{parseFloat(item.amount).toFixed(6)} USDT</Text>
    </View>
  );
}

export default function CryptoWithdrawScreen() {
  const dispatch = useAppDispatch();
  const navigation = useNavigation<any>();

  const wallets = useAppSelector((state) => state.wallet.wallets);
  const usdtWallet = wallets.find((w) => w.currency === 'USDT');
  const availableBalance = usdtWallet?.available_balance ?? '0';

  const [toAddress, setToAddress] = useState('');
  const [amount, setAmount] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});
  const [records, setRecords] = useState<CryptoWithdrawalItem[]>([]);
  const [recordsLoading, setRecordsLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const loadRecords = useCallback(async (isRefresh = false) => {
    if (isRefresh) setRefreshing(true);
    try {
      const res = await walletApi.listCryptoWithdrawals(20, 0);
      setRecords(res.list);
    } catch {
      // silent fail for records
    } finally {
      setRecordsLoading(false);
      setRefreshing(false);
    }
  }, []);

  useEffect(() => {
    dispatch(fetchWalletsRequest());
    loadRecords();
  }, [dispatch, loadRecords]);

  // 即時地址格式驗證
  const handleAddressChange = (text: string) => {
    setToAddress(text);
    if (text.trim() && !TRON_ADDRESS_RE.test(text.trim())) {
      setErrors((prev) => ({ ...prev, toAddress: '請輸入有效的 Tron (TRC-20) 地址（T 開頭，34 字元）' }));
    } else {
      setErrors((prev) => { const next = { ...prev }; delete next.toAddress; return next; });
    }
  };

  const validate = (): boolean => {
    const errs: Record<string, string> = {};
    if (!toAddress.trim()) {
      errs.toAddress = '請輸入提領地址';
    } else if (!TRON_ADDRESS_RE.test(toAddress.trim())) {
      errs.toAddress = '請輸入有效的 Tron (TRC-20) 地址';
    }
    const num = parseFloat(amount);
    if (!amount.trim() || isNaN(num)) {
      errs.amount = '請輸入有效金額';
    } else if (num < MIN_AMOUNT) {
      errs.amount = `最低提領金額為 ${MIN_AMOUNT} USDT`;
    } else if (num > parseFloat(availableBalance)) {
      errs.amount = '提領金額不可超過可用餘額';
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
              setToAddress('');
              setAmount('');
              dispatch(fetchWalletsRequest());
              loadRecords();
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
    <ScrollView
      style={styles.screen}
      contentContainerStyle={styles.container}
      keyboardShouldPersistTaps="handled"
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => loadRecords(true)} tintColor={colors.primary} />}
    >
      <View style={styles.noticeCard}>
        <Text style={styles.noticeText}>• 最低提領金額：{MIN_AMOUNT} USDT</Text>
        <Text style={styles.noticeText}>• 提領地址須為 Tron（TRC-20）網路</Text>
        <Text style={styles.noticeText}>• 提領申請送出後將凍結餘額，等待鏈上確認</Text>
      </View>

      <View style={styles.balanceRow}>
        <Text style={styles.balanceLabel}>可用餘額</Text>
        <Text style={styles.balanceValue}>{formatBalance(availableBalance)} USDT</Text>
      </View>

      <Text style={styles.label}>提領地址 <Text style={styles.req}>*</Text></Text>
      <TextInput
        style={[styles.input, !!errors.toAddress && styles.inputError]}
        value={toAddress}
        onChangeText={handleAddressChange}
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

      <View style={styles.recordsCard}>
        <Text style={styles.sectionTitle}>提領記錄</Text>
        {recordsLoading ? (
          <ActivityIndicator style={styles.loader} color={colors.primary} />
        ) : records.length === 0 ? (
          <View style={styles.emptyBox}>
            <Text style={styles.emptyText}>暫無提領記錄</Text>
          </View>
        ) : (
          records.map((item) => <WithdrawalRecordRow key={item.id} item={item} />)
        )}
      </View>
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
    marginBottom: 16,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
  noticeText: { fontSize: 13, color: colors.textSecondary, lineHeight: 22 },

  balanceRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    paddingHorizontal: 16,
    paddingVertical: 12,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
  balanceLabel: { fontSize: 13, color: colors.textSecondary },
  balanceValue: { fontSize: 15, fontWeight: '700', color: colors.amberText },

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

  recordsCard: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    padding: 16,
    marginTop: 24,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
  sectionTitle: { fontSize: 14, fontWeight: '700', color: colors.textPrimary, marginBottom: 4 },
  recordRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    paddingVertical: 10,
    borderTopWidth: 1,
    borderTopColor: colors.borderCard,
  },
  recordLeft: { flex: 1, marginRight: 8 },
  recordStatus: { fontSize: 13, fontWeight: '600', marginBottom: 2 },
  recordDate: { fontSize: 11, color: colors.textTertiary, marginBottom: 2 },
  recordAddress: { fontSize: 10, color: colors.textTertiary, fontFamily: 'monospace' },
  recordAmount: { fontSize: 14, fontWeight: '700', color: colors.danger },
  loader: { marginVertical: 24 },
  emptyBox: { alignItems: 'center', paddingVertical: 24 },
  emptyText: { fontSize: 13, color: colors.textTertiary },
});
