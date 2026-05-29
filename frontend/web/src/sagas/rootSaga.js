import { all, takeLatest } from 'redux-saga/effects'
import { signIn } from 'src/slices/authSlice'
import { signInSaga } from 'src/sagas/authSaga'

export default function* rootSaga() {
  yield all([
    takeLatest(signIn.type, signInSaga),
  ])
}
