import { httpClientWithAuth } from './httpClient';
import type { ApiResponse } from '@/interfaces/store';
import type {
  ListWalletsResponse,
  ListWalletLedgersParams,
  ListWalletLedgersResponse,
  CryptoDepositInfo,
  CryptoWithdrawRequest,
  CryptoWithdrawResponse,
  FiatDepositRequest,
  FiatDepositResponse,
  FiatWithdrawRequest,
  FiatWithdrawResponse,
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

  fiatDeposit: async (req: FiatDepositRequest): Promise<FiatDepositResponse> => {
    const response = await httpClientWithAuth.postWithToken<ApiResponse<FiatDepositResponse>>(
      '/app/wallets/fiat/deposit',
      req
    );
    return response.data.data;
  },

  fiatWithdraw: async (req: FiatWithdrawRequest): Promise<FiatWithdrawResponse> => {
    const response = await httpClientWithAuth.postWithToken<ApiResponse<FiatWithdrawResponse>>(
      '/app/wallets/fiat/withdraw',
      req
    );
    return response.data.data;
  },
};
