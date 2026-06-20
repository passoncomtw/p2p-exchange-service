import * as React from 'react';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { NavigationContainer } from '@react-navigation/native';
import { useTranslation } from 'react-i18next';
import * as tokens from '@/theme';

import { IconSymbol } from '@/components/ui/IconSymbol';
import V1CreateOrderScreen from './screens/V1CreateOrderScreen';
import V1MyOrdersScreen from './screens/V1MyOrdersScreen';

const { colors } = tokens;
const Tab = createBottomTabNavigator();

// v1 使用者端底部 tab：掛單 / 我的掛單（免登入）。
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
          tabBarIcon: ({ color }) => <IconSymbol name="clipboard-outline" size={26} color={color} />,
        }}
      />
      <Tab.Screen
        name="MyOrders"
        component={V1MyOrdersScreen}
        options={{
          title: t('order.nav.myOrders'),
          tabBarIcon: ({ color }) => <IconSymbol name="document-text-outline" size={26} color={color} />,
        }}
      />
    </Tab.Navigator>
  );
}

export function V1Navigation(props: any) {
  return (
    <NavigationContainer {...props}>
      <V1Tabs />
    </NavigationContainer>
  );
}

export default V1Navigation;
