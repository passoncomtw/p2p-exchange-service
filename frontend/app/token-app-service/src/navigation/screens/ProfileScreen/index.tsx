/**
 * ProfileScreen - 我的頁面
 */

import React from 'react';
import { View, Text, StyleSheet, ScrollView, Pressable, StatusBar, Alert, Clipboard } from 'react-native';
import { useAppSelector, useAppDispatch } from '@/navigation/store/hooks';
import { logoutRequest } from '@/navigation/store/actions/authActions';

export default function ProfileScreen() {
  const dispatch = useAppDispatch();
  const { user } = useAppSelector((state) => state.auth);

  // 獲取用戶名首字母作為頭像
  const getAvatarLetter = () => {
    if (!user?.account) return '?';
    return user.account.charAt(0).toUpperCase();
  };

  const handleProfile = () => {
    // TODO: 導航到個人資料頁面
    Alert.alert('個人資料', '功能開發中...');
  };

  const handleChangePassword = () => {
    // TODO: 導航到修改密碼頁面
    Alert.alert('修改密碼', '功能開發中...');
  };

  const handleContactService = () => {
    Alert.alert('聯繫客服', '功能開發中...');
  };

  const handleCopyReferralCode = () => {
    if (user?.referralCode) {
      Clipboard.setString(user.referralCode);
      Alert.alert('已複製', `推薦碼: ${user.referralCode}`);
    }
  };

  const handleLogout = () => {
    Alert.alert(
      '確認退出',
      '您確定退出登入嗎？',
      [
        { text: '取消', style: 'cancel' },
        {
          text: '退出',
          style: 'destructive',
          onPress: () => dispatch(logoutRequest()),
        },
      ]
    );
  };

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />
      
      {/* 頂部導航 */}
      <View style={styles.topBar}>
        <Text style={styles.topBarTitle}>我</Text>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* 用戶頭像和名稱 */}
        <View style={styles.header}>
          <View style={styles.avatar}>
            <Text style={styles.avatarText}>{getAvatarLetter()}</Text>
          </View>
          <Text style={styles.username}>{user?.account || '未登入'}</Text>
        </View>

        {/* 菜單列表 */}
        <View style={styles.menuList}>
          <Pressable 
            style={({ pressed }) => [
              styles.menuItem,
              pressed && styles.menuItemPressed,
            ]}
            onPress={handleProfile}
          >
            <Text style={styles.menuIcon}>👤</Text>
            <Text style={styles.menuTitle}>個人資料</Text>
            <Text style={styles.menuArrow}>›</Text>
          </Pressable>

          <Pressable 
            style={({ pressed }) => [
              styles.menuItem,
              pressed && styles.menuItemPressed,
            ]}
            onPress={handleChangePassword}
          >
            <Text style={styles.menuIcon}>🔒</Text>
            <Text style={styles.menuTitle}>修改密碼</Text>
            <Text style={styles.menuArrow}>›</Text>
          </Pressable>

          <Pressable 
            style={({ pressed }) => [
              styles.menuItem,
              pressed && styles.menuItemPressed,
            ]}
            onPress={handleContactService}
          >
            <Text style={styles.menuIcon}>💬</Text>
            <Text style={styles.menuTitle}>聯繫客服</Text>
            <Text style={styles.menuArrow}>›</Text>
          </Pressable>

          <Pressable 
            style={({ pressed }) => [
              styles.menuItem,
              pressed && styles.menuItemPressed,
            ]}
            onPress={handleCopyReferralCode}
          >
            <Text style={styles.menuIcon}>🎁</Text>
            <Text style={styles.menuTitle}>推薦碼</Text>
            <Text style={styles.menuText}>{user?.referralCode || '-'}</Text>
          </Pressable>

          <View style={styles.menuItem}>
            <Text style={styles.menuIcon}>ℹ️</Text>
            <Text style={styles.menuTitle}>版本</Text>
            <Text style={styles.menuText}>V 1.0.0</Text>
          </View>
        </View>

        {/* 退出登入按鈕 */}
        <View style={styles.logoutContainer}>
          <Pressable 
            style={({ pressed }) => [
              styles.logoutButton,
              pressed && styles.logoutButtonPressed,
            ]}
            onPress={handleLogout}
          >
            <Text style={styles.logoutButtonText}>退出登入</Text>
          </Pressable>
        </View>
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
  },
  topBarTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    color: '#000',
  },
  content: {
    flex: 1,
  },
  header: {
    backgroundColor: '#fff',
    paddingVertical: 32,
    alignItems: 'center',
    marginBottom: 12,
  },
  avatar: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: '#007AFF',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 16,
  },
  avatarText: {
    fontSize: 36,
    fontWeight: 'bold',
    color: '#fff',
  },
  username: {
    fontSize: 20,
    fontWeight: '600',
    color: '#333',
  },
  menuList: {
    backgroundColor: '#fff',
    marginBottom: 12,
  },
  menuItem: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 16,
    paddingHorizontal: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#F0F0F0',
  },
  menuItemPressed: {
    backgroundColor: '#F8F8F8',
  },
  menuIcon: {
    fontSize: 24,
    marginRight: 16,
  },
  menuTitle: {
    fontSize: 16,
    color: '#333',
    flex: 1,
  },
  menuText: {
    fontSize: 14,
    color: '#999',
    marginRight: 8,
  },
  menuArrow: {
    fontSize: 24,
    color: '#CCCCCC',
    fontWeight: '300',
  },
  logoutContainer: {
    padding: 16,
  },
  logoutButton: {
    backgroundColor: '#fff',
    paddingVertical: 15,
    borderRadius: 8,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#F44336',
  },
  logoutButtonPressed: {
    backgroundColor: '#FFEBEE',
  },
  logoutButtonText: {
    color: '#F44336',
    fontSize: 16,
    fontWeight: '600',
  },
});

