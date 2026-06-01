import { call, put, takeLatest } from 'redux-saga/effects';
import { SagaIterator } from 'redux-saga';
import { PayloadAction } from '@reduxjs/toolkit';
import logger from '@pkg/logger';
import { handleSagaError } from '@pkg/utils/sagaHelpers';
import { ordersApi } from '@/apis';
import type { GetPendingOrdersParams } from '@/apis/ordersApi';
import { MARKET_ACTIONS } from '../actions/marketActions';
import {
  fetchBuyOrdersStart,
  fetchBuyOrdersSuccess,
  fetchBuyOrdersFailure,
  fetchSellOrdersStart,
  fetchSellOrdersSuccess,
  fetchSellOrdersFailure,
} from '../slices/marketSlice';

/**
 * Fetch Buy Orders Saga
 * 取得買幣掛單列表（type=0，我要賣時顯示）
 */
function* fetchBuyOrdersSaga(action: PayloadAction<GetPendingOrdersParams | undefined>): SagaIterator {
  yield put(fetchBuyOrdersStart());

  try {
    const params = {
      type: 0, // 買幣掛單
      size: 20,
      page: 1,
      ...action.payload,
    };

    logger.info('取得買幣掛單列表', { params });

    const data = yield call(ordersApi.getPendingOrdersList, params);

    logger.info('取得買幣掛單列表成功', {
      count: data.rows.length,
      total: data.total,
      page: data.page,
    });

    yield put(fetchBuyOrdersSuccess(data));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得買幣掛單列表失敗');
    yield put(fetchBuyOrdersFailure(errorResult));
  }
}

/**
 * Fetch Sell Orders Saga
 * 取得賣幣掛單列表（type=1，我要買時顯示）
 */
function* fetchSellOrdersSaga(action: PayloadAction<GetPendingOrdersParams | undefined>): SagaIterator {
  yield put(fetchSellOrdersStart());

  try {
    const params = {
      type: 1, // 賣幣掛單
      size: 20,
      page: 1,
      ...action.payload,
    };

    logger.info('取得賣幣掛單列表', { params });

    const data = yield call(ordersApi.getPendingOrdersList, params);

    logger.info('取得賣幣掛單列表成功', {
      count: data.rows.length,
      total: data.total,
      page: data.page,
    });

    yield put(fetchSellOrdersSuccess(data));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得賣幣掛單列表失敗');
    yield put(fetchSellOrdersFailure(errorResult));
  }
}

/**
 * Watcher Saga
 * 監聽市場相關的 actions
 */
export function* watchMarketSagas(): SagaIterator {
  yield takeLatest(MARKET_ACTIONS.FETCH_BUY_ORDERS_REQUEST, fetchBuyOrdersSaga);
  yield takeLatest(MARKET_ACTIONS.FETCH_SELL_ORDERS_REQUEST, fetchSellOrdersSaga);
}

