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
}

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
