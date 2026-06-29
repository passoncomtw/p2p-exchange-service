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
   * POST /app/auth/login
   * Request:  { username, password }
   * Response: { access_token, expireIn, user: { id, account, name } }
   */
  login: async (credentials: LoginCredentials): Promise<LoginData> => {
    const response = await httpClient.post<ApiResponse<LoginData>>('/app/auth/login', {
      username: credentials.account, // App 內部用 account，後端 field 是 username
      password: credentials.password,
    });
    return response.data.data;
  },

  /**
   * 註冊（後端尚未實作，暫保留介面）
   */
  register: async (credentials: RegisterCredentials): Promise<RegisterData> => {
    const response = await httpClient.post<ApiResponse<RegisterData>>('/app/auth/register', {
      username: credentials.account,
      password: credentials.password,
      email: credentials.email,
    });
    return response.data.data;
  },

  /**
   * 登出（後端 JWT 為 stateless，清除本地 token 即可，無需呼叫後端）
   */
  logout: async (): Promise<void> => {
    return Promise.resolve();
  },
};
