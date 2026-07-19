import * as React from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
} from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { useTranslation } from 'react-i18next';
import * as tokens from '@/theme';
import { useAppDispatch, useAppSelector } from '@/navigation/store/hooks';
import { loginRequest } from '@/navigation/store/actions/authActions';

const { colors } = tokens;

// 依 design 的手機登入畫面;串接既有 auth saga(POST /app/auth/login)。
export default function V1LoginScreen() {
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const { loading } = useAppSelector((s) => s.auth);

  const [account, setAccount] = React.useState('');
  const [password, setPassword] = React.useState('');
  const [accountErr, setAccountErr] = React.useState('');
  const [passwordErr, setPasswordErr] = React.useState('');

  const handleLogin = () => {
    let ok = true;
    if (!account.trim()) {
      setAccountErr(t('order.login.accountRequired'));
      ok = false;
    }
    if (!password.trim()) {
      setPasswordErr(t('order.login.passwordRequired'));
      ok = false;
    }
    if (!ok) return;
    dispatch(loginRequest({ account: account.trim(), password }));
  };

  return (
    <SafeAreaView style={styles.safe}>
      <KeyboardAvoidingView
        style={styles.flex}
        behavior={Platform.OS === 'ios' ? 'padding' : undefined}
      >
        <ScrollView contentContainerStyle={styles.container} keyboardShouldPersistTaps="handled">
          <View style={styles.logo}>
            <Text style={styles.logoText}>P</Text>
          </View>
          <Text style={styles.brand}>{t('order.login.brand')}</Text>
          <Text style={styles.subtitle}>{t('order.login.subtitle')}</Text>

          <Text style={styles.label}>
            {t('order.login.account')} <Text style={styles.req}>*</Text>
          </Text>
          <TextInput
            style={[styles.input, accountErr && styles.inputError]}
            value={account}
            onChangeText={(v) => {
              setAccount(v);
              setAccountErr('');
            }}
            placeholder={t('order.login.accountPlaceholder')}
            placeholderTextColor={colors.textPlaceholder}
            autoCapitalize="none"
            autoCorrect={false}
          />
          {accountErr ? <Text style={styles.fieldErr}>{accountErr}</Text> : null}

          <View style={{ height: 16 }} />
          <Text style={styles.label}>
            {t('order.login.password')} <Text style={styles.req}>*</Text>
          </Text>
          <TextInput
            style={[styles.input, passwordErr && styles.inputError]}
            value={password}
            onChangeText={(v) => {
              setPassword(v);
              setPasswordErr('');
            }}
            placeholder={t('order.login.passwordPlaceholder')}
            placeholderTextColor={colors.textPlaceholder}
            secureTextEntry
          />
          {passwordErr ? <Text style={styles.fieldErr}>{passwordErr}</Text> : null}

          <TouchableOpacity
            style={[styles.submit, loading && styles.submitDisabled]}
            onPress={handleLogin}
            disabled={loading}
            accessibilityRole="button"
          >
            {loading ? <ActivityIndicator size="small" color="#1F2327" style={{ marginRight: 8 }} /> : null}
            <Text style={styles.submitText}>
              {loading ? t('order.login.submitting') : t('order.login.submit')}
            </Text>
          </TouchableOpacity>

          <Text style={styles.demo}>{t('order.login.demoHint')}</Text>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const styles = StyleSheet.create({
  safe: { flex: 1, backgroundColor: colors.bgCard },
  flex: { flex: 1 },
  container: { flexGrow: 1, justifyContent: 'center', paddingHorizontal: 30, paddingVertical: 40 },
  logo: {
    width: 56,
    height: 56,
    borderRadius: 15,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 18,
  },
  logoText: { fontSize: 26, fontWeight: '700', color: '#1F2327' },
  brand: { fontSize: 22, fontWeight: '700', color: colors.textPrimary, marginBottom: 6 },
  subtitle: { fontSize: 13, color: colors.textTertiary, marginBottom: 28 },
  label: { fontSize: 12, color: colors.textSecondary, marginBottom: 6 },
  req: { color: colors.danger },
  input: {
    height: 44,
    borderWidth: 1,
    borderColor: colors.borderInput,
    borderRadius: 4,
    paddingHorizontal: 12,
    fontSize: 14,
    color: colors.textPrimary,
    backgroundColor: colors.bgCard,
  },
  inputError: { borderColor: colors.danger },
  fieldErr: { fontSize: 11, color: colors.danger, marginTop: 5 },
  submit: {
    marginTop: 22,
    height: 48,
    borderRadius: 4,
    backgroundColor: colors.primary,
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
  },
  submitDisabled: { backgroundColor: colors.primaryDisabled },
  submitText: { fontSize: 16, fontWeight: '600', color: '#1F2327' },
  demo: { marginTop: 18, fontSize: 11, color: '#BDBDBD', lineHeight: 18 },
});
