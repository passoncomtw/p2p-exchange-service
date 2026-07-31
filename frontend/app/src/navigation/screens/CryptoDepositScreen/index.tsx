import React, { useEffect, useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Clipboard,
  RefreshControl,
} from 'react-native';
import QRCode from 'react-native-qrcode-svg';
import { useAppDispatch } from '@/navigation/store/hooks';
import { pushNotification } from '@/navigation/store/slices/notificationSlice';
import { walletApi } from '@/apis/walletApi';
import { colors } from '@/theme';
import type { CryptoDepositInfo, CryptoDepositItem } from '@/interfaces/wallet';

const STATUS_LABEL: Record<string, string> = {
  pending: '確認中',
  confirmed: '已到帳',
  failed: '失敗',
};

const STATUS_COLOR: Record<string, string> = {
  pending: colors.textSecondary,
  confirmed: colors.statusCompleted,
  failed: colors.danger,
};

function formatDate(iso: string): string {
  const d = new Date(iso);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

function CopyRow({ label, value }: { label: string; value: string }) {
  const dispatch = useAppDispatch();

  const handleCopy = () => {
    Clipboard.setString(value);
    dispatch(pushNotification({ type: 'success', message: `${label}已複製` }));
  };

  return (
    <View style={styles.copyRow}>
      <View style={styles.copyRowContent}>
        <Text style={styles.copyLabel}>{label}</Text>
        <Text style={styles.copyValue} selectable>{value}</Text>
      </View>
      <TouchableOpacity style={styles.copyBtn} onPress={handleCopy} accessibilityRole="button">
        <Text style={styles.copyBtnText}>複製</Text>
      </TouchableOpacity>
    </View>
  );
}

function DepositRecordRow({ item }: { item: CryptoDepositItem }) {
  const statusColor = STATUS_COLOR[item.status] ?? colors.textSecondary;
  return (
    <View style={styles.recordRow}>
      <View style={styles.recordLeft}>
        <Text style={[styles.recordStatus, { color: statusColor }]}>{STATUS_LABEL[item.status] ?? item.status}</Text>
        <Text style={styles.recordDate}>{formatDate(item.createdAt)}</Text>
        <Text style={styles.recordTxHash} numberOfLines={1}>{item.txHash}</Text>
      </View>
      <Text style={styles.recordAmount}>+{parseFloat(item.amount).toFixed(6)} USDT</Text>
    </View>
  );
}

export default function CryptoDepositScreen() {
  const dispatch = useAppDispatch();
  const [info, setInfo] = useState<CryptoDepositInfo | null>(null);
  const [records, setRecords] = useState<CryptoDepositItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);

  const load = useCallback(async (isRefresh = false) => {
    if (isRefresh) setRefreshing(true);
    try {
      const [infoRes, listRes] = await Promise.all([
        walletApi.getCryptoDepositInfo(),
        walletApi.listCryptoDeposits(20, 0),
      ]);
      setInfo(infoRes);
      setRecords(listRes.list);
    } catch {
      dispatch(pushNotification({ type: 'error', message: '取得充值資訊失敗，請稍後再試' }));
    } finally {
      setLoading(false);
      setRefreshing(false);
    }
  }, [dispatch]);

  useEffect(() => { load(); }, [load]);

  if (loading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator color={colors.primary} size="large" />
      </View>
    );
  }

  if (!info) {
    return (
      <View style={styles.center}>
        <Text style={styles.errorText}>無法取得充值資訊</Text>
      </View>
    );
  }

  const qrValue = `tron:${info.address}?memo=${info.memo}`;

  return (
    <ScrollView
      style={styles.screen}
      contentContainerStyle={styles.container}
      refreshControl={<RefreshControl refreshing={refreshing} onRefresh={() => load(true)} tintColor={colors.primary} />}
    >
      <View style={styles.noticeCard}>
        <Text style={styles.noticeTitle}>充值說明</Text>
        <Text style={styles.noticeText}>• 網路：{info.network}（TRC-20）</Text>
        <Text style={styles.noticeText}>• 幣種：{info.currency}</Text>
        <Text style={styles.noticeText}>• 轉帳時請務必填寫備註（Memo），否則無法到帳</Text>
        <Text style={styles.noticeText}>• 需等待 6 個區塊確認後才會入帳</Text>
      </View>

      <View style={styles.infoCard}>
        <Text style={styles.cardTitle}>充值地址</Text>
        <View style={styles.qrContainer}>
          <QRCode value={qrValue} size={160} backgroundColor="#FFFFFF" color="#000000" />
        </View>
        <CopyRow label="收款地址" value={info.address} />
        <View style={styles.separator} />
        <CopyRow label="備註 (Memo)" value={info.memo} />
      </View>

      <View style={styles.warningCard}>
        <Text style={styles.warningText}>
          ⚠ 請確認您轉入的是 {info.network} 網路的 {info.currency}，其他幣種或網路將無法找回
        </Text>
      </View>

      <View style={styles.recordsCard}>
        <Text style={styles.cardTitle}>充值記錄</Text>
        {records.length === 0 ? (
          <View style={styles.emptyBox}>
            <Text style={styles.emptyText}>暫無充值記錄</Text>
          </View>
        ) : (
          records.map((item) => <DepositRecordRow key={item.id} item={item} />)
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.bgContent },
  container: { padding: 16, paddingBottom: 32 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', backgroundColor: colors.bgContent },
  errorText: { fontSize: 14, color: colors.textSecondary },

  noticeCard: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
  noticeTitle: { fontSize: 14, fontWeight: '700', color: colors.textPrimary, marginBottom: 8 },
  noticeText: { fontSize: 13, color: colors.textSecondary, lineHeight: 22 },

  infoCard: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
  cardTitle: { fontSize: 14, fontWeight: '700', color: colors.textPrimary, marginBottom: 12 },

  qrContainer: {
    alignItems: 'center',
    padding: 12,
    backgroundColor: '#FFFFFF',
    borderRadius: 8,
    marginBottom: 16,
    alignSelf: 'center',
  },

  copyRow: { flexDirection: 'row', alignItems: 'center' },
  copyRowContent: { flex: 1, marginRight: 8 },
  copyLabel: { fontSize: 11, color: colors.textTertiary, marginBottom: 2 },
  copyValue: { fontSize: 13, color: colors.textPrimary, fontFamily: 'monospace' },
  copyBtn: { paddingHorizontal: 12, paddingVertical: 6, backgroundColor: colors.primary, borderRadius: 4 },
  copyBtnText: { fontSize: 12, fontWeight: '600', color: '#1F2327' },
  separator: { height: 1, backgroundColor: colors.borderCard, marginVertical: 12 },

  warningCard: {
    backgroundColor: colors.amberBg,
    borderRadius: 8,
    padding: 12,
    borderWidth: 1,
    borderColor: colors.amberBorder,
    marginBottom: 12,
  },
  warningText: { fontSize: 12, color: colors.amberText, lineHeight: 20 },

  recordsCard: {
    backgroundColor: colors.bgCard,
    borderRadius: 8,
    padding: 16,
    borderWidth: 1,
    borderColor: colors.borderCard,
  },
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
  recordTxHash: { fontSize: 10, color: colors.textTertiary, fontFamily: 'monospace' },
  recordAmount: { fontSize: 14, fontWeight: '700', color: colors.statusCompleted },
  emptyBox: { alignItems: 'center', paddingVertical: 24 },
  emptyText: { fontSize: 13, color: colors.textTertiary },
});
