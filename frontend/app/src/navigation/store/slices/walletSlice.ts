import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import type { SagaErrorResult } from '@pkg/utils/sagaHelpers';
import type { WalletItem, WalletLedgerItem } from '@/interfaces/wallet';

interface WalletState {
  wallets: WalletItem[];
  walletsLoading: boolean;
  walletsError: string | null;
  ledgers: WalletLedgerItem[];
  ledgersTotal: number;
  ledgersLoading: boolean;
  ledgersError: string | null;
  selectedCurrency: string;
}

const initialState: WalletState = {
  wallets: [],
  walletsLoading: false,
  walletsError: null,
  ledgers: [],
  ledgersTotal: 0,
  ledgersLoading: false,
  ledgersError: null,
  selectedCurrency: 'USDT',
};

const walletSlice = createSlice({
  name: 'wallet',
  initialState,
  reducers: {
    fetchWalletsStart(state) {
      state.walletsLoading = true;
      state.walletsError = null;
    },
    fetchWalletsSuccess(state, action: PayloadAction<WalletItem[]>) {
      state.walletsLoading = false;
      state.wallets = action.payload;
    },
    fetchWalletsFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.walletsLoading = false;
      state.walletsError = action.payload.message;
    },
    fetchLedgersStart(state) {
      state.ledgersLoading = true;
      state.ledgersError = null;
    },
    fetchLedgersSuccess(
      state,
      action: PayloadAction<{ list: WalletLedgerItem[]; total: number }>
    ) {
      state.ledgersLoading = false;
      state.ledgers = action.payload.list;
      state.ledgersTotal = action.payload.total;
    },
    fetchLedgersFailure(state, action: PayloadAction<SagaErrorResult>) {
      state.ledgersLoading = false;
      state.ledgersError = action.payload.message;
    },
    setSelectedCurrency(state, action: PayloadAction<string>) {
      state.selectedCurrency = action.payload;
    },
    resetWallet() {
      return initialState;
    },
  },
});

export const {
  fetchWalletsStart,
  fetchWalletsSuccess,
  fetchWalletsFailure,
  fetchLedgersStart,
  fetchLedgersSuccess,
  fetchLedgersFailure,
  setSelectedCurrency,
  resetWallet,
} = walletSlice.actions;

export default walletSlice.reducer;
