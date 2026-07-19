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

export interface CryptoDepositInfo {
  address: string;
  memo: string;
  network: string;
  currency: string;
  contract_address: string;
}

export interface CryptoWithdrawRequest {
  to_address: string;
  amount: string;
}

export interface CryptoWithdrawResponse {
  id: number;
  status: string;
}

export interface FiatDepositRequest {
  amount: number;
}

export interface FiatDepositResponse {
  merchant_trade_no: string;
  payment_url: string;
  form_params: Record<string, string>;
}

export interface FiatWithdrawRequest {
  amount: string;
  bank_code: string;
  bank_account: string;
  account_name: string;
}

export interface FiatWithdrawResponse {
  id: number;
  status: string;
}
