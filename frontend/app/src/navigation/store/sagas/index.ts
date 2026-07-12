import { all, fork } from 'redux-saga/effects';
import { watchAuthSagas } from './authSagas';
import { watchOrdersSagas } from './ordersSaga';
import { watchBankCardsSagas } from './bankCardsSaga';
import { watchErrorSaga } from './errorSaga';
import { watchWalletSagas } from './walletSaga';

export default function* rootSaga() {
  yield all([
    fork(watchAuthSagas),
    fork(watchOrdersSagas),
    fork(watchBankCardsSagas),
    fork(watchErrorSaga),
    fork(watchWalletSagas),
  ]);
}
