import httpClient from './httpClient';
import type {
  LoginCredentials,
  RegisterCredentials,
  LoginData,
  RegisterData,
  ApiResponse,
} from '@/interfaces/store';

export const authApi = {
  /**
   * 登入
   */
  login: async (credentials: LoginCredentials): Promise<LoginData> => {
    const response = await httpClient.post<ApiResponse<LoginData>>('/auth/login', {
      account: credentials.account,
      password: credentials.password,
      notificationToken: credentials.notificationToken || 'expo_default_token',
    });
    return response.data.data;
  },

  /**
   * 註冊
   */
  register: async (credentials: RegisterCredentials): Promise<RegisterData> => {
    const response = await httpClient.post<ApiResponse<RegisterData>>('/users', {
      account: credentials.account,
      email: credentials.email,
      name: credentials.nickname,
      password: credentials.password,
      transactionCode: credentials.transactionPassword,
      referralCode: credentials.referralCode || '',
      type: 0,
    });
    return response.data.data;
  },

  /**
   * 登出
   */
  logout: async (): Promise<void> => {
    await httpClient.post<ApiResponse<void>>('/auth/logout', {});
  },
};

