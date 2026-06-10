export interface ListingItem {
  id: number;
  userId: number;
  type: 'buy' | 'sell';
  cryptoCurrency: string;
  fiatCurrency: string;
  totalAmount: number;
  remainingAmount: number;
  price: number;
  minOrderFiat: number;
  maxOrderFiat: number;
  platformFeeBase: number;
  platformFeeRate: number;
  paymentFeeBase: number;
  paymentFeeRate: number;
  paymentTimeLimit: number;
  paymentMethodId: number | null;
  status: 'active' | 'paused' | 'completed' | 'cancelled';
  createdAt: string;
}

export interface CreateListingParams {
  type: 'buy' | 'sell';
  cryptoCurrency?: string;
  fiatCurrency?: string;
  totalAmount: number;
  price: number;
  minOrderFiat: number;
  maxOrderFiat: number;
  paymentTimeLimit?: number;
  paymentMethodId?: number | null;
}

export interface ListListingsParams {
  type?: 'buy' | 'sell';
  status?: string;
  limit?: number;
  offset?: number;
}

export interface ListListingsResponse {
  list: ListingItem[];
}
