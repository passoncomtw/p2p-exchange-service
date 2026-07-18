import React, { useEffect, useCallback, useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Pressable,
  StatusBar,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useDispatch } from 'react-redux';
import { useAppSelector } from '@/navigation/store/hooks';
import { fetchWalletsRequest, fetchLedgersRequest } from '@/navigation/store/actions/walletActions';
import { setSelectedCurrency } from '@/navigation/store/slices/walletSlice';
import type { WalletLedgerItem } from '@/interfaces/wallet';

const LEDGER_TYPE_LABEL: Record<string, string> = {
  freeze: '凍結',
  unfreeze: '解凍',
  deposit: '充值',
  transfer_in: '轉入',
  transfer_out: '轉出',
};

const LEDGER_TYPE_COLOR: Record<string, string> = {
  freeze: '#E53935',
  unfreeze: '#43A047',
  deposit: '#43A047',
  transfer_in: '#43A047',
  transfer_out: '#E53935',
};

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
  const color = LEDGER_TYPE_COLOR[item.type] ?? '#333';
  return (
    <View style={styles.ledgerRow}>
      <View style={styles.ledgerLeft}>
        <Text style={styles.ledgerType}>{label}</Text>
        {item.ref_order_no ? (
          <Text style={styles.ledgerOrderNo} numberOfLines={1}>
            {item.ref_order_no}
          </Text>
        ) : null}
        <Text style={styles.ledgerDate}>{formatDate(item.created_at)}</Text>
      </View>
      <View style={styles.ledgerRight}>
        <Text style={[styles.ledgerAmount, { color }]}>{formatAmount(item.amount)}</Text>
        <Text style={styles.ledgerBalance}>餘額 {formatBalance(item.balance_after)}</Text>
      </View>
    </View>
  );
}

