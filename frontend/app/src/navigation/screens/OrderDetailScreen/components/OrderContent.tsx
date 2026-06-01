import React from 'react';
import { View, Text, Pressable, StyleSheet, Platform, Alert } from 'react-native';
import { theme, commonStyles } from '@/theme';
import { formatDateTime } from '@/utils/formatUtils';
import { OrderItem } from '@/interfaces';
import { ORDER_STATUS_MAP } from '@/constants/orders';
import { User } from '@/interfaces/store';
import { useAppDispatch } from '@/navigation/store/hooks';
import type { AppDispatch } from '@/navigation/store/configureStore';
import { markOrderAsPaidRequest, applyOrderRequest } from '@/navigation/store/actions/ordersActions';

const handleMarkAsPaid = (orderId: string, dispatch: AppDispatch) => () => {
  Alert.alert(
    '確認付款',
    '確認已完成匯款？',
    [
      {
        text: '取消',
        style: 'cancel',
      },
      {
        text: '確認',
        onPress: () => {
          dispatch(markOrderAsPaidRequest({
            orderId,
            onSuccess: () => {
              Alert.alert('成功', '已標記為已付款');
            },
            onError: (error) => {
              Alert.alert('錯誤', error || '標記付款失敗');
            },
          }));
        },
      },
    ]
  );
};

const handleApplyOrder = (orderId: string, dispatch: AppDispatch) => () => {
  Alert.alert(
    '確認放行',
    '確認已收到款項並放行訂單？',
    [
      {
        text: '取消',
        style: 'cancel',
      },
      {
        text: '確認',
        onPress: () => {
          dispatch(applyOrderRequest({
            orderId,
            onSuccess: () => {
              Alert.alert('成功', '訂單已放行');
            },
            onError: (error) => {
              Alert.alert('錯誤', error || '放行訂單失敗');
            },
          }));
        },
      },
    ]
  );
};

