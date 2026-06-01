import { AUTH_SAGA_ACTIONS } from "../sagas/authSagas";

export interface LoginRequest {
  account: string;
  password: string;
}

export interface RegisterRequest {
  nickname: string;
  account: string;
  email: string;
  password: string;
  transactionPassword: string;
  referralCode?: string;
  notificationToken?: string;
}

export interface LoginSuccess {
  id: string;
  name: string;
}

export const loginRequest = (payload: LoginRequest) => ({
  type: AUTH_SAGA_ACTIONS.LOGIN_REQUEST,
  payload,
});

export const registerRequest = (payload: RegisterRequest) => ({
  type: AUTH_SAGA_ACTIONS.REGISTER_REQUEST,
  payload,
});

export const logoutRequest = () => ({
  type: AUTH_SAGA_ACTIONS.LOGOUT_REQUEST,
});
