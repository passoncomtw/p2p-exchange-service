import React, { useEffect, useCallback, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Pressable,
  TouchableOpacity,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { useDispatch } from 'react-redux';
import { useNavigation } from '@react-navigation/native';
import { useAppSelector } from '@/navigation/store/hooks';
import { fetchWalletsRequest, fetchLedgersRequest } from '@/navigation/store/actions/walletActions';
import { setSelectedCurrency } from '@/navigation/store/slices/walletSlice';
import { colors } from '@/theme';
import type { WalletLedgerItem } from '@/interfaces/wallet';

const LEDGER_TYPE_LABEL: Record<string, string> = {
  freeze: '凍結',
  unfreeze: '解凍',
  deposit: '充值',
  transfer_in: '轉入',
  transfer_out: '轉出',
  fee_deduct: '手續費',
  crypto_deposit: '虛擬幣入帳',
  crypto_withdraw: '虛擬幣提領',
  fiat_deposit: '法幣入金',
  fiat_withdraw: '法幣出金',
};

const LEDGER_POSITIVE_TYPES = new Set(['unfreeze', 'deposit', 'transfer_in', 'crypto_deposit', 'fiat_deposit']);

function formatBalance(val: string): string {
  const num = parseFloat(val);
  if (isNaN(num)) return '0.00';
  return num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 });
}

function formatAmount(val: string): string {
  const num = parseFloat(val);
  if (isNaN(num)) return '0';
  const prefix = num >= 0 ? '+' : '';
  return `${prefix}${num.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 6 })}`;
}

