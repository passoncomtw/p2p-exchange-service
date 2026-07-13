import type { ListOrdersParams, CreateOrderParams } from '@/interfaces/order';
import type { CreateListingParams } from '@/interfaces/listing';

export const ORDERS_ACTIONS = {
  CREATE_LISTING_REQUEST: 'orders/createListingRequest',
  CREATE_ORDER_REQUEST: 'orders/createOrderRequest',
  FETCH_ORDER_LIST_REQUEST: 'orders/fetchOrderListRequest',
  FETCH_ORDER_DETAIL_REQUEST: 'orders/fetchOrderDetailRequest',
  MARK_ORDER_AS_PAID_REQUEST: 'orders/markOrderAsPaidRequest',
  APPLY_ORDER_REQUEST: 'orders/applyOrderRequest',
  CANCEL_ORDER_REQUEST: 'orders/cancelOrderRequest',
  DISPUTE_ORDER_REQUEST: 'orders/disputeOrderRequest',
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

export interface FetchOrderDetailPayload {
  orderId: string;
}

export interface ApplyOrderPayload {
  orderId: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

export interface CancelOrderPayload {
  orderId: string;
  reason: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

export interface DisputeOrderPayload {
  orderId: string;
  reason: string;
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

export const fetchOrderDetailRequest = (payload: FetchOrderDetailPayload) => ({
  type: ORDERS_ACTIONS.FETCH_ORDER_DETAIL_REQUEST,
  payload,
});

export const markOrderAsPaidRequest = (payload: MarkOrderAsPaidPayload) => ({
  type: ORDERS_ACTIONS.MARK_ORDER_AS_PAID_REQUEST,
  payload,
});

export const applyOrderRequest = (payload: ApplyOrderPayload) => ({
  type: ORDERS_ACTIONS.APPLY_ORDER_REQUEST,
  payload,
});

export const cancelOrderRequest = (payload: CancelOrderPayload) => ({
  type: ORDERS_ACTIONS.CANCEL_ORDER_REQUEST,
  payload,
});

export const disputeOrderRequest = (payload: DisputeOrderPayload) => ({
  type: ORDERS_ACTIONS.DISPUTE_ORDER_REQUEST,
  payload,
});
