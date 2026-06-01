/**
 * WalletScreen - 錢包頁面
 */

import React, { useState } from 'react';
import { View, Text, StyleSheet, ScrollView, Pressable, StatusBar } from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAppSelector } from '@/navigation/store/hooks';

export default function WalletScreen() {
  const navigation = useNavigation();
  const { user } = useAppSelector((state) => state.auth);
  const [balanceVisible, setBalanceVisible] = useState(true);

  // 從用戶資料獲取錢包餘額
  const usefulBalance = user?.wallet?.usefulBalance || 0;
  const guaranteedBalance = user?.wallet?.guaranteedBalance || 0;
  const totalBalance = usefulBalance + guaranteedBalance;

  // 格式化數字顯示（千分位）
  const formatNumber = (num: number) => {
    return num.toLocaleString('zh-TW');
  };

  // 切換餘額顯示/隱藏
  const toggleBalanceVisibility = () => {
    setBalanceVisible(!balanceVisible);
  };

  // 顯示餘額或隱藏為 ****
  const displayBalance = (amount: number) => {
    if (balanceVisible) {
      return formatNumber(amount);
    }
    return '****';
  };

  // 導航到交易頁面並設置 tab
  const navigateToTrade = (tab: 'buy' | 'sell') => {
    navigation.navigate('Trade', { initialTab: tab });
  };

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />
      
      {/* 頂部標題 */}
      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>錢包</Text>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* 餘額卡片 */}
        <View style={styles.balanceCard}>
          {/* 頂部：E幣標題 + 眼睛按鈕 */}
          <View style={styles.balanceHeader}>
            <View style={styles.coinIcon} />
            <Text style={styles.coinLabel}>E幣</Text>
            <Pressable 
              onPress={toggleBalanceVisibility}
              style={styles.eyeButton}
            >
              <Text style={styles.eyeIcon}>{balanceVisible ? '👁️' : '👁️‍🗨️'}</Text>
            </Pressable>
          </View>

          {/* 主要餘額區域 */}
          <View style={styles.balanceMainRow}>
            <View>
              <Text style={styles.balanceAmount}>
                {displayBalance(totalBalance)}
              </Text>
              <Text style={styles.balanceCny}>
                = CNY ¥{displayBalance(totalBalance)}
              </Text>
            </View>
            <Pressable 
              style={styles.orderLink}
              onPress={() => navigation.navigate('OrderList')}
            >
              <Text style={styles.orderLinkText}>查看訂單</Text>
              <Text style={styles.orderLinkArrow}>→</Text>
            </Pressable>
          </View>

          {/* 分隔線 */}
          <View style={styles.balanceDivider} />

          {/* 餘額詳情 */}
          <View style={styles.balanceDetails}>
            <View style={styles.balanceItem}>
              <Text style={styles.balanceLabel}>可用餘額 (E幣)</Text>
              <Text style={styles.balanceValue}>
                {displayBalance(usefulBalance)}
              </Text>
            </View>
            <View style={styles.balanceItem}>
              <Text style={styles.balanceLabel}>擔保餘額 (E幣)</Text>
              <Text style={styles.balanceValue}>
                {displayBalance(guaranteedBalance)}
              </Text>
            </View>
          </View>
        </View>

        {/* 操作列表 */}
        <View style={styles.actionList}>
          <Pressable 
            style={({ pressed }) => [
              styles.actionItem,
              pressed && styles.actionItemPressed,
            ]}
            onPress={() => navigateToTrade('buy')}
          >
            <View style={[styles.actionIcon, { backgroundColor: '#E3F2FD' }]}>
              <Text style={styles.actionIconText}>💰</Text>
            </View>
            <View style={styles.actionContent}>
              <Text style={styles.actionTitle}>我要買</Text>
            </View>
            <Text style={styles.actionArrow}>›</Text>
          </Pressable>

          <Pressable 
            style={({ pressed }) => [
              styles.actionItem,
              pressed && styles.actionItemPressed,
            ]}
            onPress={() => navigateToTrade('sell')}
          >
            <View style={[styles.actionIcon, { backgroundColor: '#FFF3E0' }]}>
              <Text style={styles.actionIconText}>💸</Text>
            </View>
            <View style={styles.actionContent}>
              <Text style={styles.actionTitle}>我要賣</Text>
            </View>
            <Text style={styles.actionArrow}>›</Text>
          </Pressable>

          <Pressable 
            style={({ pressed }) => [
              styles.actionItem,
              pressed && styles.actionItemPressed,
            ]}
          >
            <View style={[styles.actionIcon, { backgroundColor: '#F3E5F5' }]}>
              <Text style={styles.actionIconText}>💳</Text>
            </View>
            <View style={styles.actionContent}>
              <Text style={styles.actionTitle}>收付方式</Text>
            </View>
            <Text style={styles.actionArrow}>›</Text>
          </Pressable>
        </View>
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#7B68C8', // 紫色背景
    paddingLeft: 16,
    paddingRight: 16,
  },
  topBar: {
    backgroundColor: '#7B68C8', // 紫色背景
    paddingHorizontal: 16,
    paddingVertical: 16,
    paddingTop: 60,
    alignItems: 'center',
  },
  topBarTitle: {
    fontSize: 22,
    fontWeight: 'bold',
    color: '#fff', // 白色文字
  },
  content: {
    flex: 1,
  },
  balanceCard: {
    backgroundColor: '#7B68C8', // 紫色背景
    borderRadius: 0, // 無圓角
    padding: 24,
    paddingTop: 32,
    marginBottom: 0,
  },
  balanceHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 40,
  },
  coinIcon: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#FFD700',
    marginRight: 12,
  },
  coinLabel: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#fff',
    marginRight: 8,
  },
  eyeButton: {
    padding: 8,
    marginLeft: 'auto',
  },
  eyeIcon: {
    fontSize: 24,
  },
  balanceMainRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 24,
  },
  balanceAmount: {
    fontSize: 48,
    fontWeight: 'bold',
    color: '#fff',
    marginBottom: 8,
    letterSpacing: 1,
  },
  balanceCny: {
    fontSize: 16,
    color: 'rgba(255, 255, 255, 0.9)',
  },
  orderLink: {
    alignItems: 'flex-end',
    paddingTop: 8,
  },
  orderLinkText: {
    fontSize: 14,
    color: '#fff',
    marginBottom: 8,
  },
  orderLinkArrow: {
    fontSize: 32,
    color: '#fff',
    fontWeight: '300',
  },
  balanceDivider: {
    height: 1,
    backgroundColor: 'rgba(255, 255, 255, 0.4)',
    marginVertical: 24,
  },
  balanceDetails: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingBottom: 16,
  },
  balanceItem: {
    flex: 1,
  },
  balanceLabel: {
    fontSize: 14,
    color: 'rgba(255, 255, 255, 0.9)',
    marginBottom: 8,
  },
  balanceValue: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#fff',
  },
  actionList: {
    backgroundColor: '#fff',
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    overflow: 'hidden',
    paddingTop: 8,
  },
  actionItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 20,
    paddingVertical: 24,
    backgroundColor: '#fff',
  },
  actionItemPressed: {
    backgroundColor: '#F8F8F8',
  },
  actionIcon: {
    width: 56,
    height: 56,
    borderRadius: 16,
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 16,
  },
  actionIconText: {
    fontSize: 28,
  },
  actionContent: {
    flex: 1,
  },
  actionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#333',
  },
  actionArrow: {
    fontSize: 28,
    color: '#CCCCCC',
    fontWeight: '300',
  },
});

