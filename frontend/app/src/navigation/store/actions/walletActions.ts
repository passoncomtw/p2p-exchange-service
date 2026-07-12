import type { ListWalletLedgersParams } from '@/interfaces/wallet';

export const WALLET_ACTIONS = {
  FETCH_WALLETS_REQUEST: 'wallet/fetchWalletsRequest',
  FETCH_LEDGERS_REQUEST: 'wallet/fetchLedgersRequest',
} as const;

export const fetchWalletsRequest = () => ({
  type: WALLET_ACTIONS.FETCH_WALLETS_REQUEST,
});

export const fetchLedgersRequest = (params: ListWalletLedgersParams) => ({
  type: WALLET_ACTIONS.FETCH_LEDGERS_REQUEST,
  payload: params,
});
