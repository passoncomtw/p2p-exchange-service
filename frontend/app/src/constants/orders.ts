import { theme } from '@/theme';

export const ORDER_STATUS_MAP: Record<string, { label: string; color: string }> = {
  matched:   { label: '待付款',       color: theme.status.warning },
  paid:      { label: '已付款待放行', color: theme.status.info },
  releasing: { label: '放行中',       color: theme.status.info },
  completed: { label: '已完成',       color: theme.status.success },
  cancelled: { label: '已取消',       color: theme.text.tertiary },
  timeout:   { label: '已超時',       color: theme.text.tertiary },
  disputed:  { label: '申訴中',       color: theme.status.error },
};
