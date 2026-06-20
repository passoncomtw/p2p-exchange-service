// v1 訂單 API（共用 shared 的 fetch client，僅 base URL 依平台設定）
import { createApiClient } from '@shared'

const client = createApiClient({
  baseUrl: import.meta.env.VITE_API_URL,
})

export const ordersApi = client.orders
