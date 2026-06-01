import { takeEvery, put } from 'redux-saga/effects';
import { PayloadAction } from '@reduxjs/toolkit';
import { SagaIterator } from 'redux-saga';
import logger from '@pkg/logger';
import { SagaErrorResult } from '@pkg/utils/sagaHelpers';

/**
 * Root Error Saga - 統一處理所有錯誤
 * 
 * 自動監聽所有以 Failure 結尾的 action，檢查是否為 401 錯誤
 * 如果是 401 錯誤，自動觸發登出流程
 * 
 * 優點：
 * 1. 集中式錯誤處理，不需要在每個 saga 中重複寫 401 檢查
 * 2. 易於維護和擴展（未來可以添加其他錯誤處理邏輯）
 * 3. 符合 Redux Saga 的設計哲學
 */

/**
 * 處理錯誤 action
 */
function* handleErrorAction(action: PayloadAction<SagaErrorResult>): SagaIterator {
  const { payload } = action;
  
  // 檢查 payload 是否包含 statusCode
  if (payload && typeof payload === 'object' && 'statusCode' in payload) {
    const { statusCode } = payload;
    
    // 401 Unauthorized - Token 過期或無效
    if (statusCode === 401) {
      logger.warn('檢測到 401 錯誤，自動登出', {
        action: action.type,
      });
      
      // 觸發登出流程
      yield put({ type: 'auth/logoutRequest' });
    }
    
    // 未來可以在這裡添加其他錯誤處理邏輯
    // 例如：403 無權限、500 伺服器錯誤重試等
  }
}

/**
 * 監聽所有 Failure action
 */
export function* watchErrorSaga(): SagaIterator {
  yield takeEvery(
    ((action: PayloadAction<any>) => action.type.endsWith('Failure')) as any,
    handleErrorAction
  );
}

