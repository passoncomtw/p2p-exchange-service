import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import { Platform, ActivityIndicator, View, StyleSheet } from 'react-native';
import { useAppSelector } from './store/hooks';
import { setNavigationRef } from '../apis/httpClient';

import WalletScreen from './screens/WalletScreen';
import TradeScreen from './screens/TradeScreen';
import OrdersScreen from './screens/OrdersScreen';
import OrderListScreen from './screens/OrderListScreen';
import OrderDetailScreen from './screens/OrderDetailScreen';
import ProfileScreen from './screens/ProfileScreen';
import CreateOrderScreen from './screens/CreateOrderScreen';
import ConfirmOrderScreen from './screens/ConfirmOrderScreen';
import NotFound from './screens/NotFoundScreen';
import LoginScreen from './screens/LoginScreen';
import RegisterScreen from './screens/RegisterScreen';

import { HapticTab } from '@/components/HapticTab';
import { IconSymbol } from '@/components/ui/IconSymbol';
import TabBarBackground from '@/components/ui/TabBarBackground';

/**
 * 底部 Tab 導航定義
 */
const Tab = createBottomTabNavigator();

function HomeTabs() {
  return (
    <Tab.Navigator
      screenOptions={{
        headerShown: false,
        tabBarButton: HapticTab,
        tabBarBackground: TabBarBackground,
        tabBarStyle: Platform.select({
          ios: {
            // Use a transparent background on iOS to show the blur effect
            position: 'absolute' as const,
          },
          default: {},
        }),
        tabBarActiveTintColor: '#007AFF',
        tabBarInactiveTintColor: '#8E8E93',
      }}
    >
      <Tab.Screen
        name="Wallet"
        component={WalletScreen}
        options={{
          title: '錢包',
          tabBarIcon: ({ color }) => <IconSymbol name="wallet-outline" size={28} color={color} />,
        }}
      />
      <Tab.Screen
        name="Trade"
        component={TradeScreen}
        options={{
          title: '交易',
          tabBarIcon: ({ color }) => <IconSymbol name="swap-horizontal-outline" size={28} color={color} />,
        }}
      />
      <Tab.Screen
        name="Orders"
        component={OrdersScreen}
        options={{
          title: '掛單',
          tabBarIcon: ({ color }) => <IconSymbol name="clipboard-outline" size={28} color={color} />,
        }}
      />
      <Tab.Screen
        name="OrderList"
        component={OrderListScreen}
        options={{
          title: '訂單',
          tabBarIcon: ({ color }) => <IconSymbol name="document-text-outline" size={28} color={color} />,
        }}
      />
      <Tab.Screen
        name="Profile"
        component={ProfileScreen}
        options={{
          title: '我',
          tabBarIcon: ({ color }) => <IconSymbol name="person-outline" size={28} color={color} />,
        }}
      />
    </Tab.Navigator>
  );
}

/**
 * 導航堆疊定義
 */
const Stack = createNativeStackNavigator();

/**
 * 載入畫面組件
 * 在驗證狀態載入時顯示
 */
function LoadingScreen() {
  return (
    <View style={styles.loadingContainer}>
      <ActivityIndicator size="large" />
    </View>
  );
}

/**
 * 已驗證用戶的導航堆疊
 */
function AuthenticatedStack() {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false }}>
      <Stack.Screen name="HomeTabs" component={HomeTabs} />
      <Stack.Screen 
        name="CreateOrderBuy" 
        component={CreateOrderScreen}
      />
      <Stack.Screen 
        name="CreateOrderSell" 
        component={CreateOrderScreen}
      />
      <Stack.Screen 
        name="ConfirmOrder" 
        component={ConfirmOrderScreen}
      />
      <Stack.Screen 
        name="OrderDetail" 
        component={OrderDetailScreen}
      />
      <Stack.Screen 
        name="NotFound" 
        component={NotFound}
        options={{ title: '404', headerShown: true }}
      />
    </Stack.Navigator>
  );
}

/**
 * 未驗證用戶的公開頁面導航堆疊
 */
function PublicStack() {
  return (
    <Stack.Navigator screenOptions={{ headerShown: false }}>
      <Stack.Screen 
        name="Login" 
        component={LoginScreen}
        options={{ title: '登入' }}
      />
      <Stack.Screen 
        name="Register" 
        component={RegisterScreen}
        options={{ title: '註冊' }}
      />
    </Stack.Navigator>
  );
}

/**
 * 根導航組件
 * 使用 early return 模式根據驗證狀態渲染對應的導航堆疊
 */
export function Navigation(props: any) {
  const { isAuthenticated, loading } = useAppSelector((state) => state.auth);

  // Early return: 驗證狀態載入中
  if (loading) {
    return (
      <NavigationContainer {...props} ref={(ref) => setNavigationRef(ref)}>
        <LoadingScreen />
      </NavigationContainer>
    );
  }

  // Early return: 已驗證用戶
  if (isAuthenticated) {
    return (
      <NavigationContainer {...props} ref={(ref) => setNavigationRef(ref)}>
        <AuthenticatedStack />
      </NavigationContainer>
    );
  }

  // 預設: 未驗證用戶
  return (
    <NavigationContainer {...props} ref={(ref) => setNavigationRef(ref)}>
      <PublicStack />
    </NavigationContainer>
  );
}

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#fff',
  },
});

// 型別定義
type RootStackParamList = {
  HomeTabs: undefined;
  CreateOrderBuy: { type: 'buy' };
  CreateOrderSell: { type: 'sell' };
  ConfirmOrderBuy: {
    orderNumber: string;
    amount: number;
    totalPrice: number;
    sellerName: string;
    paymentTimeout: number;
    bankName: string;
    bankAccount: string;
    accountHolder: string;
  };
  ConfirmOrderSell: {
    orderNumber: string;
    amount: number;
    totalPrice: number;
    buyerName: string;
    paymentTimeout: number;
    bankName: string;
    bankAccount: string;
  };
  OrderDetail: {
    orderId: string;
  };
  NotFound: undefined;
};

type PublicStackParamList = {
  Login: undefined;
  Register: undefined;
};

type TabParamList = {
  Wallet: undefined;
  Trade: { initialTab?: 'buy' | 'sell' };
  Orders: undefined;
  OrderList: undefined;
  Profile: undefined;
};

declare global {
  namespace ReactNavigation {
    interface RootParamList extends RootStackParamList, PublicStackParamList, TabParamList {}
  }
}
