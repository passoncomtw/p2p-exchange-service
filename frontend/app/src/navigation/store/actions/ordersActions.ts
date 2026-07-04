import type { ListOrdersParams, CreateOrderParams } from '@/interfaces/order';
import type { CreateListingParams } from '@/interfaces/listing';

export const ORDERS_ACTIONS = {
  CREATE_LISTING_REQUEST: 'orders/createListingRequest',
  CREATE_ORDER_REQUEST: 'orders/createOrderRequest',
  FETCH_ORDER_LIST_REQUEST: 'orders/fetchOrderListRequest',
  MARK_ORDER_AS_PAID_REQUEST: 'orders/markOrderAsPaidRequest',
  APPLY_ORDER_REQUEST: 'orders/applyOrderRequest',
} as const;

export interface CreateListingPayload {
  data: CreateListingParams;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

export interface CreateOrderPayload {
  data: CreateOrderParams;
  onSuccess?: (orderId: string) => void;
  onError?: (error: string) => void;
}

export interface MarkOrderAsPaidPayload {
  orderId: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

export interface ApplyOrderPayload {
  orderId: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

export const createListingRequest = (payload: CreateListingPayload) => ({
  type: ORDERS_ACTIONS.CREATE_LISTING_REQUEST,
  payload,
});

export const createOrderRequest = (payload: CreateOrderPayload) => ({
  type: ORDERS_ACTIONS.CREATE_ORDER_REQUEST,
  payload,
});

export const fetchOrderListRequest = (params?: ListOrdersParams) => ({
  type: ORDERS_ACTIONS.FETCH_ORDER_LIST_REQUEST,
  payload: params,
});

export const markOrderAsPaidRequest = (payload: MarkOrderAsPaidPayload) => ({
  type: ORDERS_ACTIONS.MARK_ORDER_AS_PAID_REQUEST,
  payload,
});

export const applyOrderRequest = (payload: ApplyOrderPayload) => ({
  type: ORDERS_ACTIONS.APPLY_ORDER_REQUEST,
  payload,
});
