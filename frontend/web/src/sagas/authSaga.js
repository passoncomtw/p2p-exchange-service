import { call, put } from 'redux-saga/effects'
import { loginApi } from 'src/apis/auth'
import { signInSuccess, signInError } from 'src/slices/authSlice'

export function* signInSaga({ payload }) {
  try {
    const response = yield call(loginApi, payload)
    const { token, expiresIn } = response.data.data
    yield put(signInSuccess({ token, username: payload.username, expiresIn }))
  } catch (error) {
    const message = error.response?.data?.message ?? 'Login failed'
    yield put(signInError(message))
  }
}
