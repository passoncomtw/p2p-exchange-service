// 共用 API client（fetch 實作，Web 與 App 共用，僅 base URL 依平台設定）
// 不以 localStorage 作為主要資料源；所有資料透過後端 REST 存取。
import type { Order } from '../domain/order';
import type { OrderStatus } from '../domain/order';
import type {
  ApiEnvelope,
  CreateOrderRequest,
  OrderListData,
  OkData,
} from '../dto/order';

export interface ApiClientOptions {
  baseUrl: string;
  // 可選的 fetch 實作（測試或特殊環境注入）；預設使用全域 fetch。
  fetchImpl?: typeof fetch;
}

export class ApiError extends Error {
  code: number;
  constructor(code: number, message: string) {
    super(message);
    this.name = 'ApiError';
    this.code = code;
  }
}

export interface OrdersApi {
  create(input: CreateOrderRequest): Promise<Order>;
  listMine(): Promise<Order[]>;
  cancel(id: string): Promise<void>;
  adminList(status?: OrderStatus): Promise<Order[]>;
  adminGet(id: string): Promise<Order>;
  adminComplete(id: string): Promise<void>;
}

export interface ApiClient {
  orders: OrdersApi;
}

export function createApiClient(options: ApiClientOptions): ApiClient {
  const baseUrl = options.baseUrl.replace(/\/$/, '');
  const doFetch = options.fetchImpl ?? fetch;

  async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await doFetch(`${baseUrl}${path}`, {
      ...init,
      headers: {
        'Content-Type': 'application/json',
        ...(init?.headers ?? {}),
      },
    });

    let body: ApiEnvelope<T> | null = null;
    try {
      body = (await res.json()) as ApiEnvelope<T>;
    } catch {
      throw new ApiError(res.status, `HTTP ${res.status}`);
    }

    if (!res.ok || (body && body.code !== 0)) {
      throw new ApiError(body?.code ?? res.status, body?.message ?? `HTTP ${res.status}`);
    }
    return body.data;
  }

  const orders: OrdersApi = {
    create: (input) =>
      request<Order>('/v1/orders', {
        method: 'POST',
        body: JSON.stringify(input),
      }),

    listMine: () =>
      request<OrderListData>('/v1/orders/mine').then((d) => d.list),

    cancel: (id) =>
      request<OkData>(`/v1/orders/${encodeURIComponent(id)}/cancel`, {
        method: 'POST',
      }).then(() => undefined),

    adminList: (status) => {
      const query = status ? `?status=${encodeURIComponent(status)}` : '';
      return request<OrderListData>(`/v1/admin/orders${query}`).then((d) => d.list);
    },

    adminGet: (id) =>
      request<Order>(`/v1/admin/orders/${encodeURIComponent(id)}`),

    adminComplete: (id) =>
      request<OkData>(`/v1/admin/orders/${encodeURIComponent(id)}/complete`, {
        method: 'POST',
      }).then(() => undefined),
  };

  return { orders };
}
