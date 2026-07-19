import { httpClientWithAuth } from './httpClient';
import type { ApiResponse } from '@/interfaces/store';
import type {
  ListWalletsResponse,
  ListWalletLedgersParams,
  ListWalletLedgersResponse,
  CryptoDepositInfo,
  CryptoWithdrawRequest,
  CryptoWithdrawResponse,
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

  getCryptoDepositInfo: async (): Promise<CryptoDepositInfo> => {
    const response = await httpClientWithAuth.getWithToken<ApiResponse<CryptoDepositInfo>>(
      '/app/wallets/crypto/deposit-info'
    );
    return response.data.data;
  },

  cryptoWithdraw: async (req: CryptoWithdrawRequest): Promise<CryptoWithdrawResponse> => {
    const response = await httpClientWithAuth.postWithToken<ApiResponse<CryptoWithdrawResponse>>(
      '/app/wallets/crypto/withdraw',
      req
    );
    return response.data.data;
  },
};
