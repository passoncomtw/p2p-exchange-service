import { configureStore, ThunkAction, Action, combineReducers } from '@reduxjs/toolkit';
import createSagaMiddleware from 'redux-saga';
import devToolsEnhancer from 'redux-devtools-expo-dev-plugin';
import { persistStore, persistReducer } from 'redux-persist';
import AsyncStorage from '@react-native-async-storage/async-storage';
import authReducer from './slices/authSlices';
import ordersReducer from './slices/ordersSlice';
import bankCardsReducer from './slices/bankCardsSlice';
import marketReducer from './slices/marketSlice';
import rootSaga from './sagas';

/**
 * Root Reducer
 */
const rootReducer = combineReducers({
  auth: authReducer,
  orders: ordersReducer,
  bankCards: bankCardsReducer,
  market: marketReducer,
});

/**
 * Redux Persist 配置
 */
const persistConfig = {
  key: 'root',
  storage: AsyncStorage,
  whitelist: ['auth'], // 只持久化 auth state
  version: 1,
};

/**
 * Persisted Reducer
 */
const persistedReducer = persistReducer(persistConfig, rootReducer);

/**
 * 創建 Saga 中間件
 */
const sagaMiddleware = createSagaMiddleware();

/**
 * 配置 Redux Store
 * 
 * 使用 Redux Toolkit 的 configureStore，整合了：
 * - redux-thunk 中間件（用於簡單的非同步邏輯）
 * - redux-saga 中間件（用於複雜的非同步流程控制）
 * - redux-persist（狀態持久化）
 * - Expo Redux DevTools Plugin（開發環境下的除錯工具）
 * 
 * 參考文檔：https://docs.expo.dev/debugging/devtools-plugins/#redux
 */
export const store = configureStore({
  reducer: persistedReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      thunk: true,
      serializableCheck: {
        // 忽略 redux-persist 的 action
        ignoredActions: [
          'persist/PERSIST',
          'persist/REHYDRATE',
          'persist/PAUSE',
          'persist/PURGE',
          'persist/REGISTER',
        ],
        ignoredActionPaths: ['register', 'rehydrate'],
        ignoredPaths: ['_persist'],
      },
    }).concat(sagaMiddleware),
  devTools: false,
  enhancers: (getDefaultEnhancers) => 
    getDefaultEnhancers().concat(devToolsEnhancer()),
});

/**
 * Persistor
 */
export const persistor = persistStore(store);

/**
 * Run Saga
 */
sagaMiddleware.run(rootSaga);

/**
 * 從 store 本身推斷出 RootState 和 AppDispatch 型別
 * 這些型別將在整個應用程式中使用
 */
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

/**
 * Thunk Action 的型別定義
 * 用於定義非同步 action creators
 */
export type AppThunk<ReturnType = void> = ThunkAction<
  ReturnType,
  RootState,
  unknown,
  Action<string>
>;