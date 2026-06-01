import { theme } from '@/theme';

/**
 * 訂單狀態映射
 * 0: 等待付款
 * 1: 已付款等待放行
 * 2: 買家取消
 * 3: 賣家取消
 * 4: 已放行
 */
export const ORDER_STATUS_MAP: Record<number, { label: string; color: string, statusString: string }> = {
  0: { label: '等待付款', color: theme.status.warning, statusString: 'pending_payment' },
  1: { label: '已付款等待放行', color: theme.status.info, statusString: 'pending_release' },
  2: { label: '買家取消', color: theme.text.tertiary, statusString: 'buyer_cancelled' },
  3: { label: '賣家取消', color: theme.text.tertiary, statusString: 'seller_cancelled' },
  4: { label: '已放行', color: theme.status.success, statusString: 'released' },
};

