/**
 * ConfirmOrderScreen - 統一的訂單確認頁面（購買/出售）
 */

import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  Pressable,
  StatusBar,
  TextInput,
  Alert,
  ActivityIndicator,
  Platform,
} from 'react-native';
import { useNavigation, useRoute, RouteProp, useFocusEffect } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '../../store/hooks';
import { fetchBankCardsRequest } from '../../store/actions/bankCardsActions';
import { createOrderRequest } from '../../store/actions/ordersActions';
import logger from '@pkg/logger';

type ConfirmOrderRouteProp = RouteProp<
  {
    ConfirmOrder: {
      type: 'buy' | 'sell';
      orderId: string;
      orderCreatorId?: number; // 建立掛單的使用者 ID
      userName: string;
      availableAmount: number;
      minAmount: number;
      maxAmount: number;
      price: number;
      bankName: string;
      // 銀行卡詳細資訊
      bankCardNumber?: string;
      bankCardHolderName?: string;
      bankBranchName?: string;
    };
  },
  'ConfirmOrder'
>;

export default function ConfirmOrderScreen() {
  const navigation = useNavigation();
  const route = useRoute<ConfirmOrderRouteProp>();
  const dispatch = useAppDispatch();
  const { cards: bankCards } = useAppSelector((state) => state.bankCards);
  const { creatingOrder } = useAppSelector((state) => state.orders);
  const { user } = useAppSelector((state) => state.auth);

  const {
    type,
    orderId,
    orderCreatorId,
    userName,
    availableAmount,
    minAmount,
    maxAmount,
    price = 1.0,
    bankName,
    bankCardNumber,
    bankCardHolderName,
    bankBranchName,
  } = route.params;

  const isBuy = type === 'buy';
  
  // 檢查是否為自己的掛單
  const isOwnOrder = orderCreatorId && user?.id && orderCreatorId === user.id;

  const [amount, setAmount] = useState('');
  const [transactionPassword, setTransactionPassword] = useState('');

  // 當頁面聚焦時，取得銀行卡列表
  useFocusEffect(
    React.useCallback(() => {
      logger.info('ConfirmOrderScreen - 頁面初始化', {
        orderId,
        orderCreatorId,
        currentUserId: user?.id,
        isOwnOrder,
        isBuy,
      });
      dispatch(fetchBankCardsRequest());
    }, [dispatch, orderId, orderCreatorId, user?.id, isOwnOrder, isBuy])
  );

  // 計算交易金額
  const totalPrice = amount ? parseFloat(amount) * price : 0;

  // 格式化數字
  const formatNumber = (num: number) => {
    return num.toLocaleString('zh-TW');
  };

  // 設置全部數量
  const handleSetAll = () => {
    setAmount(maxAmount.toString());
  };

  // 刷新可交易數量
  const handleRefresh = () => {
    logger.info('ConfirmOrderScreen - 刷新可交易數量');
    // TODO: 重新獲取掛單信息
  };

  // 建立訂單成功的回調
  const handleCreateOrderSuccess = React.useCallback((orderId: string) => {
    logger.info('ConfirmOrderScreen - 建立訂單成功', { orderId });
    Alert.alert('成功', `${isBuy ? '購買' : '出售'}訂單已建立！`, [
      { 
        text: '確定', 
        onPress: () => {
          // 返回上一頁
          navigation.goBack();
        } 
      },
    ]);
  }, [isBuy, navigation]);

  // 建立訂單失敗的回調
  const handleCreateOrderError = React.useCallback((errorMessage: string) => {
    Alert.alert('錯誤', errorMessage);
  }, []);

  // 提交訂單
  const handleSubmit = () => {
    if (creatingOrder) return;

    // 驗證
    const amountNum = parseFloat(amount);
    if (!amount || isNaN(amountNum)) {
      Alert.alert('提示', '請輸入數量');
      return;
    }
    if (amountNum < minAmount) {
      Alert.alert('提示', `數量不能小於最小交易量 ${minAmount}`);
      return;
    }
    if (amountNum > maxAmount) {
      Alert.alert('提示', `數量不能大於最大交易量 ${maxAmount}`);
      return;
    }

    // 檢查是否有銀行卡
    if (!bankCards || bankCards.length === 0) {
      Alert.alert('錯誤', '您尚未添加銀行卡，請先到個人設定中添加');
      return;
    }

    // 自動使用第一個（唯一的）銀行卡
    const bankCardToUse = bankCards[0];
    
    if (!transactionPassword) {
      Alert.alert('提示', '請輸入交易密碼');
      return;
    }

    logger.info('ConfirmOrderScreen - 準備建立訂單', {
      type,
      pendingOrderId: orderId,
      amount: amountNum,
      totalPrice,
      beneficiaryBankcardId: bankCardToUse.id,
      isBuy,
      bankCardInfo: {
        id: bankCardToUse.id,
        bankName: bankCardToUse.bank?.bankName,
        cardNumber: bankCardToUse.cardNumber,
      },
    });

    Alert.alert(
      '確認',
      `確定${isBuy ? '購買' : '出售'} ${amountNum} E幣？`,
      [
        { text: '取消', style: 'cancel' },
        {
          text: '確定',
          onPress: () => {
            dispatch(createOrderRequest({
              data: {
                orderId, // 掛單 ID
                amount: amountNum, // 交易金額
                beneficiaryBankcardId: bankCardToUse.id, // 受益人銀行卡 ID
                transactionCode: transactionPassword, // 交易密碼
              },
              onSuccess: handleCreateOrderSuccess,
              onError: handleCreateOrderError,
            }));
          },
        },
      ]
    );
  };

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />

      {/* 頂部導航 */}
      <View style={styles.topBar}>
        <Pressable 
          onPress={() => navigation.goBack()} 
          style={styles.backButton}
          hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
        >
          <Text style={styles.backButtonText}>‹</Text>
        </Pressable>
        <Text style={styles.topBarTitle}>
          {isBuy ? '購買' : '出售'}
        </Text>
        <View style={styles.placeholder} />
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* 可交易數量 */}
        <View style={styles.section}>
          <View style={styles.sectionHeader}>
            <Text style={styles.sectionTitle}>可交易數量</Text>
            <Pressable onPress={handleRefresh} style={styles.refreshButton}>
              <Text style={styles.refreshIcon}>↻</Text>
            </Pressable>
          </View>
          <View style={styles.availableAmountRow}>
            <Text style={styles.coinIcon}>💰</Text>
            <Text style={styles.availableAmountText}>
              {formatNumber(availableAmount)} E幣
            </Text>
          </View>
        </View>

        {/* 數量輸入 */}
        <View style={styles.section}>
          <Text style={styles.label}>數量</Text>
          <View style={styles.inputRow}>
            <View style={styles.inputField}>
              <TextInput
                style={styles.input}
                placeholder={`請輸入${isBuy ? '購買' : '出售'}數量`}
                placeholderTextColor="#999"
                value={amount}
                onChangeText={setAmount}
                keyboardType="numeric"
              />
            </View>
            <Text style={styles.unitText}>E幣</Text>
            <Pressable 
              style={styles.allButton}
              onPress={handleSetAll}
            >
              <Text style={styles.allButtonText}>全部</Text>
            </Pressable>
          </View>
          {amount && parseFloat(amount) > maxAmount && (
            <Text style={styles.errorText}>數量不能大於最大交易量 {maxAmount}</Text>
          )}
          {amount && parseFloat(amount) < minAmount && parseFloat(amount) > 0 && (
            <Text style={styles.errorText}>數量不能小於最小交易量 {minAmount}</Text>
          )}
        </View>

        {/* 交易金額 */}
        <View style={styles.section}>
          <View style={styles.row}>
            <Text style={styles.rowLabel}>交易金額</Text>
            <Text style={styles.priceText}>CNY ¥{formatNumber(totalPrice)}</Text>
          </View>
        </View>

        {/* 收款帳戶 - 僅出售時顯示 */}
        {!isBuy && bankCards && bankCards.length > 0 && (
          <View style={styles.section}>
            <Text style={styles.label}>收款帳戶</Text>
            <View style={styles.bankAccountDisplay}>
              <Text style={styles.selectedAccountIcon}>🏦</Text>
              <View style={styles.selectedAccountDetails}>
                <Text style={styles.selectedAccountType}>
                  {bankCards[0].bank?.bankName}
                </Text>
                <Text style={styles.selectedAccountNumber}>
                  {bankCards[0].cardNumber}
                </Text>
              </View>
            </View>
            <Text style={styles.hint}>買方將以您提供的收款帳戶進行打款</Text>
          </View>
        )}

        {/* 交易密碼 */}
        <View style={styles.section}>
          <Text style={styles.label}>交易密碼</Text>
          <View style={styles.inputField}>
            <TextInput
              style={styles.input}
              placeholder="請輸入交易密碼"
              placeholderTextColor="#999"
              value={transactionPassword}
              onChangeText={setTransactionPassword}
              secureTextEntry
            />
          </View>
        </View>

        {/* 交易信息 */}
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>交易信息</Text>

          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>
              {isBuy ? '賣家暱稱' : '買家暱稱'}
            </Text>
            <Text style={styles.infoValue}>{userName}</Text>
          </View>

          <View style={styles.infoRow}>
            <Text style={styles.infoLabel}>支付方式</Text>
            <Text style={styles.infoValue}>{bankName}</Text>
          </View>

          {/* 賣家收款帳戶資訊 - 僅購買時顯示 */}
          {isBuy && bankCardNumber && (
            <>
              <View style={styles.divider} />
              <Text style={styles.bankCardSectionTitle}>賣家收款帳戶</Text>
              
              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>銀行</Text>
                <Text style={styles.infoValue}>{bankName}</Text>
              </View>

              {bankBranchName && (
                <View style={styles.infoRow}>
                  <Text style={styles.infoLabel}>分行</Text>
                  <Text style={styles.infoValue}>{bankBranchName}</Text>
                </View>
              )}

              <View style={styles.infoRow}>
                <Text style={styles.infoLabel}>卡號</Text>
                <Text style={styles.infoValueMonospace}>{bankCardNumber}</Text>
              </View>

              {bankCardHolderName && (
                <View style={styles.infoRow}>
                  <Text style={styles.infoLabel}>戶名</Text>
                  <Text style={styles.infoValue}>{bankCardHolderName}</Text>
                </View>
              )}
            </>
          )}
        </View>

        {/* 自己的掛單提示 */}
        {isOwnOrder && (
          <View style={styles.ownOrderWarning}>
            <Text style={styles.ownOrderWarningText}>⚠️ 這是您自己的掛單，無法進行交易</Text>
          </View>
        )}

        {/* 提交按鈕 */}
        <Pressable 
          style={({ pressed }) => [
            styles.submitButton,
            pressed && !creatingOrder && !isOwnOrder ? styles.submitButtonPressed : null,
            (creatingOrder || isOwnOrder) ? styles.submitButtonDisabled : null,
          ]}
          onPress={handleSubmit}
          disabled={!!(creatingOrder || isOwnOrder)}
        >
          {creatingOrder ? (
            <ActivityIndicator size="small" color="#fff" />
          ) : (
            <Text style={styles.submitButtonText}>
              {isBuy ? '購買E幣' : '出售E幣'}
            </Text>
          )}
        </Pressable>
      </ScrollView>

    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  topBar: {
    backgroundColor: '#fff',
    paddingHorizontal: 16,
    paddingVertical: 16,
    paddingTop: 60,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5E5',
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
    color: '#007AFF',
    fontWeight: '300',
  },
  topBarTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#000',
  },
  placeholder: {
    width: 44,
  },
  content: {
    flex: 1,
  },
  section: {
    backgroundColor: '#fff',
    padding: 16,
    marginTop: 12,
  },
  sectionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333',
  },
  refreshButton: {
    padding: 4,
  },
  refreshIcon: {
    fontSize: 20,
    color: '#007AFF',
  },
  availableAmountRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  coinIcon: {
    fontSize: 24,
    marginRight: 8,
  },
  availableAmountText: {
    fontSize: 22,
    fontWeight: 'bold',
    color: '#333',
  },
  label: {
    fontSize: 14,
    color: '#333',
    marginBottom: 8,
    fontWeight: '500',
  },
  inputRow: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  inputField: {
    backgroundColor: '#F5F5F5',
    borderRadius: 4,
    height: 48,
    justifyContent: 'center',
    paddingHorizontal: 12,
    flex: 1,
    borderWidth: 1,
    borderColor: '#DDD',
  },
  input: {
    fontSize: 16,
    color: '#333',
  },
  unitText: {
    fontSize: 14,
    color: '#999',
    marginLeft: 12,
  },
  allButton: {
    paddingHorizontal: 12,
    paddingVertical: 8,
    marginLeft: 8,
  },
  allButtonText: {
    fontSize: 14,
    color: '#007AFF',
  },
  errorText: {
    fontSize: 12,
    color: '#F44336',
    marginTop: 4,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  rowLabel: {
    fontSize: 14,
    color: '#333',
  },
  priceText: {
    fontSize: 16,
    color: '#333',
    fontWeight: '500',
  },
  selectField: {
    backgroundColor: '#fff',
    borderWidth: 1,
    borderColor: '#DDD',
    borderRadius: 4,
    height: 48,
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 12,
    justifyContent: 'space-between',
  },
  selectFieldPressed: {
    backgroundColor: '#F8F8F8',
  },
  selectPlaceholder: {
    fontSize: 16,
    color: '#999',
  },
  selectArrow: {
    fontSize: 20,
    color: '#CCC',
  },
  selectedAccountInfo: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
  },
  selectedAccountIcon: {
    fontSize: 20,
    marginRight: 10,
  },
  selectedAccountDetails: {
    flex: 1,
  },
  selectedAccountType: {
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
  },
  selectedAccountNumber: {
    fontSize: 12,
    color: '#666',
    marginTop: 2,
  },
  hint: {
    fontSize: 12,
    color: '#999',
    marginTop: 8,
    lineHeight: 18,
  },
  bankAccountDisplay: {
    backgroundColor: '#F5F5F5',
    borderRadius: 4,
    padding: 12,
    flexDirection: 'row',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#DDD',
  },
  infoRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: 12,
  },
  infoLabel: {
    fontSize: 14,
    color: '#666',
  },
  infoValue: {
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
  },
  infoValueMonospace: {
    fontSize: 14,
    color: '#333',
    fontWeight: '500',
    fontFamily: Platform.OS === 'ios' ? 'Courier' : 'monospace',
  },
  divider: {
    height: 1,
    backgroundColor: '#E5E5E5',
    marginTop: 16,
    marginBottom: 12,
  },
  bankCardSectionTitle: {
    fontSize: 14,
    color: '#007AFF',
    fontWeight: '600',
    marginBottom: 4,
  },
  ownOrderWarning: {
    backgroundColor: '#FFF3E0',
    marginHorizontal: 16,
    marginTop: 12,
    padding: 12,
    borderRadius: 8,
    borderLeftWidth: 4,
    borderLeftColor: '#FF9800',
  },
  ownOrderWarningText: {
    fontSize: 14,
    color: '#E65100',
    textAlign: 'center',
    fontWeight: '500',
  },
  submitButton: {
    backgroundColor: '#333',
    marginHorizontal: 16,
    marginVertical: 20,
    paddingVertical: 16,
    borderRadius: 8,
    alignItems: 'center',
  },
  submitButtonPressed: {
    opacity: 0.9,
  },
  submitButtonDisabled: {
    opacity: 0.6,
  },
  submitButtonText: {
    color: '#fff',
    fontSize: 16,
    fontWeight: '600',
  },
});

