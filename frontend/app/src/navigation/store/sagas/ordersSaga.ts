import { call, put, takeLatest } from 'redux-saga/effects';
import { SagaIterator } from 'redux-saga';
import { PayloadAction } from '@reduxjs/toolkit';
import logger from '@pkg/logger';
import { handleSagaError } from '@pkg/utils/sagaHelpers';
import { listingsApi, p2pOrdersApi } from '@/apis';
import { ORDERS_ACTIONS, CreateListingPayload, CreateOrderPayload, FetchOrderDetailPayload, MarkOrderAsPaidPayload, ApplyOrderPayload, CancelOrderPayload, DisputeOrderPayload } from '../actions/ordersActions';
import {
  createOrderStart,
  createOrderSuccess,
  createOrderFailure,
  createTransactionOrderStart,
  createTransactionOrderSuccess,
  createTransactionOrderFailure,
  fetchOrderListStart,
  fetchOrderListSuccess,
  fetchOrderListFailure,
  fetchOrderDetailStart,
  fetchOrderDetailSuccess,
  fetchOrderDetailFailure,
  markOrderAsPaidStart,
  markOrderAsPaidSuccess,
  markOrderAsPaidFailure,
  applyOrderStart,
  applyOrderSuccess,
  applyOrderFailure,
  cancelOrderStart,
  cancelOrderSuccess,
  cancelOrderFailure,
  disputeOrderStart,
  disputeOrderSuccess,
  disputeOrderFailure,
} from '../slices/ordersSlice';

function* createListingSaga(action: PayloadAction<CreateListingPayload>): SagaIterator {
  const { data, onSuccess, onError } = action.payload;

  yield put(createOrderStart());

  try {
    const result = yield call(listingsApi.create, data);

    logger.info('建立掛單成功', { id: result.id, type: data.type });

    yield put(createOrderSuccess(result));

    if (onSuccess) onSuccess();
  } catch (error: any) {
    const errorResult = handleSagaError(error, '建立掛單失敗', { type: data.type });
    yield put(createOrderFailure(errorResult));
    if (onError) onError(errorResult.message);
  }
}

function* createOrderSaga(action: PayloadAction<CreateOrderPayload>): SagaIterator {
  const { data, onSuccess, onError } = action.payload;

  yield put(createTransactionOrderStart());

  try {
    const order = yield call(p2pOrdersApi.create, data);

    logger.info('建立訂單成功', { orderId: order.id, listingId: data.listingId });

    yield put(createTransactionOrderSuccess());

    if (onSuccess) onSuccess(String(order.id));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '建立訂單失敗', { listingId: data.listingId });
    yield put(createTransactionOrderFailure(errorResult));
    if (onError) onError(errorResult.message);
  }
}

function* fetchOrderListSaga(action: PayloadAction<any>): SagaIterator {
  yield put(fetchOrderListStart());

  try {
    const params = action.payload || { limit: 100 };
    const orders = yield call(p2pOrdersApi.list, params);

    logger.info('取得訂單列表成功', { count: orders.length });

    yield put(fetchOrderListSuccess(orders));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得訂單列表失敗');
    yield put(fetchOrderListFailure(errorResult));
  }
}

function* fetchOrderDetailSaga(action: PayloadAction<FetchOrderDetailPayload>): SagaIterator {
  const { orderId } = action.payload;

  yield put(fetchOrderDetailStart());

  try {
    const order = yield call(p2pOrdersApi.getById, Number(orderId));

    logger.info('取得訂單詳情成功', { orderId });

    yield put(fetchOrderDetailSuccess(order));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得訂單詳情失敗', { orderId });
    yield put(fetchOrderDetailFailure());
    logger.error('fetchOrderDetailSaga 失敗', errorResult);
  }
}

function* markOrderAsPaidSaga(action: PayloadAction<MarkOrderAsPaidPayload>): SagaIterator {
  const { orderId, onSuccess, onError } = action.payload;

  yield put(markOrderAsPaidStart());

  try {
    yield call(p2pOrdersApi.markPaid, Number(orderId));

    logger.info('標記訂單為已付款成功', { orderId });

    yield put(markOrderAsPaidSuccess(orderId));

    if (onSuccess) onSuccess();
  } catch (error: any) {
    const errorResult = handleSagaError(error, '標記訂單為已付款失敗', { orderId });
    yield put(markOrderAsPaidFailure(errorResult));
    if (onError) onError(errorResult.message);
  }
}

function* applyOrderSaga(action: PayloadAction<ApplyOrderPayload>): SagaIterator {
  const { orderId, onSuccess, onError } = action.payload;

  yield put(applyOrderStart());

  try {
    yield call(p2pOrdersApi.confirm, Number(orderId));

    logger.info('確認放行訂單成功', { orderId });

    yield put(applyOrderSuccess(orderId));

    if (onSuccess) onSuccess();
  } catch (error: any) {
    const errorResult = handleSagaError(error, '確認放行訂單失敗', { orderId });
    yield put(applyOrderFailure(errorResult));
    if (onError) onError(errorResult.message);
  }
}

function* cancelOrderSaga(action: PayloadAction<CancelOrderPayload>): SagaIterator {
  const { orderId, reason, onSuccess, onError } = action.payload;

  yield put(cancelOrderStart());

  try {
    yield call(p2pOrdersApi.cancel, Number(orderId), reason);

    logger.info('取消訂單成功', { orderId });

    yield put(cancelOrderSuccess(orderId));

    if (onSuccess) onSuccess();
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取消訂單失敗', { orderId });
    yield put(cancelOrderFailure(errorResult));
    if (onError) onError(errorResult.message);
  }
}

export function* watchOrdersSagas(): SagaIterator {
  yield takeLatest(ORDERS_ACTIONS.CREATE_LISTING_REQUEST, createListingSaga);
  yield takeLatest(ORDERS_ACTIONS.CREATE_ORDER_REQUEST, createOrderSaga);
  yield takeLatest(ORDERS_ACTIONS.FETCH_ORDER_LIST_REQUEST, fetchOrderListSaga);
  yield takeLatest(ORDERS_ACTIONS.FETCH_ORDER_DETAIL_REQUEST, fetchOrderDetailSaga);
  yield takeLatest(ORDERS_ACTIONS.MARK_ORDER_AS_PAID_REQUEST, markOrderAsPaidSaga);
  yield takeLatest(ORDERS_ACTIONS.APPLY_ORDER_REQUEST, applyOrderSaga);
  yield takeLatest(ORDERS_ACTIONS.CANCEL_ORDER_REQUEST, cancelOrderSaga);
  yield takeLatest(ORDERS_ACTIONS.DISPUTE_ORDER_REQUEST, disputeOrderSaga);
}

function* disputeOrderSaga(action: PayloadAction<DisputeOrderPayload>): SagaIterator {
  const { orderId, reason, onSuccess, onError } = action.payload;

  yield put(disputeOrderStart());

  try {
    yield call(p2pOrdersApi.dispute, Number(orderId), reason);

    logger.info('申訴訂單成功', { orderId });

    yield put(disputeOrderSuccess(orderId));

    if (onSuccess) onSuccess();
  } catch (error: any) {
    const errorResult = handleSagaError(error, '申訴訂單失敗', { orderId });
    yield put(disputeOrderFailure(errorResult));
    if (onError) onError(errorResult.message);
  }
}
