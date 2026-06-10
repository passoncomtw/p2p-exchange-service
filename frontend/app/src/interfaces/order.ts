export interface Order {
  id: number;
  orderNo: string;
  listingId: number;
  listingType: 'buy' | 'sell';
  sellerId: number;
  buyerId: number;
  cryptoCurrency: string;
  fiatCurrency: string;
  cryptoAmount: number;
  price: number;
  fiatAmount: number;
  platformFeeBase: number;
  platformFeeAmount: number;
  paymentFeeBase: number;
  paymentFeeAmount: number;
  totalFee: number;
  totalAmount: number;
  paymentMethodId: number;
  status: 'matched' | 'paid' | 'releasing' | 'completed' | 'cancelled' | 'timeout' | 'disputed';
  paymentDeadline: string;
  paidAt: string | null;
  confirmedAt: string | null;
  completedAt: string | null;
  cancelledAt: string | null;
  cancelReason: string | null;
  createdAt: string;
}

export interface ListOrdersParams {
  role?: 'buyer' | 'seller';
  status?: string;
  limit?: number;
  offset?: number;
}

export interface ListOrdersResponse {
  list: Order[];
}

export interface CreateOrderParams {
  listingId: number;
  cryptoAmount: number;
}
