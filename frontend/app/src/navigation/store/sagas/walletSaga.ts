import { call, put, takeLatest } from 'redux-saga/effects';
import { SagaIterator } from 'redux-saga';
import { PayloadAction } from '@reduxjs/toolkit';
import logger from '@pkg/logger';
import { handleSagaError } from '@pkg/utils/sagaHelpers';
import { walletApi } from '@/apis/walletApi';
import { WALLET_ACTIONS } from '../actions/walletActions';
import type { ListWalletLedgersParams } from '@/interfaces/wallet';
import {
  fetchWalletsStart,
  fetchWalletsSuccess,
  fetchWalletsFailure,
  fetchLedgersStart,
  fetchLedgersSuccess,
  fetchLedgersFailure,
} from '../slices/walletSlice';

function* fetchWalletsSaga(): SagaIterator {
  yield put(fetchWalletsStart());
  try {
    const data = yield call(walletApi.listWallets);
    logger.info('取得錢包列表成功', { count: data.list?.length });
    yield put(fetchWalletsSuccess(data.list ?? []));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得錢包列表失敗');
    yield put(fetchWalletsFailure(errorResult));
  }
}

function* fetchLedgersSaga(action: PayloadAction<ListWalletLedgersParams>): SagaIterator {
  yield put(fetchLedgersStart());
  try {
    const data = yield call(walletApi.listLedgers, action.payload);
    logger.info('取得帳本記錄成功', { currency: action.payload.currency, total: data.total });
    yield put(fetchLedgersSuccess({ list: data.list ?? [], total: data.total ?? 0 }));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得帳本記錄失敗');
    yield put(fetchLedgersFailure(errorResult));
  }
}

export function* watchWalletSagas(): SagaIterator {
  yield takeLatest(WALLET_ACTIONS.FETCH_WALLETS_REQUEST, fetchWalletsSaga);
  yield takeLatest(WALLET_ACTIONS.FETCH_LEDGERS_REQUEST, fetchLedgersSaga);
}
