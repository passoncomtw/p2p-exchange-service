import { httpClientWithAuth } from './httpClient';
import type { ApiResponse } from '@/interfaces/store';

export interface PaymentMethodItem {
  id: number;
  type: string;
  bankName: string;
  accountName: string;
  accountNumber: string;
  isActive: boolean;
}

interface ListPaymentMethodsResponse {
  list: PaymentMethodItem[];
}

interface CreatePaymentMethodParams {
  type: 'bank_transfer';
  bankName: string;
  accountName: string;
  accountNumber: string;
}

export const paymentMethodsApi = {
  list: async (): Promise<PaymentMethodItem[]> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<{ data: ListPaymentMethodsResponse }>>(
      '/app/payment-methods'
    );
    return response.data.data.data.list;
  },

  create: async (params: CreatePaymentMethodParams): Promise<{ id: number }> => {
    const response = await httpClientWithAuth.postWithToken<ApiResponse<{ data: { id: number } }>>(
      '/app/payment-methods',
      params
    );
    return response.data.data.data;
  },

  remove: async (id: number): Promise<void> => {
    await httpClientWithAuth.deleteWithToken(`/app/payment-methods/${id}`);
  },
};