function formatDate(iso: string): string {
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function LedgerRow({ item }: { item: WalletLedgerItem }) {
  const label = LEDGER_TYPE_LABEL[item.type] ?? item.type;
  const isPositive = LEDGER_POSITIVE_TYPES.has(item.type);
  const amountColor = isPositive ? colors.statusCompleted : colors.danger;

  return (
    <View style={styles.ledgerRow}>
      <View style={styles.ledgerLeft}>
        <View style={styles.ledgerTypeRow}>
          <View style={[styles.ledgerDot, { backgroundColor: amountColor }]} />
          <Text style={styles.ledgerType}>{label}</Text>
        </View>
        {item.ref_order_no ? (
          <Text style={styles.ledgerOrderNo} numberOfLines={1}>{item.ref_order_no}</Text>
        ) : null}
        <Text style={styles.ledgerDate}>{formatDate(item.created_at)}</Text>
      </View>
      <View style={styles.ledgerRight}>
        <Text style={[styles.ledgerAmount, { color: amountColor }]}>
          {formatAmount(item.amount)}
        </Text>
        <Text style={styles.ledgerBalance}>餘額 {formatBalance(item.balance_after)}</Text>
      </View>
    </View>
  );
}

export default function WalletScreen() {
  const dispatch = useDispatch();
  const navigation = useNavigation<any>();
  const [balanceVisible, setBalanceVisible] = useState(true);

  const { wallets, walletsLoading, ledgers, ledgersLoading, selectedCurrency } = useAppSelector(
    (state) => state.wallet
  );

  const currentWallet = wallets.find((w) => w.currency === selectedCurrency) ?? null;

  const load = useCallback(() => {
    dispatch(fetchWalletsRequest());
    dispatch(fetchLedgersRequest({ currency: selectedCurrency, limit: 30 }));
  }, [dispatch, selectedCurrency]);

  useEffect(() => {
    load();
  }, [load]);

  const handleCurrencySelect = (currency: string) => {
    dispatch(setSelectedCurrency(currency));
    dispatch(fetchLedgersRequest({ currency, limit: 30 }));
  };

  const displayBalance = (val: string) => (balanceVisible ? formatBalance(val) : '••••••');

  return (
    <View style={styles.container}>
      <ScrollView
        showsVerticalScrollIndicator={false}
        refreshControl={
          <RefreshControl refreshing={walletsLoading} onRefresh={load} tintColor={colors.primary} />
        }
      >
        {/* 餘額卡片 */}
        <View style={styles.balanceCard}>
          {/* 幣種切換（多幣種才顯示） */}
          {wallets.length > 1 && (
            <ScrollView horizontal showsHorizontalScrollIndicator={false} style={styles.currencyTabsRow}>
              {wallets.map((w) => (
                <Pressable
                  key={w.currency}
                  style={[styles.currencyTab, w.currency === selectedCurrency && styles.currencyTabActive]}
                  onPress={() => handleCurrencySelect(w.currency)}
                >
                  <Text style={[styles.currencyTabText, w.currency === selectedCurrency && styles.currencyTabTextActive]}>
                    {w.currency}
                  </Text>
                </Pressable>
              ))}
            </ScrollView>
          )}

          {/* 幣種標題列 */}
          <View style={styles.cardHeader}>
            <View style={styles.coinBadge}>
              <Text style={styles.coinBadgeText}>{selectedCurrency}</Text>
            </View>
            <Pressable onPress={() => setBalanceVisible((v) => !v)} style={styles.eyeButton} accessibilityRole="button">
              <Text style={styles.eyeIcon}>{balanceVisible ? '●' : '○'}</Text>
            </Pressable>
          </View>

          {/* 主要餘額 */}
          <View style={styles.mainBalanceBlock}>
            <Text style={styles.mainBalanceLabel}>可用餘額</Text>
            <Text style={styles.mainBalanceAmount}>
              {currentWallet ? displayBalance(currentWallet.available_balance) : '0.00'}
            </Text>
          </View>

          {/* 分隔線 */}
          <View style={styles.divider} />

          {/* 餘額明細 */}
          <View style={styles.balanceDetails}>
            <View style={styles.balanceDetailItem}>
              <Text style={styles.balanceDetailLabel}>可用餘額</Text>
              <Text style={styles.balanceDetailValue}>
                {currentWallet ? displayBalance(currentWallet.available_balance) : '0.00'}
              </Text>
            </View>
            <View style={styles.balanceDetailDivider} />
            <View style={styles.balanceDetailItem}>
              <Text style={styles.balanceDetailLabel}>凍結餘額</Text>
              <Text style={[styles.balanceDetailValue, styles.frozenValue]}>
                {currentWallet ? displayBalance(currentWallet.frozen_balance) : '0.00'}
              </Text>
            </View>
          </View>
        </View>

        {/* USDT 操作按鈕 */}
        {selectedCurrency === 'USDT' && (
          <View style={styles.actionRow}>
            <TouchableOpacity
              style={styles.actionBtn}
              onPress={() => navigation.navigate('CryptoDeposit')}
              accessibilityRole="button"
            >
              <Text style={styles.actionBtnText}>充值</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.actionBtn, styles.actionBtnSecondary]}
              onPress={() => navigation.navigate('CryptoWithdraw')}
              accessibilityRole="button"
            >
              <Text style={[styles.actionBtnText, styles.actionBtnTextSecondary]}>提領</Text>
            </TouchableOpacity>
          </View>
        )}

        {/* TWD 操作按鈕 */}
        {selectedCurrency === 'TWD' && (
          <View style={styles.actionRow}>
            <TouchableOpacity
              style={styles.actionBtn}
              onPress={() => navigation.navigate('FiatDeposit')}
              accessibilityRole="button"
            >
              <Text style={styles.actionBtnText}>入金</Text>
            </TouchableOpacity>
            <TouchableOpacity
              style={[styles.actionBtn, styles.actionBtnSecondary]}
              onPress={() => navigation.navigate('FiatWithdraw')}
              accessibilityRole="button"
            >
              <Text style={[styles.actionBtnText, styles.actionBtnTextSecondary]}>提領</Text>
            </TouchableOpacity>
          </View>
        )}

        {/* 交易記錄 */}
        <View style={styles.ledgerCard}>
          <Text style={styles.sectionTitle}>交易記錄</Text>

          {ledgersLoading ? (
            <ActivityIndicator style={styles.loader} color={colors.primary} />
          ) : ledgers.length === 0 ? (
            <View style={styles.emptyBox}>
              <Text style={styles.emptyText}>暫無記錄</Text>
            </View>
          ) : (
            ledgers.map((item, idx) => <LedgerRow key={idx} item={item} />)
          )}
        </View>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.bgContent,
  },

  // ── 餘額卡片 ──────────────────────────────────────────────────────────────
  balanceCard: {
    backgroundColor: colors.bgCard,
    marginHorizontal: 16,
    marginTop: 16,
    borderRadius: 12,
    padding: 20,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
  currencyTabsRow: {
    marginBottom: 16,
  },
  currencyTab: {
    paddingHorizontal: 14,
    paddingVertical: 5,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: colors.borderCard,
    marginRight: 8,
    backgroundColor: colors.bgContent,
  },
  currencyTabActive: {
    backgroundColor: colors.primary,
    borderColor: colors.primary,
  },
  currencyTabText: {
    fontSize: 12,
    fontWeight: '600',
    color: colors.textSecondary,
  },
  currencyTabTextActive: {
    color: '#1F2327',
  },
  cardHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 20,
  },
  coinBadge: {
    backgroundColor: colors.primary,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 6,
  },
  coinBadgeText: {
    fontSize: 13,
    fontWeight: '700',
    color: '#1F2327',
  },
  eyeButton: {
    marginLeft: 'auto',
    padding: 4,
  },
  eyeIcon: {
    fontSize: 16,
    color: colors.textTertiary,
  },
  mainBalanceBlock: {
    marginBottom: 20,
  },
  mainBalanceLabel: {
    fontSize: 12,
    color: colors.textTertiary,
    marginBottom: 6,
  },
  mainBalanceAmount: {
    fontSize: 36,
    fontWeight: '700',
    color: colors.amberText,
    letterSpacing: 0.5,
  },
  divider: {
    height: 1,
    backgroundColor: colors.borderCard,
    marginBottom: 16,
  },
  balanceDetails: {
    flexDirection: 'row',
  },
  balanceDetailItem: {
    flex: 1,
  },
  balanceDetailDivider: {
    width: 1,
    backgroundColor: colors.borderCard,
    marginHorizontal: 16,
  },
  balanceDetailLabel: {
    fontSize: 12,
    color: colors.textTertiary,
    marginBottom: 4,
  },
  balanceDetailValue: {
    fontSize: 16,
    fontWeight: '600',
    color: colors.textPrimary,
  },
  frozenValue: {
    color: colors.textSecondary,
  },

  // ── 操作按鈕 ──────────────────────────────────────────────────────────────
  actionRow: {
    flexDirection: 'row',
    marginHorizontal: 16,
    marginTop: 12,
    gap: 10,
  },
  actionBtn: {
    flex: 1,
    height: 44,
    borderRadius: 8,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  actionBtnSecondary: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: colors.primary,
  },
  actionBtnText: {
    fontSize: 15,
    fontWeight: '600',
    color: '#1F2327',
  },
  actionBtnTextSecondary: {
    color: colors.primaryDeep,
  },

  // ── 交易記錄 ──────────────────────────────────────────────────────────────
  ledgerCard: {
    backgroundColor: colors.bgCard,
    marginHorizontal: 16,
    marginTop: 12,
    marginBottom: 24,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: colors.borderCard,
    paddingHorizontal: 16,
    paddingTop: 16,
    paddingBottom: 8,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: '700',
    color: colors.textPrimary,
    marginBottom: 4,
  },
  loader: {
    marginVertical: 32,
  },
  emptyBox: {
    alignItems: 'center',
    paddingVertical: 32,
  },
  emptyText: {
    fontSize: 14,
    color: colors.textTertiary,
  },
  ledgerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    paddingVertical: 12,
    borderTopWidth: 1,
    borderTopColor: colors.borderCard,
  },
  ledgerLeft: {
    flex: 1,
    marginRight: 12,
  },
  ledgerTypeRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 3,
  },
  ledgerDot: {
    width: 6,
    height: 6,
    borderRadius: 3,
    marginRight: 6,
  },
  ledgerType: {
    fontSize: 14,
    fontWeight: '600',
    color: colors.textPrimary,
  },
  ledgerOrderNo: {
    fontSize: 11,
    color: colors.textTertiary,
    marginBottom: 2,
    marginLeft: 12,
  },
  ledgerDate: {
    fontSize: 11,
    color: colors.textTertiary,
    marginLeft: 12,
  },
  ledgerRight: {
    alignItems: 'flex-end',
  },
  ledgerAmount: {
    fontSize: 15,
    fontWeight: '700',
    marginBottom: 2,
  },
  ledgerBalance: {
    fontSize: 11,
    color: colors.textTertiary,
  },

  // ── 共用 ──────────────────────────────────────────────────────────────────
  statusCompleted: {
    color: colors.statusCompleted,
  },
});
