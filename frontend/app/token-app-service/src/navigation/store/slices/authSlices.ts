import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { SagaErrorResult } from '@pkg/utils/sagaHelpers';
import type { AuthState, User } from '@/interfaces/store';

const initialState: AuthState = {
  isAuthenticated: false,
  user: null,
  accessToken: null,
  expireIn: null,
  loading: false,
  error: null,
};

/**
 * Auth Slice
 * 負責管理使用者認證相關的狀態
 */
const authSlice = createSlice({
  name: 'auth',
  initialState,
  reducers: {
    loginStart(state) {
      state.loading = true;
      state.error = null;
    },
    loginSuccess(state, action: PayloadAction<{ user: User; accessToken: string; expireIn: number }>) {
      state.loading = false;
      state.isAuthenticated = true;
      state.user = action.payload.user;
      state.accessToken = action.payload.accessToken;
      state.expireIn = action.payload.expireIn;
      state.error = null;
    },
    loginFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.loading = false;
      state.isAuthenticated = false;
      state.error = action.payload.message;
    },
    registerSuccess(state) {
      state.loading = false;
      state.error = null;
    },
    clearError(state) {
      state.error = null;
    },
    updateUser(state, action: PayloadAction<User>) {
      state.user = action.payload;
    },
    logout(state) {
      // 完全重置為初始狀態
      return initialState;
    },
  },
});

export const { loginStart, loginSuccess, loginFailure, logout, registerSuccess, clearError, updateUser } = authSlice.actions;

export default authSlice.reducer;