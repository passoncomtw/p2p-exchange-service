/**
 * CreateOrderScreen - 建立掛單頁面（共用組件）
 * 支援買幣和賣幣兩種模式
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
  KeyboardAvoidingView,
  Platform,
  Modal,
  TouchableOpacity,
  FlatList,
  ActivityIndicator,
} from 'react-native';
import { useNavigation, useRoute, RouteProp, useFocusEffect } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '../../store/hooks';
import { fetchBankCardsRequest } from '../../store/actions/bankCardsActions';
import { createListingRequest } from '../../store/actions/ordersActions';
import logger from '@pkg/logger';

// 帳戶類型
interface PaymentAccount {
  id: string;
  type: 'bank' | 'alipay' | 'wechat';
  name: string;
  accountNumber: string;
  bankName?: string;
}

type CreateOrderRouteProp = RouteProp<{ 
  CreateOrderBuy: { type: 'buy' };
  CreateOrderSell: { type: 'sell' };
}, 'CreateOrderBuy' | 'CreateOrderSell'>;

export default function CreateOrderScreen() {
  const navigation = useNavigation();
  const route = useRoute<CreateOrderRouteProp>();
  const dispatch = useAppDispatch();
  const { user } = useAppSelector((state) => state.auth);
  const { cards: bankCards, loading: bankCardsLoading } = useAppSelector((state) => state.bankCards);
  const { creating } = useAppSelector((state) => state.orders);
  
  // 從路由名稱判斷是買幣還是賣幣
  const routeName = route.name;
  const isBuy = routeName === 'CreateOrderBuy';
  
  // 從用戶資料獲取可用餘額（賣幣時使用）
  const availableBalance = user?.wallet?.usefulBalance || 0;

  // 當頁面聚焦時，取得銀行卡列表並檢查是否已有掛單
  useFocusEffect(
    React.useCallback(() => {
      if (!isBuy) {
        dispatch(fetchBankCardsRequest());
      }
    }, [dispatch, isBuy])
  );

  // 將銀行卡資料轉換為 PaymentAccount 格式
  const paymentAccounts: PaymentAccount[] = React.useMemo(() => {
    const accounts = bankCards.map((card) => {
      logger.debug('CreateOrderScreen - 轉換銀行卡資料', {
        id: card.id,
        name: card.name,
        cardNumber: card.cardNumber,
        bankName: card.bank.bankName,
      });
      return {
        id: String(card.id),
        type: 'bank' as const,
        name: card.name,
        accountNumber: card.cardNumber,
        bankName: card.bank.bankName,
      };
    });
    
    logger.info('CreateOrderScreen - 銀行卡轉換完成', { count: accounts.length });
    return accounts;
  }, [bankCards]);
  
  const [amount, setAmount] = useState('');
  const [selectedAccount, setSelectedAccount] = useState<PaymentAccount | null>(null);
  const [minAmount, setMinAmount] = useState('');
  const [showAccountModal, setShowAccountModal] = useState(false);
  const [amountError, setAmountError] = useState<string>('');

  // 計算交易金額（假設 1 E幣 = 1 CNY）
  const totalPrice = parseFloat(amount || '0') * 1;

  const formatNumber = (num: number) => {
    return num.toLocaleString('zh-TW');
  };

  // 處理數量輸入變化
  const handleAmountChange = (value: string) => {
    setAmount(value);
    setAmountError('');
    
    // 賣幣時驗證數量不能超過可用餘額
    if (!isBuy && value) {
      const amountNum = parseFloat(value);
      if (!isNaN(amountNum) && amountNum > availableBalance) {
        setAmountError('數量不可超過E幣可用餘額');
      }
    }
  };

  // 點擊「全部」按鈕（僅賣幣時顯示）
  const handleSelectAll = () => {
    if (!isBuy) {
      setAmount(formatNumber(availableBalance).replace(/,/g, ''));
      setAmountError('');
    }
  };

  // 取得帳戶類型圖標
  const getAccountTypeIcon = (type: PaymentAccount['type']) => {
    switch (type) {
      case 'bank': return '🏦';
      case 'alipay': return '💳';
      case 'wechat': return '💬';
      default: return '💰';
    }
  };

  // 取得帳戶類型名稱
  const getAccountTypeName = (type: PaymentAccount['type']) => {
    switch (type) {
      case 'bank': return '銀行卡';
      case 'alipay': return '支付寶';
      case 'wechat': return '微信';
      default: return '其他';
    }
  };

  const handleSelectAccount = () => {
    setShowAccountModal(true);
  };

  const handleAccountSelect = (account: PaymentAccount) => {
    logger.info('CreateOrderScreen - 選擇銀行卡', {
      id: account.id,
      name: account.name,
      accountNumber: account.accountNumber,
      bankName: account.bankName,
    });
    setSelectedAccount(account);
    setShowAccountModal(false);
  };

  // 建立掛單成功的回調
  const handleCreateSuccess = React.useCallback(() => {
    // 重置表單
    setAmount('');
    setSelectedAccount(null);
    setMinAmount('');
    setAmountError('');
    
    Alert.alert('成功', `${isBuy ? '購買' : '出售'}掛單建立成功`, [
      { text: '確定', onPress: () => navigation.goBack() },
    ]);
  }, [isBuy, navigation]);

  // 建立掛單失敗的回調
  const handleCreateError = React.useCallback((errorMessage: string) => {
    Alert.alert('建立失敗', errorMessage);
  }, []);

  const handleSubmit = () => {
    const amountNum = parseFloat(amount);
    if (!amount || isNaN(amountNum) || amountNum <= 0) {
      Alert.alert('錯誤', `請輸入${isBuy ? '購買' : '出售'}數量`);
      return;
    }

    if (!isBuy && amountNum > availableBalance) {
      Alert.alert('錯誤', '數量不可超過E幣可用餘額');
      return;
    }

    if (!isBuy && !selectedAccount) {
      Alert.alert('錯誤', '請選擇收款帳戶');
      return;
    }

    const minAmountNum = parseFloat(minAmount);
    if (!minAmount || isNaN(minAmountNum) || minAmountNum <= 0) {
      Alert.alert('錯誤', '請輸入最小交易量');
      return;
    }

    if (minAmountNum > amountNum) {
      Alert.alert('錯誤', '最小交易量不能大於交易數量');
      return;
    }

    const price = 1;
    const totalAmount = amountNum;
    const maxOrderFiat = totalAmount * price;

    logger.info('CreateOrderScreen - 準備建立掛單', {
      type: isBuy ? 'buy' : 'sell',
      totalAmount,
      price,
      minOrderFiat: minAmountNum,
      maxOrderFiat,
    });

    Alert.alert('確認', `確定建立${isBuy ? '購買' : '出售'}掛單？`, [
      { text: '取消', style: 'cancel' },
      { text: '確定', onPress: () => {
        dispatch(createListingRequest({
          data: {
            type: isBuy ? 'buy' : 'sell',
            totalAmount,
            price,
            minOrderFiat: minAmountNum,
            maxOrderFiat,
            ...((!isBuy && selectedAccount) ? { paymentMethodId: parseInt(selectedAccount.id) } : {}),
          },
          onSuccess: handleCreateSuccess,
          onError: handleCreateError,
        }));
      }},
    ]);
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
          {isBuy ? '掛單/購買' : '掛單/出售'}
        </Text>
        <View style={styles.placeholder} />
      </View>

      <KeyboardAvoidingView 
        style={styles.flex1}
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      >
        <ScrollView 
          style={styles.content} 
          showsVerticalScrollIndicator={false}
          keyboardShouldPersistTaps="handled"
        >
          {/* 主要表單區域 */}
          <View style={styles.section}>
            {/* 數量輸入 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>數量</Text>
              <View style={styles.inputRow}>
                <View style={[
                  styles.inputField,
                  amountError && styles.inputFieldError,
                ]}>
                  <TextInput
                    style={styles.input}
                    placeholder={`請輸入${isBuy ? '購買' : '出售'}數量`}
                    placeholderTextColor="#999"
                    value={amount}
                    onChangeText={handleAmountChange}
                    keyboardType="numeric"
                  />
                </View>
                <Text style={styles.unitText}>E幣</Text>
                {!isBuy && (
                  <Pressable 
                    style={styles.allButton}
                    onPress={handleSelectAll}
                  >
                    <Text style={styles.allButtonText}>全部</Text>
                  </Pressable>
                )}
              </View>
              {amountError && (
                <Text style={styles.errorText}>{amountError}</Text>
              )}
              {!isBuy && (
                <Text style={styles.balanceText}>
                  E幣可用餘額 {formatNumber(availableBalance)}
                </Text>
              )}
            </View>

            {/* 交易金額 */}
            <View style={styles.row}>
              <Text style={styles.rowLabel}>交易金額</Text>
              <Text style={styles.priceText}>CNY ¥{formatNumber(totalPrice)}</Text>
            </View>

            {/* 交易帳戶 - 僅賣幣時顯示 */}
            {!isBuy && (
              <View style={styles.inputGroup}>
                <Text style={styles.label}>收款帳戶</Text>
                <Pressable 
                  style={({ pressed }) => [
                    styles.selectField,
                    pressed && styles.selectFieldPressed,
                  ]}
                  onPress={handleSelectAccount}
                >
                  {selectedAccount ? (
                    <View style={styles.selectedAccountInfo}>
                      <Text style={styles.selectedAccountIcon}>
                        {getAccountTypeIcon(selectedAccount.type)}
                      </Text>
                      <View style={styles.selectedAccountDetails}>
                        <Text style={styles.selectedAccountType}>
                          {selectedAccount.bankName || getAccountTypeName(selectedAccount.type)}
                        </Text>
                        <Text style={styles.selectedAccountNumber}>
                          {selectedAccount.accountNumber}
                        </Text>
                      </View>
                    </View>
                  ) : (
                    <Text style={styles.selectPlaceholder}>請選擇收款帳戶</Text>
                  )}
                  <Text style={styles.selectArrow}>›</Text>
                </Pressable>
                <Text style={styles.hint}>買方將以您提供的收款帳戶進行打款</Text>
              </View>
            )}

            {/* 最小交易量 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>最小交易量</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入最小交易量"
                  placeholderTextColor="#999"
                  value={minAmount}
                  onChangeText={setMinAmount}
                  keyboardType="numeric"
                />
              </View>
            </View>

            {/* 提示訊息 */}
            <View style={styles.warningBox}>
              <Text style={styles.warningText}>掛單時，請務必在線</Text>
            </View>

            {/* 提交按鈕 */}
            <Pressable 
              style={({ pressed }) => [
                styles.submitButton,
                pressed && styles.submitButtonPressed,
                creating && styles.submitButtonDisabled,
              ]}
              onPress={handleSubmit}
              disabled={creating}
            >
              {creating ? (
                <ActivityIndicator size="small" color="#fff" />
              ) : (
                <Text style={styles.submitButtonText}>
                  {isBuy ? '購買E幣' : '出售E幣'}
                </Text>
              )}
            </Pressable>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>

      {/* 帳戶選擇 Modal - 僅賣幣時使用 */}
      <Modal
        visible={showAccountModal}
        transparent
        animationType="slide"
        onRequestClose={() => setShowAccountModal(false)}
      >
        <View style={styles.modalOverlay}>
          <View style={styles.modalContent}>
            <View style={styles.modalHeader}>
              <Text style={styles.modalTitle}>選擇收款帳戶</Text>
              <TouchableOpacity 
                onPress={() => setShowAccountModal(false)}
                style={styles.modalCloseButton}
              >
                <Text style={styles.modalCloseText}>✕</Text>
              </TouchableOpacity>
            </View>
            {bankCardsLoading ? (
              <View style={styles.modalLoading}>
                <Text style={styles.modalLoadingText}>載入中...</Text>
              </View>
            ) : paymentAccounts.length === 0 ? (
              <View style={styles.modalEmpty}>
                <Text style={styles.modalEmptyText}>尚未添加銀行卡</Text>
                <Text style={styles.modalEmptyHint}>請先到個人設定中添加銀行卡</Text>
              </View>
            ) : (
              <FlatList
                data={paymentAccounts}
                keyExtractor={(item) => item.id}
                renderItem={({ item }) => (
                  <TouchableOpacity
                    style={[
                      styles.accountItem,
                      selectedAccount?.id === item.id && styles.accountItemSelected,
                    ]}
                    onPress={() => handleAccountSelect(item)}
                  >
                    <Text style={styles.accountIcon}>
                      {getAccountTypeIcon(item.type)}
                    </Text>
                    <View style={styles.accountInfo}>
                      <Text style={styles.accountType}>
                        {item.bankName || getAccountTypeName(item.type)}
                      </Text>
                      <Text style={styles.accountNumber}>{item.accountNumber}</Text>
                      <Text style={styles.accountName}>{item.name}</Text>
                    </View>
                    {selectedAccount?.id === item.id && (
                      <Text style={styles.accountCheckmark}>✓</Text>
                    )}
                  </TouchableOpacity>
                )}
                ItemSeparatorComponent={() => <View style={styles.accountSeparator} />}
                style={styles.accountList}
              />
            )}
          </View>
        </View>
      </Modal>

    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  flex1: {
    flex: 1,
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
  inputGroup: {
    marginBottom: 20,
  },
  label: {
    fontSize: 14,
    color: '#333',
    marginBottom: 8,
    fontWeight: '500',
  },
  labelRow: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  infoIcon: {
    fontSize: 14,
    color: '#666',
    marginLeft: 4,
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
  inputFieldError: {
    borderColor: '#F44336',
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
  balanceText: {
    fontSize: 12,
    color: '#666',
    marginTop: 4,
  },
  row: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 20,
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
  selectText: {
    fontSize: 16,
    color: '#333',
  },
  selectPlaceholder: {
    fontSize: 16,
    color: '#999',
  },
  selectArrow: {
    fontSize: 20,
    color: '#CCC',
  },
  hint: {
    fontSize: 12,
    color: '#999',
    marginTop: 8,
    lineHeight: 18,
  },
  warningBox: {
    backgroundColor: '#FFF3E0',
    padding: 12,
    borderRadius: 4,
    marginBottom: 20,
  },
  warningText: {
    fontSize: 14,
    color: '#FF9800',
  },
  submitButton: {
    backgroundColor: '#333',
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
  // 選中帳戶樣式
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
  // Modal 樣式
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    justifyContent: 'flex-end',
  },
  modalContent: {
    backgroundColor: '#fff',
    borderTopLeftRadius: 16,
    borderTopRightRadius: 16,
    maxHeight: '70%',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5E5',
  },
  modalTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
  },
  modalCloseButton: {
    width: 32,
    height: 32,
    justifyContent: 'center',
    alignItems: 'center',
  },
  modalCloseText: {
    fontSize: 18,
    color: '#999',
  },
  accountList: {
    paddingHorizontal: 16,
  },
  accountItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 16,
    paddingHorizontal: 12,
    borderRadius: 8,
  },
  accountItemSelected: {
    backgroundColor: '#E3F2FD',
  },
  accountIcon: {
    fontSize: 28,
    marginRight: 12,
  },
  accountInfo: {
    flex: 1,
  },
  accountType: {
    fontSize: 15,
    color: '#333',
    fontWeight: '500',
    marginBottom: 2,
  },
  accountNumber: {
    fontSize: 14,
    color: '#666',
    marginBottom: 2,
  },
  accountName: {
    fontSize: 12,
    color: '#999',
  },
  accountCheckmark: {
    fontSize: 20,
    color: '#007AFF',
    fontWeight: 'bold',
  },
  accountSeparator: {
    height: 1,
    backgroundColor: '#F0F0F0',
  },
  // 選項樣式
  optionItem: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
  },
  optionItemSelected: {
    backgroundColor: '#E3F2FD',
  },
  optionText: {
    fontSize: 16,
    color: '#333',
  },
  optionTextSelected: {
    color: '#007AFF',
    fontWeight: '500',
  },
  optionCheckmark: {
    fontSize: 20,
    color: '#007AFF',
    fontWeight: 'bold',
  },
  optionSeparator: {
    height: 1,
    backgroundColor: '#F0F0F0',
  },
  modalLoading: {
    padding: 40,
    alignItems: 'center',
  },
  modalLoadingText: {
    fontSize: 14,
    color: '#666',
  },
  modalEmpty: {
    padding: 40,
    alignItems: 'center',
  },
  modalEmptyText: {
    fontSize: 16,
    color: '#333',
    marginBottom: 8,
  },
  modalEmptyHint: {
    fontSize: 14,
    color: '#999',
  },
});

