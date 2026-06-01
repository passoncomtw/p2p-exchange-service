import { call, put, takeLatest } from 'redux-saga/effects';
import { SagaIterator } from 'redux-saga';
import { PayloadAction } from '@reduxjs/toolkit';
import logger from '@pkg/logger';
import { handleSagaError } from '@pkg/utils/sagaHelpers';
import { ordersApi } from '@/apis';
import { ORDERS_ACTIONS, CreatePendingOrderPayload, DeletePendingOrderPayload, CreateOrderPayload, MarkOrderAsPaidPayload, ApplyOrderPayload } from '../actions/ordersActions';
import {
  fetchOrdersStart,
  fetchOrdersSuccess,
  fetchOrdersFailure,
  createOrderStart,
  createOrderSuccess,
  createOrderFailure,
  deleteOrderStart,
  deleteOrderSuccess,
  deleteOrderFailure,
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
} from '../slices/ordersSlice';

/**
 * Fetch Pending Orders Saga
 * 使用 httpClient 處理取得掛單列表流程
 * 
 * 流程：
 * 1. 組件 dispatch fetchPendingOrdersRequest()
 * 2. Saga 監聽到 action，先設置 loading 狀態
 * 3. Saga 調用 API 取得資料
 * 4. 取得資料後，Saga dispatch fetchOrdersSuccess() 更新 store
 * 5. 如果有錯誤，Saga dispatch fetchOrdersFailure() 設置錯誤訊息
 */
function* fetchPendingOrdersSaga(): SagaIterator {
  // 步驟 1: 開始取得掛單（設置 loading 狀態）
  yield put(fetchOrdersStart());

  try {
    // 步驟 2: 使用 httpClient 呼叫取得掛單 API
    const ordersData = yield call(ordersApi.getPendingOrders);

    logger.info('取得掛單列表成功', {
      hasBuy: !!ordersData.buy,
      hasSell: !!ordersData.sell,
    });

    // 步驟 3: 取得資料後，更新 Redux State
    yield put(fetchOrdersSuccess(ordersData));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得掛單列表失敗');
    yield put(fetchOrdersFailure(errorResult));
  }
}

/**
 * Create Pending Order Saga
 * 使用 httpClient 處理建立掛單流程
 * 
 * 支援 onSuccess 和 onError 回調函數，讓 UI 可以在 API 完成後執行特定邏輯
 */
function* createPendingOrderSaga(action: PayloadAction<CreatePendingOrderPayload>): SagaIterator {
  const { data, onSuccess, onError } = action.payload;

  // 步驟 1: 開始建立掛單（設置 creating 狀態）
  yield put(createOrderStart());

  try {
    // 步驟 2: 使用 httpClient 呼叫建立掛單 API
    const orderData = yield call(ordersApi.createPendingOrder, data);

    logger.info('建立掛單成功', {
      orderId: orderData.id,
      type: data.type === 0 ? '買幣' : '賣幣',
      amount: data.amount,
    });

    // 步驟 3: 建立成功後，更新 Redux State
    yield put(createOrderSuccess(orderData));

    // 步驟 4: 重新取得掛單列表以確保資料同步
    yield put({ type: ORDERS_ACTIONS.FETCH_PENDING_ORDERS_REQUEST });

    // 步驟 5: 調用成功回調（如果有提供）
    if (onSuccess) {
      onSuccess();
    }
  } catch (error: any) {
    const orderType = data.type === 0 ? '買幣' : '賣幣';
    const errorResult = handleSagaError(error, '建立掛單失敗', { type: orderType });
    
    yield put(createOrderFailure(errorResult));
    
    if (onError) {
      onError(errorResult.message);
    }
  }
}

/**
 * Delete Pending Order Saga
 * 使用 httpClient 處理刪除掛單流程
 * 
 * 支援 onSuccess 和 onError 回調函數，讓 UI 可以在 API 完成後執行特定邏輯
 */
function* deletePendingOrderSaga(action: PayloadAction<DeletePendingOrderPayload>): SagaIterator {
  const { orderId, onSuccess, onError } = action.payload;

  // 步驟 1: 開始刪除掛單（設置 deleting 狀態）
  yield put(deleteOrderStart());

  try {
    // 步驟 2: 使用 httpClient 呼叫刪除掛單 API
    yield call(ordersApi.deletePendingOrder, orderId);

    logger.info('刪除掛單成功', {
      orderId,
    });

    // 步驟 3: 刪除成功後，更新 Redux State
    yield put(deleteOrderSuccess(orderId));

    // 步驟 4: 重新取得掛單列表以確保資料同步
    yield put({ type: ORDERS_ACTIONS.FETCH_PENDING_ORDERS_REQUEST });

    // 步驟 5: 調用成功回調（如果有提供）
    if (onSuccess) {
      onSuccess();
    }
  } catch (error: any) {
    const errorResult = handleSagaError(error, '刪除掛單失敗', { orderId });
    
    yield put(deleteOrderFailure(errorResult));
    
    if (onError) {
      onError(errorResult.message);
    }
  }
}

/**
 * Create Order Saga
 * 從掛單建立訂單
 * 
 * 支援 onSuccess 和 onError 回調函數，讓 UI 可以在 API 完成後執行特定邏輯
 */
