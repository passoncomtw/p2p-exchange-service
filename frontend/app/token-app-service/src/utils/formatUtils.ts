import { Clipboard, Alert } from 'react-native';

/**
 * 格式化時間（MM:SS）
 * @param seconds 秒數
 * @returns 格式化後的時間字串，格式為 MM:SS
 */
export const formatTime = (seconds: number): string => {
  const mins = Math.floor(seconds / 60);
  const secs = seconds % 60;
  return `${String(mins).padStart(2, '0')}:${String(secs).padStart(2, '0')}`;
};

/**
 * 格式化日期時間
 * @param dateString 日期字串
 * @returns 格式化後的日期時間字串，格式為 YYYY-MM-DD HH:mm:ss
 */
export const formatDateTime = (dateString: string): string => {
  const date = new Date(dateString);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = String(date.getSeconds()).padStart(2, '0');
  return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
};

/**
 * 複製到剪貼板
 * @param text 要複製的文字
 * @param label 標籤名稱（用於提示訊息）
 */
export const copyToClipboard = (text: string, label: string): void => {
  Clipboard.setString(text);
  Alert.alert('已複製', `${label}已複製到剪貼板`);
};

