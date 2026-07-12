export interface WalletItem {
  currency: string;
  available_balance: string;
  frozen_balance: string;
}

export interface WalletLedgerItem {
  type: string;
  amount: string;
  balance_after: string;
  ref_order_no: string | null;
  created_at: string;
}

export interface ListWalletsResponse {
  list: WalletItem[];
}

export interface ListWalletLedgersParams {
  currency: string;
  limit?: number;
  offset?: number;
}

export interface ListWalletLedgersResponse {
  list: WalletLedgerItem[];
  total: number;
}