function* createOrderSaga(action: PayloadAction<CreateOrderPayload>): SagaIterator {
  const { data, onSuccess, onError } = action.payload;

  // 步驟 1: 開始建立訂單（設置 creatingOrder 狀態）
  yield put(createTransactionOrderStart());

  try {
    // 步驟 2: 使用 httpClient 呼叫建立訂單 API
    const order = yield call(ordersApi.createOrder, data);

    logger.info('建立訂單成功', {
      orderId: order.id,
      pendingOrderId: data.orderId,
      amount: data.amount,
    });

    // 步驟 3: 建立成功後，更新 Redux State
    yield put(createTransactionOrderSuccess());

    // 步驟 4: 調用成功回調（如果有提供）
    if (onSuccess) {
      onSuccess(order.id);
    }
  } catch (error: any) {
    const errorResult = handleSagaError(error, '建立訂單失敗', { pendingOrderId: data.orderId });
    
    yield put(createTransactionOrderFailure(errorResult));
    
    if (onError) {
      onError(errorResult.message);
    }
  }
}

/**
 * Fetch Order List Saga
 * 取得訂單列表
 */
function* fetchOrderListSaga(action: PayloadAction<any>): SagaIterator {
  yield put(fetchOrderListStart());

  try {
    const params = action.payload || { size: 10, page: 1 };
    const orderListData = yield call(ordersApi.getOrders, params);

    logger.info('取得訂單列表成功', {
      count: orderListData.rows.length,
      total: orderListData.total,
      page: orderListData.page,
    });

    yield put(fetchOrderListSuccess(orderListData));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得訂單列表失敗');
    yield put(fetchOrderListFailure(errorResult));
  }
}

/**
 * Mark Order As Paid Saga
 * 標記訂單為已付款
 * 
 * 支援 onSuccess 和 onError 回調函數，讓 UI 可以在 API 完成後執行特定邏輯
 */
function* markOrderAsPaidSaga(action: PayloadAction<MarkOrderAsPaidPayload>): SagaIterator {
  const { orderId, onSuccess, onError } = action.payload;

  // 步驟 1: 開始標記已付款（設置 loading 狀態）
  yield put(markOrderAsPaidStart());

  try {
    // 步驟 2: 使用 httpClient 呼叫標記已付款 API
    yield call(ordersApi.markOrderAsPaid, orderId);

    logger.info('標記訂單為已付款成功', {
      orderId,
    });

    // 步驟 3: 更新 Redux State 中該訂單的狀態為 1（已付款等待放行）
    yield put(markOrderAsPaidSuccess(orderId));

    // 步驟 4: 調用成功回調（如果有提供）
    if (onSuccess) {
      onSuccess();
    }
  } catch (error: any) {
    const errorResult = handleSagaError(error, '標記訂單為已付款失敗', { orderId });
    
    yield put(markOrderAsPaidFailure(errorResult));
    
    if (onError) {
      onError(errorResult.message);
    }
  }
}

/**
 * Apply Order Saga
 * 賣家確認收款並放行訂單
 * 
 * 支援 onSuccess 和 onError 回調函數，讓 UI 可以在 API 完成後執行特定邏輯
 */
function* applyOrderSaga(action: PayloadAction<ApplyOrderPayload>): SagaIterator {
  const { orderId, onSuccess, onError } = action.payload;

  // 步驟 1: 開始放行訂單（設置 loading 狀態）
  yield put(applyOrderStart());

  try {
    // 步驟 2: 使用 httpClient 呼叫放行訂單 API
    yield call(ordersApi.applyOrder, orderId);

    logger.info('放行訂單成功', {
      orderId,
    });

    // 步驟 3: 更新 Redux State 中該訂單的狀態為 2（已放行）
    yield put(applyOrderSuccess(orderId));

    // 步驟 4: 調用成功回調（如果有提供）
    if (onSuccess) {
      onSuccess();
    }
  } catch (error: any) {
    const errorResult = handleSagaError(error, '放行訂單失敗', { orderId });
    
    yield put(applyOrderFailure(errorResult));
    
    if (onError) {
      onError(errorResult.message);
    }
  }
}

/**
 * Watcher Saga
 * 監聽特定的 action 並觸發對應的 saga
 */
export function* watchOrdersSagas(): SagaIterator {
  // takeLatest: 如果有多個請求，只處理最新的一個
  yield takeLatest(ORDERS_ACTIONS.FETCH_PENDING_ORDERS_REQUEST, fetchPendingOrdersSaga);
  yield takeLatest(ORDERS_ACTIONS.CREATE_PENDING_ORDER_REQUEST, createPendingOrderSaga);
  yield takeLatest(ORDERS_ACTIONS.DELETE_PENDING_ORDER_REQUEST, deletePendingOrderSaga);
  yield takeLatest(ORDERS_ACTIONS.CREATE_ORDER_REQUEST, createOrderSaga);
  yield takeLatest(ORDERS_ACTIONS.FETCH_ORDER_LIST_REQUEST, fetchOrderListSaga);
  yield takeLatest(ORDERS_ACTIONS.MARK_ORDER_AS_PAID_REQUEST, markOrderAsPaidSaga);
  yield takeLatest(ORDERS_ACTIONS.APPLY_ORDER_REQUEST, applyOrderSaga);
}

