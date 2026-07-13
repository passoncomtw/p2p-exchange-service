import React, { useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Pressable,
  StatusBar,
  Platform,
  ActivityIndicator,
} from 'react-native';
import EmptyContent from './components/EmptyContent';
import OrderContent from './components/OrderContent';
import { useNavigation, useRoute, RouteProp } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '../../store/hooks';
import { fetchOrderDetailRequest } from '../../store/actions/ordersActions';
import { theme, commonStyles } from '@/theme';
import { ORDER_STATUS_MAP } from '@/constants/orders';

type OrderDetailRouteProp = RouteProp<
  {
    OrderDetail: {
      orderId: string;
    };
  },
  'OrderDetail'
>;

export default function OrderDetailScreen() {
  const navigation = useNavigation();
  const route = useRoute<OrderDetailRouteProp>();
  const dispatch = useAppDispatch();
  const orderList = useAppSelector((state) => state.orders.orderList);
  const orderDetailLoading = useAppSelector((state) => state.orders.orderDetailLoading);
  const user = useAppSelector((state) => state.auth.user);

  const { orderId } = route.params;

  useEffect(() => {
    dispatch(fetchOrderDetailRequest({ orderId }));
  }, [orderId, dispatch]);

  const order = React.useMemo(() => {
    return orderList.find((o) => o.id.toString() === orderId);
  }, [orderList, orderId]);

  const statusInfo = ORDER_STATUS_MAP[order?.status ?? 'matched'];

  if (orderDetailLoading && !order) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color={theme.secondary} />
      </View>
    );
  }

  if (!order || !user) {
    return <EmptyContent />;
  }

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />

      {/* 頂部導航欄（白色） */}
      <View style={styles.header}>
        <Pressable onPress={() => navigation.goBack()} style={styles.backButton}>
          <Text style={styles.backButtonText}>‹</Text>
        </Pressable>
        <View style={styles.headerTitleContainer}>
          <Text style={styles.headerTitle}>{statusInfo.label}</Text>
            <View style={[
              styles.orderTypeBadge,styles.orderTypeBuy
            ]}>
              <Text style={styles.orderTypeText}>
                {order.listingType === 'buy' ? '買幣' : '賣幣'}
            </Text>
          </View>
        </View>
        <View style={styles.placeholder} />
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        <OrderContent order={order} user={user} />

        {/* {orderStatus === 0 && needsPayment && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>付款給賣家</Text>
            {timeRemaining !== null && timeRemaining > 0 && (
              <View style={styles.timeRemainingRow}>
                <Text style={styles.timeRemainingLabel}>付款剩餘時間</Text>
                <Text style={[styles.timeRemainingValue, { color: theme.status.error }]}>
                  {formatTime(timeRemaining)}
                </Text>
              </View>
            )}
            <View style={styles.amountRow}>
              <Text style={styles.amountLabel}>付款金額</Text>
              <View style={styles.amountValueRow}>
                <Text style={styles.amountValue}>¥ {order.amount.toLocaleString('zh-TW')}</Text>
                <Pressable
                  onPress={() => copyToClipboard(order.amount.toString(), '付款金額')}
                  style={styles.copyButton}
                >
                  <Text style={styles.copyIcon}>📋</Text>
                </Pressable>
              </View>
            </View>
            <View style={styles.infoRow}>
              <Text style={styles.infoLabel}>數量</Text>
              <Text style={styles.infoValue}>{order.amount.toLocaleString('zh-TW')} E幣</Text>
            </View>
            {bankCards && bankCards.length > 0 && (
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>交易帳戶</Text>
                <Text style={styles.infoValue}>
                  {bankCards[0].bank?.bankName} ({bankCards[0].cardNumber.slice(-4)})
                </Text>
              </View>
            )}
          </View>
        )} */}

        {/* 等待買方付款區塊 - 僅賣方在待付款狀態時顯示 */}
        {/* {orderStatus === 0 && isSeller && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>等待買方付款</Text>
            <View style={styles.amountRow}>
              <Text style={styles.amountLabel}>收款金額</Text>
              <View style={styles.amountValueRow}>
                <Text style={styles.amountValue}>¥ {order.amount.toLocaleString('zh-TW')}</Text>
                <Pressable
                  onPress={() => copyToClipboard(order.amount.toString(), '收款金額')}
                  style={styles.copyButton}
                >
                  <Text style={styles.copyIcon}>📋</Text>
                </Pressable>
              </View>
            </View>
            <View style={styles.infoRow}>
              <Text style={styles.infoLabel}>數量</Text>
              <Text style={styles.infoValue}>{order.amount.toLocaleString('zh-TW')} E幣</Text>
            </View>
            <Text style={styles.hintText}>
              買方正在付款中，請耐心等待。買方付款完成後，訂單狀態將變更為「待放行」。
            </Text>
          </View>
        )} */}

        {/* 等待賣方確認區塊 - 僅買方在待放行狀態時顯示 */}
        {/* {orderStatus === 1 && isBuyer && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>等待賣方確認</Text>
            <View style={styles.amountRow}>
              <Text style={styles.amountLabel}>付款金額</Text>
              <Text style={styles.amountValue}>¥ {order.amount.toLocaleString('zh-TW')}</Text>
            </View>
            <View style={styles.infoRow}>
              <Text style={styles.infoLabel}>數量</Text>
              <Text style={styles.infoValue}>{order.amount.toLocaleString('zh-TW')} E幣</Text>
            </View>
            <Text style={styles.hintText}>
              您已標記為已付款，等待賣方確認收款並放行。確認後訂單將自動完成。
            </Text>
          </View>
        )} */}

        {/* 等待確認放行區塊 - 僅賣方在待放行狀態時顯示 */}
        {/* {orderStatus === 1 && isSeller && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>等待確認放行</Text>
            <View style={styles.amountRow}>
              <Text style={styles.amountLabel}>收款金額</Text>
              <Text style={styles.amountValue}>¥ {order.amount.toLocaleString('zh-TW')}</Text>
            </View>
            <View style={styles.infoRow}>
              <Text style={styles.infoLabel}>數量</Text>
              <Text style={styles.infoValue}>{order.amount.toLocaleString('zh-TW')} E幣</Text>
            </View>
            <Text style={styles.hintText}>
              買方已標記為已付款，請確認您已收到款項。確認後將放行並完成訂單。
            </Text>
          </View>
        )} */}

        {/* 訂單已完成區塊 */}
        {/* {orderStatus === 2 && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>訂單已完成</Text>
            <View style={styles.amountRow}>
              <Text style={styles.amountLabel}>{isBuyer ? '付款金額' : '收款金額'}</Text>
              <Text style={styles.amountValue}>¥ {order.amount.toLocaleString('zh-TW')}</Text>
            </View>
            <View style={styles.infoRow}>
              <Text style={styles.infoLabel}>數量</Text>
              <Text style={styles.infoValue}>{order.amount.toLocaleString('zh-TW')} E幣</Text>
            </View>
            {order.updatedAt && (
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>完成時間</Text>
                <Text style={styles.infoValue}>{formatDateTime(order.updatedAt)}</Text>
              </View>
            )}
            <Text style={styles.hintText}>
              訂單已成功完成，交易已完成。
            </Text>
          </View>
        )} */}

        {/* 收款資訊區塊 - 僅需要付款的用戶在待付款狀態時顯示 */}
        {/* {orderStatus === 0 && needsPayment && sellerInfo && (
          <View style={styles.section}>
            <Pressable
              onPress={() => setShowRecipientInfo(!showRecipientInfo)}
              style={styles.sectionHeader}
            >
              <Text style={styles.sectionTitle}>收款資訊</Text>
              <Text style={styles.collapseIcon}>{showRecipientInfo ? '▲' : '▼'}</Text>
            </Pressable>
            {showRecipientInfo && (
              <>
                <Text style={styles.hintText}>
                  以下為賣方的收款資訊，請使用您選擇的交易帳戶進行轉帳，否則不予以放行
                </Text>
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>賣家姓名</Text>
                <View style={styles.infoValueRow}>
                  <Text style={styles.infoValue}>{sellerInfo.name}</Text>
                  <Pressable
                    onPress={() => copyToClipboard(sellerInfo.name, '賣家姓名')}
                    style={styles.copyButton}
                  >
                    <Text style={styles.copyIcon}>📋</Text>
                  </Pressable>
                </View>
              </View>
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>開戶行</Text>
                <View style={styles.infoValueRow}>
                  <Text style={styles.infoValue}>{sellerInfo.bankName}</Text>
                  <Pressable
                    onPress={() => copyToClipboard(sellerInfo.bankName, '開戶行')}
                    style={styles.copyButton}
                  >
                    <Text style={styles.copyIcon}>📋</Text>
                  </Pressable>
                </View>
              </View>
              {sellerInfo.branchName && (
                <View style={styles.infoRow}>
                  <Text style={styles.infoLabel}>開戶支行</Text>
                  <View style={styles.infoValueRow}>
                    <Text style={styles.infoValue}>{sellerInfo.branchName}</Text>
                    <Pressable
                      onPress={() => copyToClipboard(sellerInfo.branchName, '開戶支行')}
                      style={styles.copyButton}
                    >
                      <Text style={styles.copyIcon}>📋</Text>
                    </Pressable>
                  </View>
                </View>
              )}
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>銀行卡卡號</Text>
                <View style={styles.infoValueRow}>
                  <Text style={styles.infoValueMonospace}>
                    {sellerInfo.cardNumber.replace(/(.{4})/g, '$1 ').trim()}
                  </Text>
                  <Pressable
                    onPress={() => copyToClipboard(sellerInfo.cardNumber, '銀行卡卡號')}
                    style={styles.copyButton}
                  >
                    <Text style={styles.copyIcon}>📋</Text>
                  </Pressable>
                </View>
              </View>
              {sellerInfo.cardHolderName && (
                <View style={styles.infoRow}>
                  <Text style={styles.infoLabel}>戶名</Text>
                  <View style={styles.infoValueRow}>
                    <Text style={styles.infoValue}>{sellerInfo.cardHolderName}</Text>
                    <Pressable
                      onPress={() => copyToClipboard(sellerInfo.cardHolderName, '戶名')}
                      style={styles.copyButton}
                    >
                      <Text style={styles.copyIcon}>📋</Text>
                    </Pressable>
                  </View>
                </View>
              )}
              </>
            )}
          </View>
        )} */}

        {/* 訂單資訊區塊（可摺疊） */}
        {/* <View style={styles.section}>
          <Pressable
            onPress={() => setShowOrderInfo(!showOrderInfo)}
            style={styles.sectionHeader}
          >
            <Text style={styles.sectionTitle}>訂單資訊</Text>
            <Text style={styles.collapseIcon}>{showOrderInfo ? '▲' : '▼'}</Text>
          </Pressable>
          {showOrderInfo && (
            <>
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>訂單編號</Text>
                <View style={styles.infoValueRow}>
                  <Text style={styles.infoValueMonospace}>{order.id}</Text>
                  <Pressable
                    onPress={() => copyToClipboard(order.id, '訂單編號')}
                    style={styles.copyButton}
                  >
                    <Text style={styles.copyIcon}>📋</Text>
                  </Pressable>
                </View>
              </View>
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>訂單成立時間</Text>
                <View style={styles.infoValueRow}>
                  <Text style={styles.infoValue}>{formatDateTime(order.createdAt)}</Text>
                  <Pressable
                    onPress={() => copyToClipboard(formatDateTime(order.createdAt), '訂單成立時間')}
                    style={styles.copyButton}
                  >
                    <Text style={styles.copyIcon}>📋</Text>
                  </Pressable>
                </View>
              </View>
            </>
          )}
        </View> */}
      </ScrollView>

      {/* 底部按鈕 */}
      {/* 需要付款的用戶：待付款狀態 - 顯示「我已付款」和「取消訂單」 */}
      {/* {orderStatus === 0 && needsPayment && (
        <View style={styles.buttonContainer}>
          <Pressable
            style={[styles.buttonPrimary, loading && styles.buttonDisabled]}
            onPress={handleMarkAsPaid}
            disabled={loading}
          >
            {loading ? (
              <ActivityIndicator size="small" color={theme.background.primary} />
            ) : (
              <Text style={styles.buttonPrimaryText}>我已付款</Text>
            )}
          </Pressable>
          <Pressable
            style={[styles.buttonSecondary, loading && styles.buttonDisabled]}
            onPress={handleCancelOrder}
            disabled={loading}
          >
            <Text style={styles.buttonSecondaryText}>取消訂單</Text>
          </Pressable>
        </View>
      )} */}

      {/* 賣方：待放行狀態 - 顯示「確認放行」 */}
      {/* {orderStatus === 1 && isSeller && (
        <View style={styles.buttonContainer}>
          <Pressable
            style={[styles.buttonPrimary, loading && styles.buttonDisabled]}
            onPress={handleConfirmRelease}
            disabled={loading}
          >
            {loading ? (
              <ActivityIndicator size="small" color={theme.background.primary} />
            ) : (
              <Text style={styles.buttonPrimaryText}>確認放行</Text>
            )}
          </Pressable>
        </View>
      )} */}
    </View>
  );
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

