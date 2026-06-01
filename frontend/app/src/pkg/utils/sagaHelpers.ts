import logger from '../logger';

export interface SagaErrorResult {
  code: string;
  message: string;
  statusCode?: number;
}

export function handleSagaError(
  error: unknown,
  context?: string,
  extra?: Record<string, any>
): SagaErrorResult {
  const prefix = context ? `[${context}] ` : '';

  if (error instanceof Error) {
    const axiosError = error as any;
    const statusCode = axiosError?.response?.status;
    const serverMessage = axiosError?.response?.data?.message;
    const code = axiosError?.response?.data?.code || axiosError?.code || 'UNKNOWN';

    const message = serverMessage || error.message || '未知錯誤';

    logger.error(`${prefix}${message}`, { code, statusCode, ...extra });

    return { code, message, statusCode };
  }

  logger.error(`${prefix}未知錯誤`, { error, ...extra });
  return { code: 'UNKNOWN', message: '發生未知錯誤' };
}

export interface ApiResponse<T = any> {
  success: boolean;
  message: string;
  data: T;
  code: string;
}
