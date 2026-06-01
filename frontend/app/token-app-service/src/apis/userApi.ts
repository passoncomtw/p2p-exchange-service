import { httpClientWithAuth } from './httpClient';
import type { ApiResponse, User } from '@/interfaces/store';

export const userApi = {
  /**
   * 取得用戶資訊
   * 需要認證 token
   */
  getUser: async (userId: number): Promise<User> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<User>>(`/users/${userId}`);
    return response.data.data;
  },
};

