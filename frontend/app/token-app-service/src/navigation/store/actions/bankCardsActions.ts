/**
 * BankCards Actions
 * 定義銀行卡相關的 action types
 */

export const BANKCARDS_ACTIONS = {
  FETCH_BANKCARDS_REQUEST: 'FETCH_BANKCARDS_REQUEST',
} as const;

/**
 * Action Creators
 */
export const fetchBankCardsRequest = () => ({
  type: BANKCARDS_ACTIONS.FETCH_BANKCARDS_REQUEST,
});

