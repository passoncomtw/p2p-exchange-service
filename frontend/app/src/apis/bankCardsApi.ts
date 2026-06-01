import { httpClientWithAuth } from './httpClient';
import type { ApiResponse } from '@/interfaces/store';

/**
 * 銀行資訊
 */
export interface Bank {
  id: number;
  bankCode: string;
  bankName: string;
}

/**
 * 銀行卡資訊
 * 注意：此接口定義 GET /bankcards API 返回的數據格式
 */
export interface BankCard {
  id: number;
  bankId: number;
  name: string;
  cardNumber: string;
  branchName: string; // 必需，銀行卡必定有分行名稱
  status: number;
  createdAt: string;
  bank: Bank;
}

export const bankCardsApi = {
  /**
   * 取得使用者的銀行卡列表
   * 需要認證 token
   */
  getBankCards: async (): Promise<BankCard[]> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<BankCard[]>>('/bankcards');
    return response.data.data;
  },
};

