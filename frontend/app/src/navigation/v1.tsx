import * as React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { useSafeAreaInsets } from 'react-native-safe-area-context';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { NavigationContainer } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import * as tokens from '@/theme';

import { IconSymbol } from '@/components/ui/IconSymbol';
import { useAppDispatch, useAppSelector } from '@/navigation/store/hooks';
import { logoutRequest } from '@/navigation/store/actions/authActions';
import { navigationRef } from '@/navigation/navigationRef';
import V1CreateOrderScreen from './screens/V1CreateOrderScreen';
import V1MyOrdersScreen from './screens/V1MyOrdersScreen';
import V1LoginScreen from './screens/V1LoginScreen';
import V1TradeMarketScreen from './screens/V1TradeMarketScreen';
import V1OrdersScreen from './screens/V1OrdersScreen';
import V1ListingDetailScreen from './screens/V1ListingDetailScreen';
import V1OrderDetailScreen from './screens/V1OrderDetailScreen';
import V1AddPaymentMethodScreen from './screens/V1AddPaymentMethodScreen';
import WalletScreen from './screens/WalletScreen';

const { colors } = tokens;
const Tab = createBottomTabNavigator();
const RootStack = createNativeStackNavigator();
const MainStack = createNativeStackNavigator();

function AppBar() {
  const { t } = useTranslation();
  const dispatch = useAppDispatch();
  const insets = useSafeAreaInsets();
  return (
    <View style={[styles.appBarWrap, { paddingTop: insets.top }]}>
      <View style={styles.appBar}>
        <View style={styles.logo}>
          <Text style={styles.logoText}>P</Text>
        </View>
        <Text style={styles.brand}>{t('order.login.brand')}</Text>
        <View style={{ flex: 1 }} />
        <TouchableOpacity onPress={() => dispatch(logoutRequest())} accessibilityRole="button">
          <Text style={styles.logout}>{t('order.login.logout')}</Text>
        </TouchableOpacity>
      </View>
    </View>
  );
}

function V1Tabs() {
  const { t } = useTranslation();
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarActiveTintColor: colors.primaryDeep,
        tabBarInactiveTintColor: colors.textTertiary,
      }}
    >
      <Tab.Screen
        name="CreateOrder"
        component={V1CreateOrderScreen}
        options={{
          title: t('order.nav.create'),
          tabBarIcon: ({ color }) => <IconSymbol name="create-outline" size={24} color={color} />,
        }}
      />
      <Tab.Screen
        name="Trade"
        component={V1TradeMarketScreen}
        options={{
          title: t('order.nav.trade'),
          tabBarIcon: ({ color }) => <IconSymbol name="swap-horizontal-outline" size={24} color={color} />,
        }}
      />
      <Tab.Screen
        name="MyOrders"
        component={V1MyOrdersScreen}
        options={{
          title: t('order.nav.myOrders'),
          tabBarIcon: ({ color }) => <IconSymbol name="list-outline" size={24} color={color} />,
        }}
      />
      <Tab.Screen
        name="Orders"
        component={V1OrdersScreen}
        options={{
          title: t('order.nav.orders'),
          tabBarIcon: ({ color }) => <IconSymbol name="receipt-outline" size={24} color={color} />,
        }}
      />
      <Tab.Screen
        name="Wallet"
        component={WalletScreen}
        options={{
          title: t('order.nav.wallet'),
          tabBarIcon: ({ color }) => <IconSymbol name="wallet-outline" size={24} color={color} />,
        }}
      />
    </Tab.Navigator>
  );
}

function TabsShell() {
  return (
    <View style={styles.shell}>
      <AppBar />
      <V1Tabs />
    </View>
  );
}

function MainNavigator() {
  const { t } = useTranslation();
  return (
    <MainStack.Navigator>
      <MainStack.Screen
        name="Tabs"
        component={TabsShell}
        options={{ headerShown: false }}
      />
      <MainStack.Screen
        name="ListingDetail"
        component={V1ListingDetailScreen}
        options={{ title: t('order.pageTitle.listingDetail') }}
      />
      <MainStack.Screen
        name="OrderDetail"
        component={V1OrderDetailScreen}
        options={{ title: t('order.pageTitle.orderDetail') }}
      />
      <MainStack.Screen
        name="AddPaymentMethod"
        component={V1AddPaymentMethodScreen}
        options={{ title: t('order.pageTitle.addPayment') }}
      />
    </MainStack.Navigator>
  );
}

function RootNavigator() {
  const isAuthenticated = useAppSelector((s) => s.auth.isAuthenticated);
  return (
    <RootStack.Navigator screenOptions={{ headerShown: false }}>
      {isAuthenticated ? (
        <RootStack.Screen name="Main" component={MainNavigator} />
      ) : (
        <RootStack.Screen
          name="Login"
          component={V1LoginScreen}
          options={{ animationTypeForReplace: 'pop' }}
        />
      )}
    </RootStack.Navigator>
  );
}

export function V1Navigation(props: any) {
  return (
    <NavigationContainer ref={navigationRef} {...props}>
      <RootNavigator />
    </NavigationContainer>
  );
}

export default V1Navigation;

const styles = StyleSheet.create({
  shell: { flex: 1, backgroundColor: colors.bgContent },
  appBarWrap: { backgroundColor: colors.bgCard, borderBottomWidth: 1, borderBottomColor: colors.borderCard },
  appBar: { height: 52, flexDirection: 'row', alignItems: 'center', paddingHorizontal: 16, gap: 10 },
  logo: {
    width: 26,
    height: 26,
    borderRadius: 7,
    backgroundColor: colors.primary,
    alignItems: 'center',
    justifyContent: 'center',
  },
  logoText: { fontSize: 13, fontWeight: '700', color: '#1F2327' },
  brand: { fontSize: 16, fontWeight: '600', color: colors.textPrimary },
  logout: { fontSize: 12, color: colors.textTertiary, paddingVertical: 6, paddingHorizontal: 4 },
});
