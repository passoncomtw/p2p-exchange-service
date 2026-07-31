import { useEffect, useRef } from 'react';
import { useAppDispatch, useAppSelector } from '../navigation/store/hooks';
import { fetchOrderListRequest } from '../navigation/store/actions/ordersActions';

const WS_BASE_URL = (process.env.EXPO_PUBLIC_API_BASE_URL || 'https://p2p-exchange-api.passon.tw')
  .replace(/^https/, 'wss')
  .replace(/^http/, 'ws');

const RECONNECT_DELAY_MS = 3000;
const MAX_RECONNECT_ATTEMPTS = 10;

export function useAppWebSocket(): void {
  const dispatch = useAppDispatch();
  const token = useAppSelector((state) => state.auth.accessToken);
  const isAuthenticated = useAppSelector((state) => state.auth.isAuthenticated);

  const wsRef = useRef<WebSocket | null>(null);
  const attemptsRef = useRef(0);
  const unmountedRef = useRef(false);
  const tokenRef = useRef(token);
  tokenRef.current = token;

  useEffect(() => {
    if (!isAuthenticated || !token) return;

    unmountedRef.current = false;
    attemptsRef.current = 0;

    function connect() {
      if (unmountedRef.current) return;
      if (attemptsRef.current >= MAX_RECONNECT_ATTEMPTS) return;

      const ws = new WebSocket(`${WS_BASE_URL}/ws/app?token=${tokenRef.current}`);
      wsRef.current = ws;

      ws.onopen = () => {
        attemptsRef.current = 0;
      };

      ws.onmessage = (ev) => {
        try {
          const msg = JSON.parse(ev.data as string) as { type: string };
          if (msg?.type === 'order.status.changed') {
            dispatch(fetchOrderListRequest({ limit: 100 }));
          }
        } catch {
          // ignore malformed frames
        }
      };

      ws.onclose = () => {
        if (unmountedRef.current) return;
        attemptsRef.current += 1;
        setTimeout(connect, RECONNECT_DELAY_MS);
      };

      ws.onerror = () => {
        ws.close();
      };
    }

    connect();

    return () => {
      unmountedRef.current = true;
      wsRef.current?.close();
    };
  }, [isAuthenticated, token, dispatch]);
}
