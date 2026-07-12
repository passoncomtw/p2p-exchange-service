import { httpClientWithAuth } from './httpClient';
import type { ApiResponse } from '@/interfaces/store';
import type {
  ListWalletsResponse,
  ListWalletLedgersParams,
  ListWalletLedgersResponse,
} from '@/interfaces/wallet';

export const walletApi = {
  listWallets: async (): Promise<ListWalletsResponse> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<ListWalletsResponse>>(
      '/app/wallets'
    );
    return response.data.data;
  },

  listLedgers: async (params: ListWalletLedgersParams): Promise<ListWalletLedgersResponse> => {
    const { currency, limit = 20, offset = 0 } = params;
    const response = await httpClientWithAuth.getWithToken<ApiResponse<ListWalletLedgersResponse>>(
      `/app/wallets/${currency}/ledgers`,
      { params: { limit, offset } }
    );
    return response.data.data;
  },
};
