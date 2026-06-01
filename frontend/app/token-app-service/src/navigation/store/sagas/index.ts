import { all, fork } from 'redux-saga/effects';
import { watchAuthSagas } from './authSagas';
import { watchOrdersSagas } from './ordersSaga';
import { watchBankCardsSagas } from './bankCardsSaga';
import { watchMarketSagas } from './marketSaga';
import { watchErrorSaga } from './errorSaga';

export default function* rootSaga() {
  yield all([
    fork(watchAuthSagas),
    fork(watchOrdersSagas),
    fork(watchBankCardsSagas),
    fork(watchMarketSagas),
    fork(watchErrorSaga), // Root Error Saga - 統一處理所有錯誤
  ]);
}
