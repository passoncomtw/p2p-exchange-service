import { call, put, takeLatest, delay } from 'redux-saga/effects';
import { PayloadAction } from '@reduxjs/toolkit';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { loginStart, loginSuccess, loginFailure, logout, registerSuccess } from '../slices/authSlices';
import { fetchBankCardsSuccess, resetBankCards, BankCard } from '../slices/bankCardsSlice';
import { fetchOrdersSuccess, resetOrders } from '../slices/ordersSlice';
import logger from '@pkg/logger';
import { handleSagaError } from '@pkg/utils/sagaHelpers';
import { authApi } from '@/apis';
import type { LoginCredentials, RegisterCredentials, LoginData, RegisterData } from '@/apis';

/**
 * Action Types for Saga
 * 定義 Saga 專用的 action types
 */
export const AUTH_SAGA_ACTIONS = {
  LOGIN_REQUEST: 'auth/loginRequest',
  REGISTER_REQUEST: 'auth/registerRequest',
  LOGOUT_REQUEST: 'auth/logoutRequest',
} as const;

/**
 * Login Saga
 * 使用 httpClient 處理登入流程
 */
function* loginSaga(action: PayloadAction<LoginCredentials>) {
  // 開始登入（設置 loading 狀態）
  yield put(loginStart());

  try {
    // 使用 httpClient 呼叫登入 API
    const data: LoginData = yield call(authApi.login, action.payload);

      logger.info('登入成功', {
        userId: data.user.id,
        userName: data.user.name,
        account: data.user.account,
      bankCardsCount: data.user.bankCards?.length || 0,
      });

    // 儲存 token 到 AsyncStorage
      if (data.access_token) {
      yield call([AsyncStorage, 'setItem'], 'authToken', data.access_token);
      yield call([AsyncStorage, 'setItem'], 'expireIn', data.expireIn.toString());
      yield call([AsyncStorage, 'setItem'], 'user', JSON.stringify(data.user));
      }

      // Dispatch loginSuccess（更新 Redux State）
      yield put(loginSuccess({
        user: data.user,
        accessToken: data.access_token,
        expireIn: data.expireIn,
      }));

    // 清除舊的掛單資料
    yield put(fetchOrdersSuccess({ buy: null, sell: null }));

    // 同步銀行卡資料到 bankCards store
    if (data.user.bankCards && data.user.bankCards.length > 0) {
      logger.info('🔄 登入成功 - 同步銀行卡資料到 store', {
        userId: data.user.id,
        userName: data.user.name,
        count: data.user.bankCards.length,
        cardIds: data.user.bankCards.map(c => c.id),
        cards: data.user.bankCards.map(c => ({
          id: c.id,
          name: c.name,
          cardNumber: c.cardNumber,
        })),
      });

      yield put(fetchBankCardsSuccess(data.user.bankCards as BankCard[]));
    } else {
      // 如果沒有銀行卡，清空 store
      logger.info('🔄 登入成功 - 清空銀行卡資料（無銀行卡）', {
        userId: data.user.id,
        userName: data.user.name,
      });
      yield put(fetchBankCardsSuccess([]));
    }
  } catch (error: any) {
    const errorResult = handleSagaError(error, '登入失敗');
    yield put(loginFailure(errorResult));
  }
}

/**
 * Register Saga
 * 使用 httpClient 處理註冊流程
 * 註冊成功後返回登入頁面
 */
function* registerSaga(action: PayloadAction<RegisterCredentials>) {
  // 開始註冊（設置 loading 狀態）
  yield put(loginStart());

  try {
    // 使用 httpClient 呼叫註冊 API
    const data: RegisterData = yield call(authApi.register, action.payload);

      logger.info('註冊成功', {
        userId: data.id,
        userName: data.name,
        account: data.account,
      });

      // 清除 loading 狀態並標記註冊成功
      yield put(registerSuccess());
  } catch (error: any) {
    const errorResult = handleSagaError(error, '註冊失敗');
    yield put(loginFailure(errorResult));
  }
}

/**
 * Logout Saga
 * 使用 httpClient 處理登出流程
 */
function* logoutSaga() {
  logger.info('開始登出流程');

  try {
    // 使用 httpClient 呼叫登出 API
    yield call(authApi.logout);
      logger.info('登出 API 呼叫成功');
  } catch (error: any) {
    logger.warn('登出 API 呼叫失敗', error);
    // 即使 API 失敗，也要清除本地數據並登出
  } finally {
    // 清除 AsyncStorage（包括 Redux Persist 緩存）
    yield call([AsyncStorage, 'multiRemove'], ['authToken', 'expireIn', 'user', 'persist:root']);

    logger.info('🧹 清除所有本地資料和 Store 狀態');

    // 重置所有 Store 的資料為初始狀態
    yield put(logout()); // 重置 auth state
    yield put(resetBankCards()); // 重置 bankCards state
    yield put(resetOrders()); // 重置 orders state

    logger.info('✅ 登出完成 - 所有資料已重置為初始狀態', {
      auth: '已重置',
      bankCards: '已重置',
      orders: '已重置',
      asyncStorage: '已清除',
      persistCache: '已清除',
    });

      // 延遲一下讓使用者看到登出動畫
      yield delay(300);
  }
}

/**
 * Watcher Saga
 * 監聽特定的 action 並觸發對應的 saga
 */
export function* watchAuthSagas() {
  // takeLatest: 如果有多個請求，只處理最新的一個
  yield takeLatest(AUTH_SAGA_ACTIONS.LOGIN_REQUEST, loginSaga);
  yield takeLatest(AUTH_SAGA_ACTIONS.REGISTER_REQUEST, registerSaga);
  yield takeLatest(AUTH_SAGA_ACTIONS.LOGOUT_REQUEST, logoutSaga);
}
