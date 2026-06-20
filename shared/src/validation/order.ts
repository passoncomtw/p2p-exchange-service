// 建立訂單的輸入驗證（前後端共用）
// 回傳以 i18n 訊息鍵表示的錯誤，UI 端再行翻譯。
import {
  ORDER_TYPES,
  ASSETS,
  FIATS,
  PAYMENT_METHODS,
  type OrderType,
  type Asset,
  type Fiat,
  type PaymentMethod,
} from '../domain/order';

export interface CreateOrderInput {
  type: OrderType;
  asset: Asset;
  fiat: Fiat;
  price: number;
  quantity: number;
  paymentMethod: PaymentMethod;
}

export interface FieldError {
  field: keyof CreateOrderInput;
  messageKey: string;
}

// 驗證建立訂單輸入，回傳錯誤陣列（空陣列代表通過）
export function validateCreateOrder(input: Partial<CreateOrderInput>): FieldError[] {
  const errors: FieldError[] = [];

  if (!input.type || !ORDER_TYPES.includes(input.type)) {
    errors.push({ field: 'type', messageKey: 'order.error.typeRequired' });
  }
  if (!input.asset || !ASSETS.includes(input.asset)) {
    errors.push({ field: 'asset', messageKey: 'order.error.assetRequired' });
  }
  if (!input.fiat || !FIATS.includes(input.fiat)) {
    errors.push({ field: 'fiat', messageKey: 'order.error.fiatRequired' });
  }
  if (!isPositiveNumber(input.price)) {
    errors.push({ field: 'price', messageKey: 'order.error.pricePositive' });
  }
  if (!isPositiveNumber(input.quantity)) {
    errors.push({ field: 'quantity', messageKey: 'order.error.quantityPositive' });
  }
  if (!input.paymentMethod || !PAYMENT_METHODS.includes(input.paymentMethod)) {
    errors.push({ field: 'paymentMethod', messageKey: 'order.error.paymentMethodRequired' });
  }

  return errors;
}

function isPositiveNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0;
}
