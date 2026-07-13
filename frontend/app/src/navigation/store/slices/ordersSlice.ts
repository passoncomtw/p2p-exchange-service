import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { SagaErrorResult } from '@pkg/utils/sagaHelpers';
import type { Order } from '@/interfaces/order';

interface OrdersState {
  creating: boolean;
  createError: string | null;
  creatingOrder: boolean;
  createOrderError: string | null;
  orderList: Order[];
  orderListLoading: boolean;
  orderListError: string | null;
  orderDetailLoading: boolean;
}

const initialState: OrdersState = {
  creating: false,
  createError: null,
  creatingOrder: false,
  createOrderError: null,
  orderList: [],
  orderListLoading: false,
  orderListError: null,
  orderDetailLoading: false,
};

const ordersSlice = createSlice({
  name: 'orders',
  initialState,
  reducers: {
    createOrderStart(state) {
      state.creating = true;
      state.createError = null;
    },
    createOrderSuccess(state, _action: PayloadAction<any>) {
      state.creating = false;
      state.createError = null;
    },
    createOrderFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.creating = false;
      state.createError = action.payload.message;
    },
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
    fetchOrderListStart(state) {
      state.orderListLoading = true;
      state.orderListError = null;
    },
    fetchOrderListSuccess(state, action: PayloadAction<Order[]>) {
      state.orderListLoading = false;
      state.orderList = action.payload;
      state.orderListError = null;
    },
    fetchOrderListFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.orderListLoading = false;
      state.orderListError = action.payload.message;
    },
    fetchOrderDetailStart(state) {
      state.orderDetailLoading = true;
    },
    fetchOrderDetailSuccess(state, action: PayloadAction<Order>) {
      state.orderDetailLoading = false;
      const idx = state.orderList.findIndex((o) => o.id === action.payload.id);
      if (idx >= 0) {
        state.orderList[idx] = action.payload;
      } else {
        state.orderList.push(action.payload);
      }
    },
    fetchOrderDetailFailure(state) {
      state.orderDetailLoading = false;
    },
    markOrderAsPaidStart(state) {},
    markOrderAsPaidSuccess(state, action: PayloadAction<string>) {
      const orderId = action.payload;
      const order = state.orderList.find((o) => o.id.toString() === orderId);
      if (order) {
        order.status = 'paid';
      }
    },
    markOrderAsPaidFailure(state, action: PayloadAction<SagaErrorResult>) {},
    applyOrderStart(state) {},
    applyOrderSuccess(state, action: PayloadAction<string>) {
      const orderId = action.payload;
      const order = state.orderList.find((o) => o.id.toString() === orderId);
      if (order) {
        order.status = 'completed';
      }
    },
    applyOrderFailure(state, action: PayloadAction<SagaErrorResult>) {},
    resetOrders() {
      return initialState;
    },
  },
});

export const {
  fetchOrderDetailStart,
  fetchOrderDetailSuccess,
  fetchOrderDetailFailure,
  createOrderStart,
  createOrderSuccess,
  createOrderFailure,
  createTransactionOrderStart,
  createTransactionOrderSuccess,
  createTransactionOrderFailure,
  fetchOrderListStart,
  fetchOrderListSuccess,
  fetchOrderListFailure,
  markOrderAsPaidStart,
  markOrderAsPaidSuccess,
  markOrderAsPaidFailure,
  applyOrderStart,
  applyOrderSuccess,
  applyOrderFailure,
  resetOrders,
} = ordersSlice.actions;

export default ordersSlice.reducer;
