/**
 * LoginScreen - E幣錢包登入頁面
 * 
 * 設計特點：
 * - 簡潔的白色背景
 * - 圓形 Logo (E幣)
 * - 兩個輸入框：帳號、密碼
 * - 輔助連結：忘記密碼、24h客服
 * - 主要按鈕：登入（藍色）
 * - 次要按鈕：免費註冊（白色邊框）
 * - 底部版本號
 */

import React, { useState } from 'react';
import { 
  View, 
  TextInput, 
  StyleSheet, 
  Pressable, 
  ActivityIndicator, 
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Text as RNText,
  StatusBar
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '@/navigation/store/hooks';
import { loginRequest } from '@/navigation/store/actions/authActions';

export default function LoginScreen() {
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const { loading, error } = useAppSelector((state) => state.auth);
  
  // 預填測試帳號
  const [account, setAccount] = useState('');
  const [password, setPassword] = useState('');

  const handleLogin = () => {
    // 驗證輸入
    if (!account.trim() || !password.trim()) {
      Alert.alert('錯誤', '請輸入帳號和密碼');
      return;
    }

    // 使用 Saga 處理登入
    dispatch(loginRequest({ account: account.trim(), password }));
  };

  const handleForgotPassword = () => {
    Alert.alert('忘記登入密碼', '此功能開發中...');
  };

  const handleCustomerService = () => {
    Alert.alert('24h 客服', '此功能開發中...');
  };

  const handleRegister = () => {
    navigation.navigate('Register' as never);
  };

  return (
    <View style={styles.container}>
      <StatusBar barStyle="dark-content" />
      
      <KeyboardAvoidingView 
        behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
        style={styles.keyboardView}
      >
        <ScrollView 
          contentContainerStyle={styles.scrollContent}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          <View style={styles.loginContainer}>
            <View style={styles.inputGroup}>
              {/* Logo */}
              <View style={styles.logoContainer}>
                <View style={styles.logo}>
                  <RNText style={styles.logoText}>E幣</RNText>
                </View>
              </View>

              {/* 錯誤訊息 */}
              {error && (
                <View style={styles.errorContainer}>
                  <RNText style={styles.errorText}>❌ {error}</RNText>
                </View>
              )}

              {/* 帳號輸入框 */}
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入帳號"
                  placeholderTextColor="#999"
                  value={account}
                  onChangeText={setAccount}
                  autoCapitalize="none"
                  autoCorrect={false}
                  editable={!loading}
                />
              </View>

              {/* 密碼輸入框 */}
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入登入密碼"
                  placeholderTextColor="#999"
                  value={password}
                  onChangeText={setPassword}
                  secureTextEntry
                  editable={!loading}
                  onSubmitEditing={handleLogin}
                  returnKeyType="go"
                />
              </View>
            </View>

            {/* 按鈕區域 */}
            <View style={styles.buttonContainer}>
              {/* 登入按鈕 */}
              <Pressable 
                style={({ pressed }) => [
                  styles.btnPrimary,
                  loading && styles.btnDisabled,
                  pressed && !loading && styles.btnPressed
                ]}
                onPress={handleLogin}
                disabled={loading}
              >
                {loading ? (
                  <ActivityIndicator color="#fff" size="small" />
                ) : (
                  <RNText style={styles.btnPrimaryText}>登入</RNText>
                )}
              </Pressable>

              {/* 免費註冊按鈕 */}
              <Pressable 
                style={({ pressed }) => [
                  styles.btnSecondary,
                  pressed && styles.btnPressed
                ]}
                onPress={handleRegister}
                disabled={loading}
              >
                <RNText style={styles.btnSecondaryText}>免費註冊</RNText>
              </Pressable>
            </View>

            {/* 版本號 */}
            <View style={styles.versionContainer}>
              <RNText style={styles.versionText}>version 1.0.0</RNText>
            </View>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#fff',
  },
  keyboardView: {
    flex: 1,
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: 'center',
  },
  loginContainer: {
    flex: 1,
    justifyContent: 'space-between',
    paddingVertical: 40,
  },
  
  // 輸入區域
  inputGroup: {
    flex: 1,
    justifyContent: 'center',
    paddingHorizontal: 24,
  },
  
  // Logo
  logoContainer: {
    alignItems: 'center',
    marginBottom: 60,
  },
  logo: {
    width: 100,
    height: 100,
    borderRadius: 50,
    backgroundColor: '#007AFF',
    justifyContent: 'center',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.15,
    shadowRadius: 8,
    elevation: 6,
  },
  logoText: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#fff',
  },
  
  // 錯誤訊息
  errorContainer: {
    backgroundColor: '#FFEBEE',
    padding: 12,
    borderRadius: 8,
    marginBottom: 16,
  },
  errorText: {
    color: '#C62828',
    fontSize: 14,
    textAlign: 'center',
  },
  
  // 輸入框
  inputField: {
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    marginBottom: 16,
    height: 50,
    justifyContent: 'center',
    paddingHorizontal: 16,
  },
  input: {
    fontSize: 16,
    color: '#333',
    height: '100%',
  },
  
  // 輔助連結
  helperLinks: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    marginTop: 8,
    marginBottom: 20,
  },
  linkText: {
    fontSize: 14,
    color: '#007AFF',
  },
  
  // 按鈕區域
  buttonContainer: {
    paddingHorizontal: 24,
    paddingBottom: 20,
  },
  
  // 主要按鈕 (登入)
  btnPrimary: {
    backgroundColor: '#007AFF',
    borderRadius: 8,
    height: 50,
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 12,
    shadowColor: '#007AFF',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.3,
    shadowRadius: 4,
    elevation: 3,
  },
  btnPrimaryText: {
    color: '#fff',
    fontSize: 17,
    fontWeight: '600',
  },
  
  // 次要按鈕 (免費註冊)
  btnSecondary: {
    backgroundColor: '#fff',
    borderRadius: 8,
    height: 50,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#007AFF',
  },
  btnSecondaryText: {
    color: '#007AFF',
    fontSize: 17,
    fontWeight: '600',
  },
  
  // 按鈕狀態
  btnDisabled: {
    opacity: 0.5,
  },
  btnPressed: {
    opacity: 0.7,
    transform: [{ scale: 0.98 }],
  },
  
  // 版本號
  versionContainer: {
    alignItems: 'center',
    paddingBottom: 20,
  },
  versionText: {
    fontSize: 12,
    color: '#999',
  },
});
