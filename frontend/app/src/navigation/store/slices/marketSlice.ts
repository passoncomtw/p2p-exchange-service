import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { PendingOrder } from '@/apis/ordersApi';
import type { SagaErrorResult } from '@pkg/utils/sagaHelpers';

/**
 * Market State
 * 管理交易市場的掛單列表
 */
interface MarketState {
  // 買幣掛單列表（我要賣時顯示）
  buyOrders: PendingOrder[];
  buyOrdersPage: number;
  buyOrdersTotal: number;
  buyOrdersLoading: boolean;
  buyOrdersError: string | null;

  // 賣幣掛單列表（我要買時顯示）
  sellOrders: PendingOrder[];
  sellOrdersPage: number;
  sellOrdersTotal: number;
  sellOrdersLoading: boolean;
  sellOrdersError: string | null;
}

const initialState: MarketState = {
  buyOrders: [],
  buyOrdersPage: 1,
  buyOrdersTotal: 0,
  buyOrdersLoading: false,
  buyOrdersError: null,

  sellOrders: [],
  sellOrdersPage: 1,
  sellOrdersTotal: 0,
  sellOrdersLoading: false,
  sellOrdersError: null,
};

/**
 * Market Slice
 * 負責管理交易市場掛單列表的狀態
 */
const marketSlice = createSlice({
  name: 'market',
  initialState,
  reducers: {
    // 買幣掛單列表 (type=0, 我要賣時顯示)
    fetchBuyOrdersStart(state) {
      state.buyOrdersLoading = true;
      state.buyOrdersError = null;
    },
    fetchBuyOrdersSuccess(
      state,
      action: PayloadAction<{ rows: PendingOrder[]; page: number; total: number }>
    ) {
      state.buyOrdersLoading = false;
      state.buyOrders = action.payload.rows;
      state.buyOrdersPage = action.payload.page;
      state.buyOrdersTotal = action.payload.total;
      state.buyOrdersError = null;
    },
    fetchBuyOrdersFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.buyOrdersLoading = false;
      state.buyOrdersError = action.payload.message;
    },

    // 賣幣掛單列表 (type=1, 我要買時顯示)
    fetchSellOrdersStart(state) {
      state.sellOrdersLoading = true;
      state.sellOrdersError = null;
    },
    fetchSellOrdersSuccess(
      state,
      action: PayloadAction<{ rows: PendingOrder[]; page: number; total: number }>
    ) {
      state.sellOrdersLoading = false;
      state.sellOrders = action.payload.rows;
      state.sellOrdersPage = action.payload.page;
      state.sellOrdersTotal = action.payload.total;
      state.sellOrdersError = null;
    },
    fetchSellOrdersFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.sellOrdersLoading = false;
      state.sellOrdersError = action.payload.message;
    },

    // 重置市場狀態
    resetMarket() {
      return initialState;
    },
  },
});

export const {
  fetchBuyOrdersStart,
  fetchBuyOrdersSuccess,
  fetchBuyOrdersFailure,
  fetchSellOrdersStart,
  fetchSellOrdersSuccess,
  fetchSellOrdersFailure,
  resetMarket,
} = marketSlice.actions;

export default marketSlice.reducer;

