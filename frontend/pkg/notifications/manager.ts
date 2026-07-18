import type { INotificationProvider, NotificationMessage } from './types'

export class NotificationManager {
  constructor(private readonly provider: INotificationProvider) {}

  setup(): Promise<string | null> {
    return this.provider.setup()
  }

  onMessage(cb: (message: NotificationMessage) => void): () => void {
    return this.provider.onMessage(cb)
  }

  onNotificationTap(cb: (message: NotificationMessage) => void): () => void {
    return this.provider.onNotificationTap(cb)
  }

  getInitialNotification(): Promise<NotificationMessage | null> {
    return this.provider.getInitialNotification()
  }
}
