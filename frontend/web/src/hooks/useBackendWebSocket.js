import { useEffect, useRef, useCallback } from 'react'
import { useSelector } from 'react-redux'

const WS_BASE_URL = (import.meta.env.VITE_API_URL || 'http://localhost:8888')
  .replace(/^http/, 'ws')

const RECONNECT_DELAY_MS = 3000
const MAX_RECONNECT_ATTEMPTS = 5

/**
 * useBackendWebSocket — 連接後台 WebSocket，接收 NATS P2P_NOTIFY 廣播事件。
 *
 * @param {(event: {type: string, data: object}) => void} onMessage 收到訊息的回呼
 * @returns {{ connected: boolean }}
 */
export function useBackendWebSocket(onMessage) {
  const token = useSelector((state) => state.auth.token)
  const wsRef = useRef(null)
  const attemptsRef = useRef(0)
  const unmountedRef = useRef(false)
  const onMessageRef = useRef(onMessage)
  onMessageRef.current = onMessage

  const connect = useCallback(() => {
    if (!token || unmountedRef.current) return
    if (attemptsRef.current >= MAX_RECONNECT_ATTEMPTS) return

    const url = `${WS_BASE_URL}/ws/backend?token=${token}`
    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onopen = () => {
      attemptsRef.current = 0
    }

    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data)
        if (msg?.type && onMessageRef.current) {
          onMessageRef.current(msg)
        }
      } catch {
        // ignore malformed frames
      }
    }

    ws.onclose = () => {
      if (unmountedRef.current) return
      attemptsRef.current += 1
      setTimeout(connect, RECONNECT_DELAY_MS)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [token])

  useEffect(() => {
    unmountedRef.current = false
    connect()
    return () => {
      unmountedRef.current = true
      wsRef.current?.close()
    }
  }, [connect])
}
