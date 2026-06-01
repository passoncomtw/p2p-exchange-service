import { call, put, takeLatest } from 'redux-saga/effects';
import { SagaIterator } from 'redux-saga';
import logger from '@pkg/logger';
import { handleSagaError } from '@pkg/utils/sagaHelpers';
import { bankCardsApi } from '@/apis/bankCardsApi';
import { BANKCARDS_ACTIONS } from '../actions/bankCardsActions';
import {
  fetchBankCardsStart,
  fetchBankCardsSuccess,
  fetchBankCardsFailure,
} from '../slices/bankCardsSlice';

/**
 * Fetch BankCards Saga
 * 使用 httpClient 處理取得銀行卡列表流程
 *
 * 流程：
 * 1. 組件 dispatch fetchBankCardsRequest()
 * 2. Saga 監聽到 action，先設置 loading 狀態
 * 3. Saga 調用 API 取得資料
 * 4. 取得資料後，Saga dispatch fetchBankCardsSuccess() 更新 store
 * 5. 如果有錯誤，Saga dispatch fetchBankCardsFailure() 設置錯誤訊息
 */
function* fetchBankCardsSaga(): SagaIterator {
  // 步驟 1: 開始取得銀行卡（設置 loading 狀態）
  yield put(fetchBankCardsStart());

  try {
    // 步驟 2: 使用 httpClient 呼叫取得銀行卡 API
    const bankCards = yield call(bankCardsApi.getBankCards);

    logger.info('取得銀行卡列表成功', {
      count: bankCards.length,
      cardIds: bankCards.map((c: any) => c.id),
      cards: bankCards.map((c: any) => ({
        id: c.id,
        name: c.name,
        cardNumber: c.cardNumber,
      })),
    });

    // 步驟 3: 取得資料後，更新 Redux State
    yield put(fetchBankCardsSuccess(bankCards));
  } catch (error: any) {
    const errorResult = handleSagaError(error, '取得銀行卡列表失敗');
    yield put(fetchBankCardsFailure(errorResult));
  }
}

/**
 * Watcher Saga
 * 監聽特定的 action 並觸發對應的 saga
 */
export function* watchBankCardsSagas(): SagaIterator {
  // takeLatest: 如果有多個請求，只處理最新的一個
  yield takeLatest(BANKCARDS_ACTIONS.FETCH_BANKCARDS_REQUEST, fetchBankCardsSaga);
}

