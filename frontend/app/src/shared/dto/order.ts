// API 資料傳輸物件（DTO）— 與後端 v1 端點對齊
import type { Order } from '../domain/order';
import type { CreateOrderInput } from '../validation/order';

// 建立訂單請求本體
export type CreateOrderRequest = CreateOrderInput;

// 後端統一回傳格式：{ code, message, data }
export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

export interface OrderListData {
  list: Order[];
}

export interface OkData {
  ok: boolean;
}

export type { Order, CreateOrderInput };
