declare module '@pkg/utils/sagaHelpers' {
  import { Generator } from 'redux-saga/effects';

  export interface ApiResponse<T = any> {
    success: boolean;
    message: string;
    data: T;
    code: string;
  }

  export interface SagaErrorResult {
    code: string;
    message: string;
    statusCode?: number;
  }

  export function handleSagaError(error: unknown, context?: string, extra?: Record<string, any>): SagaErrorResult;

  interface ErrorPayload {
    code: string;
    message: string;
  }

  export interface FetchAPIOptions<TPayload = any, TResponse = any> {
    apiResult: (config: {
      customHeaders: Record<string, string>;
      payload: TPayload;
    }) => Promise<ApiResponse<TResponse>>;
    headers?: Record<string, string>;
    payload: TPayload;
    action: string;
    message?: string;
    resultHandler?: ((data: TResponse) => any) | null;
    onSuccess?: ((data: TResponse) => void) | ((data: TResponse) => Generator<any, void, any>);
    onError?: ((error: ErrorPayload) => void) | ((error: ErrorPayload) => Generator<any, void, any>);
    tokenSelector?: (state: any) => string | null;
  }

  export default function fetchAPIResult<TPayload = any, TResponse = any>(
    options: FetchAPIOptions<TPayload, TResponse>
  ): Generator<any, void, any>;

  export function apiRequest<T = any>(
    apiBaseUrl: string,
    endpoint: string,
    method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH',
    body?: any,
    customHeaders?: Record<string, string>
  ): Promise<ApiResponse<T>>;
}
