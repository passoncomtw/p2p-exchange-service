import httpClient, { httpClientWithAuth } from './httpClient';
import type { ApiResponse } from '@/interfaces/store';
import type {
  ListingItem,
  CreateListingParams,
  ListListingsParams,
  ListListingsResponse,
} from '@/interfaces/listing';

export const listingsApi = {
  list: async (params?: ListListingsParams): Promise<ListingItem[]> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<{ data: ListListingsResponse }>>(
      '/app/listings',
      { params }
    );
    return response.data.data.data.list;
  },

  getById: async (id: number): Promise<ListingItem> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<{ data: ListingItem }>>(
      `/app/listings/${id}`
    );
    return response.data.data.data;
  },

  create: async (params: CreateListingParams): Promise<{ id: number }> => {
    const response = await httpClientWithAuth.postWithToken<ApiResponse<{ data: { id: number } }>>(
      '/app/listings',
      {
        ...params,
        cryptoCurrency: params.cryptoCurrency ?? 'USDT',
        fiatCurrency: params.fiatCurrency ?? 'TWD',
        paymentTimeLimit: params.paymentTimeLimit ?? 900,
      }
    );
    return response.data.data.data;
  },

  mine: async (params?: ListListingsParams): Promise<ListingItem[]> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<{ data: ListListingsResponse }>>(
      '/app/listings/mine',
      { params }
    );
    return response.data.data.data.list;
  },

  cancel: async (id: number): Promise<void> => {
    await httpClientWithAuth.putWithToken(`/app/listings/${id}/cancel`);
  },
};