const Footer = (props: { order: OrderItem, user: User; dispatch: AppDispatch }) => {
  const { order, user, dispatch } = props;
  const isBuyPendingOrder = order.pendingOrder.type === 0;

  if (order.pendingOrder.type === 0) {
    if (user.id === order.user.id) {
      if (order.status === 0) {
        return (<View style={styles.buttonContainer}>
          <Pressable style={styles.buttonPrimary} onPress={handleMarkAsPaid(order.id, dispatch)}>
            <Text style={styles.buttonPrimaryText}>匯款已完成</Text>
          </Pressable>
        </View>)
      }
      return (<View style={styles.infoRow}>
        <Text style={styles.infoLabel}>等待確認放行</Text>
      </View>)
    } else {
      if (order.status === 1) {
        return (<View style={styles.buttonContainer}>
          <Pressable style={styles.buttonPrimary} onPress={handleApplyOrder(order.id, dispatch)}>
            <Text style={styles.buttonPrimaryText}>確認放行</Text>
          </Pressable>
        </View>)
      }
      return (<View style={styles.infoRow}>
        <Text style={styles.infoLabel}>等待其他動作</Text>
      </View>)
    }
  }
  if (order.status === 0 && isBuyPendingOrder && user.id === order.user.id) {
    return (<View style={styles.buttonContainer}>
      <Pressable style={styles.buttonPrimary} onPress={handleMarkAsPaid(order.id, dispatch)}>
        <Text style={styles.buttonPrimaryText}>匯款已完成</Text>
      </Pressable>
    </View>)
  }

  return (
    <View style={styles.section}>
      <View style={styles.infoRow}>
        <Text style={styles.infoLabel}>等待對方完成付款動作</Text>
      </View>
    </View>
  )
}
const OrderContent = (props: { order: OrderItem, user: User }) => {
  const { order, user } = props;
  const dispatch = useAppDispatch();

  const isBuyPendingOrder = order.pendingOrder.type === 0;

  return (
    <View>
      <View style={styles.section}>
        <Pressable
          onPress={() => false}
          style={styles.sectionHeader}
        >
        </Pressable>
        <>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>掛單資訊</Text>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>掛單編號</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{order.pendingOrder.id}</Text>
            </View>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>掛單類型</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{isBuyPendingOrder ? '買幣掛單' : '賣幣掛單'}</Text>
            </View>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>掛單建立者</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{order.pendingOrder.user.name}</Text>
            </View>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>掛單成立時間</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{formatDateTime(order.pendingOrder.createdAt)}</Text>
            </View>
          </View>
        </>
      </View>
      <View style={styles.section}>
        <>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>訂單資訊</Text>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>訂單編號</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{order.id}</Text>
            </View>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>訂單狀態</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{ORDER_STATUS_MAP[order.status].label}</Text>
            </View>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>訂單建立者</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{order.user.name}</Text>
            </View>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>訂單成立時間</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{formatDateTime(order.createdAt)}</Text>
            </View>
          </View>
        </>
      </View>
      <View style={styles.section}>
        <>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>交易資料</Text>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>{isBuyPendingOrder ? "買幣掛單 匯款給訂單使用者" : "賣幣掛單 匯款給掛單使用者"}</Text>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>交易金額</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{order.amount}</Text>
            </View>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>銀行名稱</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{isBuyPendingOrder ? order.bankcard.bank.bankName : order.pendingOrder.bankcard.bank.bankName}</Text>
            </View>
          </View>
          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>銀行帳號</Text>
            <View style={styles.infoValueRow}>
              <Text style={styles.infoValue}>{isBuyPendingOrder ? order.bankcard.cardNumber : order.pendingOrder.bankcard.cardNumber}</Text>
            </View>
          </View>
        </>
      </View>

      <Footer order={order} user={user} dispatch={dispatch} />
    </View>
  )
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: theme.background.secondary,
  },
  header: {
    backgroundColor: '#fff',
    paddingHorizontal: theme.spacing.lg,
    paddingTop: 60,
    paddingBottom: theme.spacing.lg,
    borderBottomWidth: 1,
    borderBottomColor: theme.border.default,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  backButton: {
    width: 44,
    height: 44,
    justifyContent: 'center',
    alignItems: 'center',
  },
  backButtonText: {
    fontSize: 32,
    color: theme.secondary,
    fontWeight: '300',
  },
  headerTitleContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: theme.spacing.sm,
  },
  headerTitle: {
    fontSize: theme.fontSize.xl,
    fontWeight: '600',
    color: theme.secondary,
  },
  orderTypeBadge: {
    paddingHorizontal: theme.spacing.sm,
    paddingVertical: theme.spacing.xs,
    borderRadius: theme.radius.sm,
  },
  orderTypeBuy: {
    backgroundColor: theme.status.info,
  },
  orderTypeSell: {
    backgroundColor: theme.status.warning,
  },
  orderTypeText: {
    fontSize: theme.fontSize.xs,
    fontWeight: '600',
    color: theme.background.primary,
  },
  placeholder: {
    width: 44,
  },
  content: {
    flex: 1,
  },
  section: {
    backgroundColor: theme.background.primary,
    padding: theme.spacing.lg,
    marginTop: theme.spacing.md,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: theme.spacing.md,
  },
  sectionTitle: {
    fontSize: theme.fontSize.lg,
    fontWeight: '600',
    color: theme.text.primary,
  },
  collapseIcon: {
    fontSize: theme.fontSize.md,
    color: theme.text.tertiary,
  },
  timeRemainingRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: theme.spacing.lg,
    paddingVertical: theme.spacing.md,
    paddingHorizontal: theme.spacing.md,
    backgroundColor: theme.background.tertiary,
    borderRadius: theme.radius.lg,
  },
  timeRemainingLabel: {
    fontSize: theme.fontSize.md,
    color: theme.text.secondary,
  },
  timeRemainingValue: {
    fontSize: theme.fontSize.xl,
    fontWeight: '600',
  },
  amountRow: {
    marginBottom: theme.spacing.md,
  },
  amountLabel: {
    fontSize: theme.fontSize.md,
    color: theme.text.secondary,
    marginBottom: theme.spacing.sm,
  },
  amountValueRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  amountValue: {
    fontSize: theme.fontSize.xxl,
    fontWeight: '600',
    color: theme.text.primary,
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: theme.spacing.md,
  },
  infoLabel: {
    fontSize: theme.fontSize.md,
    color: theme.text.secondary,
  },
  infoValue: {
    fontSize: theme.fontSize.md,
    color: theme.text.primary,
    fontWeight: '500',
  },
  infoValueMonospace: {
    fontSize: theme.fontSize.md,
    color: theme.text.primary,
    fontWeight: '500',
    fontFamily: Platform.OS === 'ios' ? 'Courier' : 'monospace',
  },
  infoValueRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  copyButton: {
    marginLeft: theme.spacing.sm,
    padding: theme.spacing.xs,
  },
  copyIcon: {
    fontSize: theme.fontSize.md,
  },
  hintText: {
    fontSize: theme.fontSize.sm,
    color: theme.text.tertiary,
    marginBottom: theme.spacing.md,
    lineHeight: 20,
  },
  buttonContainer: {
    padding: theme.spacing.lg,
    backgroundColor: theme.background.primary,
    borderTopWidth: 1,
    borderTopColor: theme.border.default,
  },
  buttonPrimary: {
    ...commonStyles.buttonPrimary,
    marginBottom: theme.spacing.md,
  },
  buttonPrimaryText: {
    ...commonStyles.buttonPrimaryText,
  },
  buttonSecondary: {
    backgroundColor: theme.background.secondary,
    height: 48,
    borderRadius: theme.radius.xxl,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: 1,
    borderColor: theme.border.default,
  },
  buttonSecondaryText: {
    fontSize: theme.fontSize.lg,
    fontWeight: '600',
    color: theme.text.primary,
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  loadingText: {
    marginTop: theme.spacing.md,
    fontSize: theme.fontSize.md,
    color: theme.text.secondary,
  },
});
export default OrderContent;