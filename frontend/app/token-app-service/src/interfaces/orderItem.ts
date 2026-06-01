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
export interface Bankcard {
  id: number;
  bankId: number;
  branchName: string;
  cardNumber: string;
  name: string;
  status: number;
  createdAt: string;
  bank: Bank;
}

/**
 * 用戶資訊（簡化版）
 */
export interface User {
  id: number;
  name: string;
}

/**
 * 掛單資訊
 */
export interface PendingOrder {
  id: string;
  amount: number;
  balance: number;
  bankcard: Bankcard;
  cancelAmount: number;
  cancelCount: number;
  createdAt: string;
  doneAmount: number;
  doneCount: number;
  isSplit: boolean;
  minAmount: number;
  processAmount: number;
  processCount: number;
  status: number;
  transactionMinutes: number;
  type: number; // 0: 買幣, 1: 賣幣
  user: User;
}

/**
 * 訂單項目（訂單列表 Response 的每個 item）
 */
export interface OrderItem {
  id: string;
  amount: number;
  bankcard: Bankcard;
  cancelReason?: string;
  createdAt: string;
  finishAt?: string;
  pendingOrder: PendingOrder;
  status: number; // 訂單狀態：0=待付款, 1=待放行, 2=已完成, 3=已取消, 4=申訴中
  user: User;
}

