// Browser Notification API — Web 實作
// 目前為 stub，待 Web 通知功能開發時補齊

export function requestPermission() {
  if (!('Notification' in window)) return Promise.resolve('denied')
  return Notification.requestPermission()
}

export function showNotification(title, options = {}) {
  if (Notification.permission !== 'granted') return
  return new Notification(title, options)
}
