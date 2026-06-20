import { createTheme } from '@mui/material/styles'

export const tokens = {
  // Primary / Brand
  primary: '#FFC107',
  primaryDark: '#FFB300',
  primaryDisabled: '#FFE082',
  primaryDeep: '#FF8F00',

  // Sidebar (Dark Theme)
  sidebarBg: '#2C2C3C',
  sidebarHover: '#3A3A4E',
  sidebarActive: '#1A1A28',
  sidebarText: '#CCCCCC',
  sidebarTextActive: '#FFFFFF',

  // Page Backgrounds
  bgLogin: '#EBEDF2',
  bgContent: '#F5F5F7',
  bgCard: '#FFFFFF',

  // Neutral Text
  textPrimary: '#333333',
  textSecondary: '#666666',
  textTertiary: '#999999',
  textPlaceholder: '#BFBFBF',
  textAvatar: '#9E9E9E',

  // Borders & Dividers
  borderInput: '#D9D9D9',
  borderCard: '#EBEBEB',
  borderSelect: '#E0E0E0',
  borderSidebarLogo: 'rgba(255,255,255,0.06)',

  // Scrollbar
  scrollbarThumb: '#444444',

  // Semantic
  danger: '#F44336',
  statusActive: '#4CAF50',
  statusFrozen: '#FF9800',
  statusStopped: '#9E9E9E',

  // v1 訂單狀態色（Web 專用，不與 App 共用）
  statusOpen: '#FF9800', // 待成交
  statusCompleted: '#4CAF50', // 已完成
  statusCancelled: '#9E9E9E', // 已取消

  // v1 買 / 賣類型色（Web 專用）
  typeBuy: '#4CAF50',
  typeSell: '#F44336',
}

// 依訂單狀態取得對應色票（Web）
export const statusColor = (status) =>
  ({ open: tokens.statusOpen, completed: tokens.statusCompleted, cancelled: tokens.statusCancelled }[status])

// 依訂單類型取得對應色票（Web）
export const typeColor = (type) => (type === 'buy' ? tokens.typeBuy : tokens.typeSell)

const theme = createTheme({
  palette: {
    primary: {
      main: tokens.primary,
      dark: tokens.primaryDark,
      light: tokens.primaryDisabled,
      contrastText: '#FFFFFF',
    },
    error: {
      main: tokens.danger,
    },
    background: {
      default: tokens.bgContent,
      paper: tokens.bgCard,
    },
    text: {
      primary: tokens.textPrimary,
      secondary: tokens.textSecondary,
    },
  },
  typography: {
    fontFamily: 'system-ui, -apple-system, sans-serif',
  },
})

export default theme
