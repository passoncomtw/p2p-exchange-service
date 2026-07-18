// 訂單狀態機（v1）
// 轉換規則：
//   open -> cancelled（使用者端）
//   open -> completed（Web 後台）
//   其餘轉換不開放。
import type { OrderStatus } from './order';

// 允許的狀態轉換表
const ALLOWED_TRANSITIONS: Record<OrderStatus, readonly OrderStatus[]> = {
  open: ['cancelled', 'completed'],
  completed: [],
  cancelled: [],
};

export function canTransition(from: OrderStatus, to: OrderStatus): boolean {
  return ALLOWED_TRANSITIONS[from].includes(to);
}

// 使用者端是否可取消（僅 open 可取消）
export function canCancel(status: OrderStatus): boolean {
  return canTransition(status, 'cancelled');
}

// 後台是否可標記完成（僅 open 可標記完成）
export function canComplete(status: OrderStatus): boolean {
  return canTransition(status, 'completed');
}
