import 'react-native-reanimated';
import { DarkTheme, DefaultTheme } from '@react-navigation/native';
import { useFonts } from 'expo-font';
import * as SplashScreen from 'expo-splash-screen';
import * as React from 'react';
import { useColorScheme, ActivityIndicator, View } from 'react-native';
import { Provider } from 'react-redux';
import { PersistGate } from 'redux-persist/integration/react';

import { Colors } from './constants/Colors';
import { Navigation } from './navigation';
import { store, persistor } from './navigation/store/configureStore';
import { setStoreRef } from './apis';

SplashScreen.preventAutoHideAsync();

// 設定 Redux store 引用，讓 httpClient 可以從 store 讀取 token
setStoreRef(store);

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
          colors: { ...DarkTheme.colors, primary: Colors[colorScheme ?? 'light'].tint },
        }
      : {
          ...DefaultTheme,
          colors: { ...DefaultTheme.colors, primary: Colors[colorScheme ?? 'light'].tint },
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
      >
      <Navigation
        theme={theme}
        linking={{
          enabled: 'auto',
          prefixes: [
            // Change the scheme to match your app's scheme defined in app.json
            'helloworld://',
          ],
        }}
        onReady={() => {
            SplashScreen.hideAsync();
          }}
        />
      </PersistGate>
    </Provider>
  );
}
