/**
 * Web notification provider stub。
 * 待串接 Firebase Cloud Messaging 或 WebSocket 時實作。
 */
export class WebNotificationProvider {
  async setup() {
    return null
  }

  onMessage(_cb) {
    return () => {}
  }

  onNotificationTap(_cb) {
    return () => {}
  }

  async getInitialNotification() {
    return null
  }
}
