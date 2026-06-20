// v1 訂單 API（App 端）：共用 shared 的 fetch client，僅 base URL 依平台設定。
// Android 模擬器存取本機後端請改用 10.0.2.2；iOS 模擬器可用 localhost。
import { createApiClient } from '@shared';

const baseUrl = process.env.EXPO_PUBLIC_API_BASE_URL || 'http://localhost:8888';

const client = createApiClient({ baseUrl });

export const ordersApi = client.orders;
