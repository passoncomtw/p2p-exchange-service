import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { SagaErrorResult } from '@pkg/utils/sagaHelpers';

/**
 * 銀行資訊
 */
export interface Bank {
  id: number;
  bankCode: string;
  bankName: string;
}

/**
 * 銀行卡資訊
 */
export interface BankCard {
  id: number;
  bankId: number;
  name: string;
  cardNumber: string;
  branchName: string; // 必需，銀行卡必定有分行名稱
  status: number;
  createdAt: string;
  bank: Bank;
}

/**
 * BankCards State
 */
interface BankCardsState {
  cards: BankCard[];
  loading: boolean;
  error: string | null;
}

const initialState: BankCardsState = {
  cards: [],
  loading: false,
  error: null,
};

/**
 * BankCards Slice
 * 負責管理銀行卡相關的狀態
 */
const bankCardsSlice = createSlice({
  name: 'bankCards',
  initialState,
  reducers: {
    fetchBankCardsStart(state) {
      state.loading = true;
      state.error = null;
    },
    fetchBankCardsSuccess(state, action: PayloadAction<BankCard[]>) {
      state.loading = false;
      state.cards = action.payload;
      state.error = null;
    },
    fetchBankCardsFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.loading = false;
      state.error = action.payload.message;
    },
    clearBankCardsError(state) {
      state.error = null;
    },
    resetBankCards() {
      // 重置為初始狀態（用於登出）
      return initialState;
    },
  },
});

export const {
  fetchBankCardsStart,
  fetchBankCardsSuccess,
  fetchBankCardsFailure,
  clearBankCardsError,
  resetBankCards,
} = bankCardsSlice.actions;

export default bankCardsSlice.reducer;

