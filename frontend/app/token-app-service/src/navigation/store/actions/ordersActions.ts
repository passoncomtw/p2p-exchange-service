import type { CreatePendingOrderRequest, CreateOrderRequest, GetOrdersParams } from '@/apis/ordersApi';

/**
 * Orders Actions
 * 定義掛單和訂單相關的 action creators
 */

export const ORDERS_ACTIONS = {
  FETCH_PENDING_ORDERS_REQUEST: 'orders/fetchPendingOrdersRequest',
  CREATE_PENDING_ORDER_REQUEST: 'orders/createPendingOrderRequest',
  DELETE_PENDING_ORDER_REQUEST: 'orders/deletePendingOrderRequest',
  CREATE_ORDER_REQUEST: 'orders/createOrderRequest',
  FETCH_ORDER_LIST_REQUEST: 'orders/fetchOrderListRequest',
  MARK_ORDER_AS_PAID_REQUEST: 'orders/markOrderAsPaidRequest',
  APPLY_ORDER_REQUEST: 'orders/applyOrderRequest',
} as const;

/**
 * 建立掛單請求的參數
 */
export interface CreatePendingOrderPayload {
  data: CreatePendingOrderRequest;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

/**
 * 刪除掛單請求的參數
 */
export interface DeletePendingOrderPayload {
  orderId: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

/**
 * 建立訂單請求的參數
 */
export interface CreateOrderPayload {
  data: CreateOrderRequest;
  onSuccess?: (orderId: string) => void;
  onError?: (error: string) => void;
}

/**
 * 請求取得掛單列表
 */
export const fetchPendingOrdersRequest = () => ({
  type: ORDERS_ACTIONS.FETCH_PENDING_ORDERS_REQUEST,
});

/**
 * 請求建立掛單
 */
export const createPendingOrderRequest = (payload: CreatePendingOrderPayload) => ({
  type: ORDERS_ACTIONS.CREATE_PENDING_ORDER_REQUEST,
  payload,
});

/**
 * 請求刪除掛單
 */
export const deletePendingOrderRequest = (payload: DeletePendingOrderPayload) => ({
  type: ORDERS_ACTIONS.DELETE_PENDING_ORDER_REQUEST,
  payload,
});

/**
 * 請求建立訂單（從掛單建立訂單）
 */
export const createOrderRequest = (payload: CreateOrderPayload) => ({
  type: ORDERS_ACTIONS.CREATE_ORDER_REQUEST,
  payload,
});

/**
 * 請求取得訂單列表
 */
export const fetchOrderListRequest = (params?: GetOrdersParams) => ({
  type: ORDERS_ACTIONS.FETCH_ORDER_LIST_REQUEST,
  payload: params,
});

/**
 * 標記訂單為已付款請求的參數
 */
export interface MarkOrderAsPaidPayload {
  orderId: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

/**
 * 請求標記訂單為已付款
 */
export const markOrderAsPaidRequest = (payload: MarkOrderAsPaidPayload) => ({
  type: ORDERS_ACTIONS.MARK_ORDER_AS_PAID_REQUEST,
  payload,
});

/**
 * 放行訂單請求的參數
 */
export interface ApplyOrderPayload {
  orderId: string;
  onSuccess?: () => void;
  onError?: (error: string) => void;
}

/**
 * 請求放行訂單
 */
export const applyOrderRequest = (payload: ApplyOrderPayload) => ({
  type: ORDERS_ACTIONS.APPLY_ORDER_REQUEST,
  payload,
});
