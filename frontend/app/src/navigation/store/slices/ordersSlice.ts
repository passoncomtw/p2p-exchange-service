import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { UserPendingOrdersResponse, OrderListResponse } from '@/apis/ordersApi';
import type { SagaErrorResult } from '@pkg/utils/sagaHelpers';
import type { OrderItem } from '@/interfaces';

/**
 * Orders State
 */
interface OrdersState {
  buy: any | null; // PendingOrder | null
  sell: any | null; // PendingOrder | null
  loading: boolean;
  error: string | null;
  creating: boolean; // 建立掛單中
  createError: string | null; // 建立掛單錯誤
  deleting: boolean; // 刪除掛單中
  deleteError: string | null; // 刪除掛單錯誤
  creatingOrder: boolean; // 建立訂單中
  createOrderError: string | null; // 建立訂單錯誤
  // 訂單列表相關
  orderList: OrderItem[];
  orderListPage: number;
  orderListSize: number;
  orderListTotal: number;
  orderListLoading: boolean;
  orderListError: string | null;
}

const initialState: OrdersState = {
  buy: null,
  sell: null,
  loading: false,
  error: null,
  creating: false,
  createError: null,
  deleting: false,
  deleteError: null,
  creatingOrder: false,
  createOrderError: null,
  // 訂單列表初始狀態
  orderList: [],
  orderListPage: 1,
  orderListSize: 10,
  orderListTotal: 0,
  orderListLoading: false,
  orderListError: null,
};

/**
 * Orders Slice
 * 負責管理掛單相關的狀態
 */
const ordersSlice = createSlice({
  name: 'orders',
  initialState,
  reducers: {
    fetchOrdersStart(state) {
      state.loading = true;
      state.error = null;
    },
    fetchOrdersSuccess(state, action: PayloadAction<UserPendingOrdersResponse>) {
      state.loading = false;
      state.buy = action.payload.buy;
      state.sell = action.payload.sell;
      state.error = null;
    },
    fetchOrdersFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.loading = false;
      state.error = action.payload.message;
    },
    clearOrdersError(state) {
      state.error = null;
    },
    createOrderStart(state) {
      state.creating = true;
      state.createError = null;
    },
    createOrderSuccess(state, action: PayloadAction<any>) {
      state.creating = false;
      // 根據類型更新對應的掛單
      if (action.payload.type === 0) {
        state.buy = action.payload;
      } else {
        state.sell = action.payload;
      }
      state.createError = null;
    },
    createOrderFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.creating = false;
      state.createError = action.payload.message;
    },
    clearCreateError(state) {
      state.createError = null;
    },
    deleteOrderStart(state) {
      state.deleting = true;
      state.deleteError = null;
    },
    deleteOrderSuccess(state, action: PayloadAction<string>) {
      state.deleting = false;
      // 根據 orderId 清除對應的掛單
      const orderId = action.payload;
      if (state.buy?.id === orderId) {
        state.buy = null;
      }
      if (state.sell?.id === orderId) {
        state.sell = null;
      }
      state.deleteError = null;
    },
    deleteOrderFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.deleting = false;
      state.deleteError = action.payload.message;
    },
    clearDeleteError(state) {
      state.deleteError = null;
    },
    // 建立訂單相關 reducers
    createTransactionOrderStart(state) {
      state.creatingOrder = true;
      state.createOrderError = null;
    },
    createTransactionOrderSuccess(state) {
      state.creatingOrder = false;
      state.createOrderError = null;
    },
    createTransactionOrderFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.creatingOrder = false;
      state.createOrderError = action.payload.message;
    },
    clearCreateOrderError(state) {
      state.createOrderError = null;
    },
    // 訂單列表相關 reducers
    fetchOrderListStart(state) {
      state.orderListLoading = true;
      state.orderListError = null;
    },
    fetchOrderListSuccess(state, action: PayloadAction<OrderListResponse>) {
      state.orderListLoading = false;
      state.orderList = action.payload.rows;
      state.orderListPage = action.payload.page;
      state.orderListSize = action.payload.size;
      state.orderListTotal = action.payload.total;
      state.orderListError = null;
    },
    fetchOrderListFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.orderListLoading = false;
      state.orderListError = action.payload.message;
    },
    clearOrderListError(state) {
      state.orderListError = null;
    },
    // 標記訂單為已付款相關 reducers
    markOrderAsPaidStart(state) {
      // 可以添加 loading 狀態，如果需要
    },
    markOrderAsPaidSuccess(state, action: PayloadAction<string>) {
      // 更新訂單列表中對應訂單的狀態為 1（已付款等待放行）
      const orderId = action.payload;
      const order = state.orderList.find((o) => o.id === orderId);
      if (order) {
        order.status = 1;
      }
    },
    markOrderAsPaidFailure(state, action: PayloadAction<SagaErrorResult>) {
      // 可以添加錯誤處理，如果需要
    },
    // 放行訂單相關 reducers
    applyOrderStart(state) {
      // 可以添加 loading 狀態，如果需要
    },
    applyOrderSuccess(state, action: PayloadAction<string>) {
      // 更新訂單列表中對應訂單的狀態為 4（已放行）
      const orderId = action.payload;
      const order = state.orderList.find((o) => o.id === orderId);
      if (order) {
        order.status = 4;
      }
    },
    applyOrderFailure(state, action: PayloadAction<SagaErrorResult>) {
      // 可以添加錯誤處理，如果需要
    },
    resetOrders() {
      // 重置為初始狀態（用於登出）
      return initialState;
    },
  },
});

export const {
  fetchOrdersStart,
  fetchOrdersSuccess,
  fetchOrdersFailure,
  clearOrdersError,
  createOrderStart,
  createOrderSuccess,
  createOrderFailure,
  clearCreateError,
  deleteOrderStart,
  deleteOrderSuccess,
  deleteOrderFailure,
  clearDeleteError,
  createTransactionOrderStart,
  createTransactionOrderSuccess,
  createTransactionOrderFailure,
  clearCreateOrderError,
  fetchOrderListStart,
  fetchOrderListSuccess,
  fetchOrderListFailure,
  clearOrderListError,
  markOrderAsPaidStart,
  markOrderAsPaidSuccess,
  markOrderAsPaidFailure,
  applyOrderStart,
  applyOrderSuccess,
  applyOrderFailure,
  resetOrders,
} = ordersSlice.actions;

export default ordersSlice.reducer;
