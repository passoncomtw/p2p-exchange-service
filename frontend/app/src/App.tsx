import 'react-native-reanimated';
import './i18n';
import { DarkTheme, DefaultTheme } from '@react-navigation/native';
import { useFonts } from 'expo-font';
import * as SplashScreen from 'expo-splash-screen';
import * as React from 'react';
import { useColorScheme, ActivityIndicator, View } from 'react-native';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';

import { Colors } from './constants/Colors';
import { V1Navigation } from './navigation/v1';
import { store, persistor } from './navigation/store/configureStore';
import { resetTransient } from './navigation/store/slices/authSlices';
import { setStoreRef } from './apis';
import { useAppSelector } from './navigation/store/hooks';
import { usePushNotifications } from './hooks/usePushNotifications';

SplashScreen.preventAutoHideAsync();

// 設定 Redux store 引用，讓 httpClient 可以從 store 讀取 token
setStoreRef(store);

function AppInner({ theme }: { theme: typeof DefaultTheme }) {
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);
  usePushNotifications(isAuthenticated);

  return (
    <V1Navigation
      theme={theme}
      onReady={() => {
        SplashScreen.hideAsync();
      }}
    />
  );
}

export function App() {
  const colorScheme = useColorScheme();
  const [loaded] = useFonts({
    SpaceMono: require('./assets/fonts/SpaceMono-Regular.ttf'),
  });

  if (!loaded) {
    // Async font loading only occurs in development.
    return null;
  }

  const theme =
    colorScheme === 'dark'
      ? {
          ...DarkTheme,
          colors: { ...DarkTheme.colors, primary: Colors.dark.tint },
        }
      : {
          ...DefaultTheme,
          colors: { ...DefaultTheme.colors, primary: Colors.light.tint },
        };

  return (
    <Provider store={store}>
      <PersistGate
        loading={
          <View style={{ flex: 1, justifyContent: 'center', alignItems: 'center', backgroundColor: '#fff' }}>
            <ActivityIndicator size="large" color="#007AFF" />
          </View>
        }
        persistor={persistor}
        onBeforeLift={() => { store.dispatch(resetTransient()); }}
      >
        <AppInner theme={theme} />
      </PersistGate>
    </Provider>
  );
}
