/**
 * RegisterScreen - E幣錢包註冊頁面
 * 
 * 功能：
 * - 8 個輸入欄位（暱稱、帳號、郵箱、登入密碼、確認登入密碼、交易密碼、確認交易密碼、推薦碼）
 * - 表單驗證
 * - 返回登入頁面
 */

import React, { useState, useEffect, useRef } from 'react';
import { 
  View, 
  TextInput, 
  StyleSheet, 
  Pressable, 
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  Text,
  ActivityIndicator
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAppDispatch, useAppSelector } from '@/navigation/store/hooks';
import { registerRequest } from '@/navigation/store/actions/authActions';
import { clearError } from '@/navigation/store/slices/authSlices';

export default function RegisterScreen() {
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const { loading, error } = useAppSelector((state) => state.auth);
  
  // 用於追蹤是否正在註冊
  const isRegisteringRef = useRef(false);

  const [formData, setFormData] = useState({
    nickname: '',
    account: '',
    email: '',
    password: '',
    confirmPassword: '',
    transactionPassword: '',
    confirmTransactionPassword: '',
    referralCode: '',
  });

  // 監聽註冊狀態，成功後自動返回登入頁面
  useEffect(() => {
    // 如果正在註冊，且現在已經不在 loading 狀態，且沒有錯誤
    if (isRegisteringRef.current && !loading && !error) {
      isRegisteringRef.current = false;
      // 註冊成功，返回登入頁面
      setTimeout(() => {
        navigation.goBack();
      }, 500); // 延遲一下讓使用者看到成功訊息
    }
  }, [loading, error, navigation]);

  const handleChange = (field: string, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
  };

  const validateForm = () => {
    // 必填欄位檢查
    if (!formData.nickname.trim()) {
      Alert.alert('錯誤', '請輸入暱稱');
      return false;
    }
    if (!formData.account.trim()) {
      Alert.alert('錯誤', '請輸入帳號');
      return false;
    }
    if (!formData.email.trim()) {
      Alert.alert('錯誤', '請輸入郵箱');
      return false;
    }
    
    // 郵箱格式驗證
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(formData.email)) {
      Alert.alert('錯誤', '郵箱格式不正確');
      return false;
    }

    // 登入密碼檢查
    if (!formData.password) {
      Alert.alert('錯誤', '請輸入登入密碼');
      return false;
    }
    if (formData.password.length < 6) {
      Alert.alert('錯誤', '登入密碼至少需要 6 個字符');
      return false;
    }
    if (formData.password !== formData.confirmPassword) {
      Alert.alert('錯誤', '兩次輸入的登入密碼不一致');
      return false;
    }

    // 交易密碼檢查
    if (!formData.transactionPassword) {
      Alert.alert('錯誤', '請輸入交易密碼');
      return false;
    }
    if (formData.transactionPassword.length < 6) {
      Alert.alert('錯誤', '交易密碼至少需要 6 個字符');
      return false;
    }
    if (formData.transactionPassword !== formData.confirmTransactionPassword) {
      Alert.alert('錯誤', '兩次輸入的交易密碼不一致');
      return false;
    }

    return true;
  };

  const handleRegister = () => {
    if (!validateForm()) {
      return;
    }

    // 標記正在註冊
    isRegisteringRef.current = true;

    // 使用 Saga 處理註冊
    dispatch(registerRequest({
      nickname: formData.nickname.trim(),
      account: formData.account.trim(),
      email: formData.email.trim(),
      password: formData.password,
      transactionPassword: formData.transactionPassword,
      referralCode: formData.referralCode.trim(),
    }));
  };

  return (
    <KeyboardAvoidingView
      style={styles.flex1}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <View style={styles.container}>
        {/* 頂部導航欄 */}
        <View style={styles.topBar}>
          <Pressable 
            onPress={() => navigation.goBack()} 
            style={styles.backButton}
            hitSlop={{ top: 10, bottom: 10, left: 10, right: 10 }}
          >
            <Text style={styles.backButtonText}>‹</Text>
          </Pressable>
          <Text style={styles.title}>註冊</Text>
          <View style={styles.placeholder} />
        </View>

        {/* 表單內容 */}
        <ScrollView 
          style={styles.scrollView}
          contentContainerStyle={styles.scrollContent}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          <View style={styles.formContainer}>
            {/* 暱稱 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>暱稱</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入暱稱"
                  value={formData.nickname}
                  onChangeText={(value) => handleChange('nickname', value)}
                  autoCapitalize="none"
                  autoCorrect={false}
                />
              </View>
            </View>

            {/* 帳號 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>帳號</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入帳號"
                  value={formData.account}
                  onChangeText={(value) => handleChange('account', value)}
                  autoCapitalize="none"
                  autoCorrect={false}
                />
              </View>
            </View>

            {/* 郵箱 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>郵箱</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入郵箱"
                  value={formData.email}
                  onChangeText={(value) => handleChange('email', value)}
                  keyboardType="email-address"
                  autoCapitalize="none"
                  autoCorrect={false}
                />
              </View>
            </View>

            {/* 登入密碼 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>登入密碼</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入登入密碼"
                  value={formData.password}
                  onChangeText={(value) => handleChange('password', value)}
                  secureTextEntry
                  autoCapitalize="none"
                />
              </View>
            </View>

            {/* 再次輸入登入密碼 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>再次輸入登入密碼</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請再次輸入登入密碼"
                  value={formData.confirmPassword}
                  onChangeText={(value) => handleChange('confirmPassword', value)}
                  secureTextEntry
                  autoCapitalize="none"
                />
              </View>
            </View>

            {/* 交易密碼 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>交易密碼</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入交易密碼"
                  value={formData.transactionPassword}
                  onChangeText={(value) => handleChange('transactionPassword', value)}
                  secureTextEntry
                  autoCapitalize="none"
                />
              </View>
            </View>

            {/* 再次輸入交易密碼 */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>再次輸入交易密碼</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請再次輸入交易密碼"
                  value={formData.confirmTransactionPassword}
                  onChangeText={(value) => handleChange('confirmTransactionPassword', value)}
                  secureTextEntry
                  autoCapitalize="none"
                />
              </View>
            </View>

            {/* 推薦碼（選填） */}
            <View style={styles.inputGroup}>
              <Text style={styles.label}>推薦碼（選填）</Text>
              <View style={styles.inputField}>
                <TextInput
                  style={styles.input}
                  placeholder="請輸入推薦碼"
                  value={formData.referralCode}
                  onChangeText={(value) => handleChange('referralCode', value)}
                  autoCapitalize="characters"
                  autoCorrect={false}
                />
              </View>
            </View>

            {/* 錯誤訊息 */}
            {error && (
              <View style={styles.errorContainer}>
                <Text style={styles.errorText}>❌ {error}</Text>
              </View>
            )}

            {/* 完成按鈕 */}
            <Pressable
              style={({ pressed }) => [
                styles.buttonPrimary,
                loading && styles.buttonDisabled,
                pressed && !loading && styles.buttonPressed,
              ]}
              onPress={handleRegister}
              disabled={loading}
            >
              {loading ? (
                <ActivityIndicator color="#fff" />
              ) : (
                <Text style={styles.buttonText}>完成</Text>
              )}
            </Pressable>
          </View>
        </ScrollView>
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  flex1: {
    flex: 1,
  },
  container: {
    flex: 1,
    backgroundColor: '#fff',
  },
  topBar: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 16,
    paddingVertical: 12,
    borderBottomWidth: 1,
    borderBottomColor: '#E5E5E5',
    backgroundColor: '#fff',
    zIndex: 10,
  },
  backButton: {
    width: 44,
    height: 44,
    justifyContent: 'center',
    alignItems: 'center',
    zIndex: 20,
  },
  backButtonText: {
    fontSize: 32,
    color: '#007AFF',
    fontWeight: '300',
  },
  title: {
    fontSize: 18,
    fontWeight: '600',
    color: '#000',
  },
  placeholder: {
    width: 40,
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    paddingBottom: 40,
  },
  formContainer: {
    padding: 16,
  },
  inputGroup: {
    marginBottom: 16,
  },
  label: {
    fontSize: 14,
    color: '#333',
    marginBottom: 8,
    fontWeight: '500',
  },
  inputField: {
    backgroundColor: '#F5F5F5',
    borderRadius: 8,
    height: 50,
    justifyContent: 'center',
    paddingHorizontal: 15,
  },
  input: {
    fontSize: 16,
    color: '#333',
  },
  buttonPrimary: {
    backgroundColor: '#007AFF',
    paddingVertical: 15,
    borderRadius: 8,
    alignItems: 'center',
    marginTop: 24,
    shadowColor: '#007AFF',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 5,
  },
  buttonText: {
    color: '#fff',
    fontSize: 18,
    fontWeight: '600',
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonPressed: {
    transform: [{ scale: 0.98 }],
  },
  errorContainer: {
    backgroundColor: '#FFEBEE',
    padding: 12,
    borderRadius: 8,
    marginTop: 16,
    alignItems: 'center',
  },
  errorText: {
    color: '#C62828',
    fontSize: 14,
    textAlign: 'center',
  },
});

