import type { GetPendingOrdersParams } from '@/apis/ordersApi';

/**
 * Market Actions
 * 定義市場掛單列表相關的 action creators
 */

export const MARKET_ACTIONS = {
  FETCH_BUY_ORDERS_REQUEST: 'market/fetchBuyOrdersRequest',
  FETCH_SELL_ORDERS_REQUEST: 'market/fetchSellOrdersRequest',
} as const;

/**
 * 請求取得買幣掛單列表（我要賣時顯示）
 */
export const fetchBuyOrdersRequest = (params?: GetPendingOrdersParams) => ({
  type: MARKET_ACTIONS.FETCH_BUY_ORDERS_REQUEST,
  payload: params,
});

/**
 * 請求取得賣幣掛單列表（我要買時顯示）
 */
export const fetchSellOrdersRequest = (params?: GetPendingOrdersParams) => ({
  type: MARKET_ACTIONS.FETCH_SELL_ORDERS_REQUEST,
  payload: params,
});

