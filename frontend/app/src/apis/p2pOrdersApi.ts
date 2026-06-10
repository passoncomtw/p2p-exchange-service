import { httpClientWithAuth } from './httpClient';
import type { ApiResponse } from '@/interfaces/store';
import type {
  Order,
  ListOrdersParams,
  ListOrdersResponse,
  CreateOrderParams,
} from '@/interfaces/order';

export const p2pOrdersApi = {
  list: async (params?: ListOrdersParams): Promise<Order[]> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<{ data: ListOrdersResponse }>>(
      '/app/orders',
      { params }
    );
    return response.data.data.data.list;
  },

  getById: async (id: number): Promise<Order> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<{ data: Order }>>(
      `/app/orders/${id}`
    );
    return response.data.data.data;
  },

  create: async (params: CreateOrderParams): Promise<{ id: number; orderNo: string }> => {
    const response = await httpClientWithAuth.postWithToken<ApiResponse<{ data: { id: number; orderNo: string } }>>(
      '/app/orders',
      params
    );
    return response.data.data.data;
  },

  markPaid: async (id: number): Promise<void> => {
    await httpClientWithAuth.putWithToken(`/app/orders/${id}/pay`);
  },

  confirm: async (id: number): Promise<void> => {
    await httpClientWithAuth.putWithToken(`/app/orders/${id}/confirm`);
  },

  cancel: async (id: number, reason: string): Promise<void> => {
    await httpClientWithAuth.putWithToken(`/app/orders/${id}/cancel`, { reason });
  },

  dispute: async (id: number, reason: string): Promise<void> => {
    await httpClientWithAuth.putWithToken(`/app/orders/${id}/dispute`, { reason });
  },
};