export default function WalletScreen() {
  const navigation = useNavigation();
  const dispatch = useDispatch();
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

  const displayBalance = (val: string) => (balanceVisible ? formatBalance(val) : '****');

  const navigateToTrade = (tab: 'buy' | 'sell') => {
    navigation.navigate('Trade', { initialTab: tab });
  };

  const isRefreshing = walletsLoading;

  return (
    <View style={styles.container}>
      <StatusBar barStyle="light-content" />

      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>錢包</Text>
      </View>

      <ScrollView
        style={styles.content}
        showsVerticalScrollIndicator={false}
        refreshControl={<RefreshControl refreshing={isRefreshing} onRefresh={load} tintColor="#fff" />}
      >
        {/* 餘額卡片 */}
        <View style={styles.balanceCard}>
          <View style={styles.balanceHeader}>
            <View style={styles.coinIcon} />
            <Text style={styles.coinLabel}>{selectedCurrency}</Text>
            <Pressable onPress={() => setBalanceVisible((v) => !v)} style={styles.eyeButton}>
              <Text style={styles.eyeIcon}>{balanceVisible ? '👁️' : '👁️‍🗨️'}</Text>
            </Pressable>
          </View>

          {/* 幣種 tabs（多幣種時顯示） */}
          {wallets.length > 1 && (
            <ScrollView horizontal showsHorizontalScrollIndicator={false} style={styles.currencyTabs}>
              {wallets.map((w) => (
                <Pressable
                  key={w.currency}
                  style={[styles.currencyTab, w.currency === selectedCurrency && styles.currencyTabActive]}
                  onPress={() => handleCurrencySelect(w.currency)}
                >
                  <Text
                    style={[
                      styles.currencyTabText,
                      w.currency === selectedCurrency && styles.currencyTabTextActive,
                    ]}
                  >
                    {w.currency}
                  </Text>
                </Pressable>
              ))}
            </ScrollView>
          )}

          <View style={styles.balanceMainRow}>
            <View>
              <Text style={styles.balanceAmount}>
                {currentWallet ? displayBalance(currentWallet.available_balance) : '0.00'}
              </Text>
              <Text style={styles.balanceSubLabel}>可用餘額</Text>
            </View>
            <Pressable style={styles.orderLink} onPress={() => navigation.navigate('Orders')}>
              <Text style={styles.orderLinkText}>查看訂單</Text>
              <Text style={styles.orderLinkArrow}>→</Text>
            </Pressable>
          </View>

          <View style={styles.balanceDivider} />

          <View style={styles.balanceDetails}>
            <View style={styles.balanceItem}>
              <Text style={styles.balanceLabel}>可用餘額</Text>
              <Text style={styles.balanceValue}>
                {currentWallet ? displayBalance(currentWallet.available_balance) : '0.00'}
              </Text>
            </View>
            <View style={styles.balanceItem}>
              <Text style={styles.balanceLabel}>凍結餘額</Text>
              <Text style={styles.balanceValue}>
                {currentWallet ? displayBalance(currentWallet.frozen_balance) : '0.00'}
              </Text>
            </View>
          </View>
        </View>

        {/* 操作列表 */}
        <View style={styles.actionList}>
          <Pressable
            style={({ pressed }) => [styles.actionItem, pressed && styles.actionItemPressed]}
            onPress={() => navigateToTrade('buy')}
          >
            <View style={[styles.actionIcon, { backgroundColor: '#E3F2FD' }]}>
              <Text style={styles.actionIconText}>💰</Text>
            </View>
            <Text style={styles.actionTitle}>我要買</Text>
            <Text style={styles.actionArrow}>›</Text>
          </Pressable>

          <Pressable
            style={({ pressed }) => [styles.actionItem, pressed && styles.actionItemPressed]}
            onPress={() => navigateToTrade('sell')}
          >
            <View style={[styles.actionIcon, { backgroundColor: '#FFF3E0' }]}>
              <Text style={styles.actionIconText}>💸</Text>
            </View>
            <Text style={styles.actionTitle}>我要賣</Text>
            <Text style={styles.actionArrow}>›</Text>
          </Pressable>

          <Pressable
            style={({ pressed }) => [styles.actionItem, pressed && styles.actionItemPressed]}
          >
            <View style={[styles.actionIcon, { backgroundColor: '#F3E5F5' }]}>
              <Text style={styles.actionIconText}>💳</Text>
            </View>
            <Text style={styles.actionTitle}>收付方式</Text>
            <Text style={styles.actionArrow}>›</Text>
          </Pressable>
        </View>

        {/* 帳本記錄 */}
        <View style={styles.ledgerSection}>
          <Text style={styles.ledgerTitle}>交易記錄</Text>

          {ledgersLoading ? (
            <ActivityIndicator style={{ marginTop: 24 }} color="#7B68C8" />
          ) : ledgers.length === 0 ? (
            <Text style={styles.ledgerEmpty}>暫無記錄</Text>
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
    backgroundColor: '#7B68C8',
  },
  topBar: {
    paddingHorizontal: 16,
    paddingTop: 60,
    paddingBottom: 16,
    alignItems: 'center',
  },
  topBarTitle: {
    fontSize: 22,
    fontWeight: 'bold',
    color: '#fff',
  },
  content: {
    flex: 1,
  },
  balanceCard: {
    paddingHorizontal: 24,
    paddingBottom: 24,
  },
  balanceHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 16,
  },
  coinIcon: {
    width: 28,
    height: 28,
    borderRadius: 14,
    backgroundColor: '#FFD700',
    marginRight: 10,
  },
  coinLabel: {
    fontSize: 18,
    fontWeight: 'bold',
    color: '#fff',
  },
  eyeButton: {
    padding: 8,
    marginLeft: 'auto',
  },
  eyeIcon: {
    fontSize: 22,
  },
  currencyTabs: {
    marginBottom: 16,
  },
  currencyTab: {
    paddingHorizontal: 16,
    paddingVertical: 6,
    borderRadius: 20,
    borderWidth: 1,
    borderColor: 'rgba(255,255,255,0.5)',
    marginRight: 8,
  },
  currencyTabActive: {
    backgroundColor: '#fff',
    borderColor: '#fff',
  },
  currencyTabText: {
    color: 'rgba(255,255,255,0.8)',
    fontSize: 14,
    fontWeight: '600',
  },
  currencyTabTextActive: {
    color: '#7B68C8',
  },
  balanceMainRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 20,
  },
  balanceAmount: {
    fontSize: 42,
    fontWeight: 'bold',
    color: '#fff',
    letterSpacing: 0.5,
  },
  balanceSubLabel: {
    fontSize: 13,
    color: 'rgba(255,255,255,0.7)',
    marginTop: 4,
  },
  orderLink: {
    alignItems: 'flex-end',
    paddingTop: 8,
  },
  orderLinkText: {
    fontSize: 13,
    color: '#fff',
    marginBottom: 6,
  },
  orderLinkArrow: {
    fontSize: 28,
    color: '#fff',
    fontWeight: '300',
  },
  balanceDivider: {
    height: 1,
    backgroundColor: 'rgba(255,255,255,0.3)',
    marginVertical: 20,
  },
  balanceDetails: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  balanceItem: {
    flex: 1,
  },
  balanceLabel: {
    fontSize: 13,
    color: 'rgba(255,255,255,0.7)',
    marginBottom: 6,
  },
  balanceValue: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
  },
  actionList: {
    backgroundColor: '#fff',
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    paddingTop: 8,
  },
  actionItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 20,
    paddingVertical: 20,
    backgroundColor: '#fff',
  },
  actionItemPressed: {
    backgroundColor: '#F8F8F8',
  },
  actionIcon: {
    width: 52,
    height: 52,
    borderRadius: 14,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  actionIconText: {
    fontSize: 26,
  },
  actionTitle: {
    flex: 1,
    fontSize: 17,
    fontWeight: '600',
    color: '#333',
  },
  actionArrow: {
    fontSize: 26,
    color: '#CCC',
  },
  ledgerSection: {
    backgroundColor: '#fff',
    paddingHorizontal: 20,
    paddingTop: 20,
    paddingBottom: 40,
  },
  ledgerTitle: {
    fontSize: 17,
    fontWeight: '700',
    color: '#222',
    marginBottom: 12,
  },
  ledgerEmpty: {
    textAlign: 'center',
    color: '#AAA',
    marginTop: 24,
    fontSize: 15,
  },
  ledgerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    paddingVertical: 14,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: '#EFEFEF',
  },
  ledgerLeft: {
    flex: 1,
    marginRight: 12,
  },
  ledgerType: {
    fontSize: 15,
    fontWeight: '600',
    color: '#333',
    marginBottom: 2,
  },
  ledgerOrderNo: {
    fontSize: 12,
    color: '#999',
    marginBottom: 2,
  },
  ledgerDate: {
    fontSize: 12,
    color: '#AAA',
  },
  ledgerRight: {
    alignItems: 'flex-end',
  },
  ledgerAmount: {
    fontSize: 16,
    fontWeight: '700',
    marginBottom: 2,
  },
  ledgerBalance: {
    fontSize: 12,
    color: '#AAA',
  },
});
