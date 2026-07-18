// 訂單領域型別定義（v1）
// 「掛單」= 使用者張貼一筆買幣或賣幣的訂單。

// 訂單類型：買幣 / 賣幣
export type OrderType = 'buy' | 'sell';

// 訂單狀態（v1 最簡集合）
//   open      待成交：建立後的初始狀態
//   completed 已完成：後台於詳情標記完成
//   cancelled 已取消：使用者於我的掛單取消 open 訂單
export type OrderStatus = 'open' | 'completed' | 'cancelled';

// v1 支援的幣別與法幣（下拉結構保留以利日後擴充）
export type Asset = 'USDT';
export type Fiat = 'TWD';

// v1 付款方式：銀行轉帳 / 超商代碼
export type PaymentMethod = 'bank_transfer' | 'convenience_store';

export const ORDER_TYPES: readonly OrderType[] = ['buy', 'sell'];
export const ORDER_STATUSES: readonly OrderStatus[] = ['open', 'completed', 'cancelled'];
export const ASSETS: readonly Asset[] = ['USDT'];
export const FIATS: readonly Fiat[] = ['TWD'];
export const PAYMENT_METHODS: readonly PaymentMethod[] = ['bank_transfer', 'convenience_store'];

// 訂單資料模型（與後端 API 契約對齊）
export interface Order {
  id: string;
  type: OrderType;
  asset: Asset;
  fiat: Fiat;
  price: number; // 單價，法幣 / 每單位幣
  quantity: number; // 數量
  totalAmount: number; // = price * quantity
  paymentMethod: PaymentMethod;
  status: OrderStatus;
  createdBy: string; // 目前固定 demo_user；seed 資料為其他使用者
  createdAt: string; // ISO 字串
  updatedAt: string; // ISO 字串
}

// 依單價與數量計算總額（前後端共用，避免重複實作）
export function calcTotalAmount(price: number, quantity: number): number {
  if (!Number.isFinite(price) || !Number.isFinite(quantity)) return 0;
  return Math.round(price * quantity * 100) / 100;
}
