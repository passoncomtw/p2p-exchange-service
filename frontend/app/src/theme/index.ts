/**
 * 應用主題配置
 * 極簡風格 - 五個基礎顏色系統
 */

export const theme = {
  /**
   * 1. Primary - 主色（黃色）
   * 用於：主要按鈕、頂部導航欄、重要操作
   */
  primary: '#FFC107',

  /**
   * 2. Secondary - 次要色（黑色）
   * 用於：次要按鈕、標題文字、重要資訊
   */
  secondary: '#000000',

  /**
   * 3. Background - 背景色系
   */
  background: {
    primary: '#FFFFFF',    // 主背景（白色）
    secondary: '#F5F5F5',  // 次要背景（淺灰）
    tertiary: '#FAFAFA',   // 第三背景（極淺灰）
  },

  /**
   * 4. Text - 文字色系
   */
  text: {
    primary: '#000000',    // 主要文字（黑色）
    secondary: '#666666',  // 次要文字（深灰）
    tertiary: '#999999',   // 第三文字（灰色）
    placeholder: '#CCCCCC', // 佔位符（淺灰）
  },

  /**
   * 5. Border - 邊框色系
   */
  border: {
    light: '#F0F0F0',      // 淺邊框
    default: '#E5E5E5',    // 預設邊框
    dark: '#CCCCCC',       // 深邊框
  },

  /**
   * 狀態顏色（額外但必要）
   */
  status: {
    success: '#4CAF50',    // 成功（綠色）
    error: '#F44336',      // 錯誤（紅色）
    warning: '#FF9800',    // 警告（橙色）
    info: '#2196F3',       // 資訊（藍色）
  },

  /**
   * 常用間距
   */
  spacing: {
    xs: 4,
    sm: 8,
    md: 12,
    lg: 16,
    xl: 24,
    xxl: 32,
  },

  /**
   * 字體大小
   */
  fontSize: {
    xs: 11,
    sm: 12,
    md: 14,
    lg: 16,
    xl: 18,
    xxl: 20,
    xxxl: 24,
  },

  /**
   * 圓角
   */
  radius: {
    sm: 4,
    md: 6,
    lg: 8,
    xl: 12,
    xxl: 24,
    round: 999,
  },

  /**
   * 陰影
   */
  shadow: {
    light: {
      shadowColor: '#000',
      shadowOffset: { width: 0, height: 1 },
      shadowOpacity: 0.05,
      shadowRadius: 2,
      elevation: 1,
    },
    default: {
      shadowColor: '#000',
      shadowOffset: { width: 0, height: 2 },
      shadowOpacity: 0.05,
      shadowRadius: 4,
      elevation: 2,
    },
    heavy: {
      shadowColor: '#000',
      shadowOffset: { width: 0, height: 4 },
      shadowOpacity: 0.1,
      shadowRadius: 8,
      elevation: 4,
    },
  },
} as const;

/**
 * 通用樣式
 */
export const commonStyles = {
  // 容器
  container: {
    flex: 1,
    backgroundColor: theme.background.secondary,
  },

  // 頂部導航欄（黃色）
  header: {
    backgroundColor: theme.primary,
    paddingHorizontal: theme.spacing.lg,
    paddingTop: 60,
    paddingBottom: theme.spacing.lg,
  },

  headerTitle: {
    fontSize: theme.fontSize.xl,
    fontWeight: '600' as const,
    color: theme.secondary,
  },

  // 卡片
  card: {
    backgroundColor: theme.background.primary,
    borderRadius: theme.radius.xl,
    padding: theme.spacing.lg,
    ...theme.shadow.default,
  },

  // 輸入框
  input: {
    borderWidth: 1,
    borderColor: theme.border.default,
    borderRadius: theme.radius.lg,
    paddingHorizontal: theme.spacing.md,
    height: 48,
    fontSize: theme.fontSize.lg,
    color: theme.text.primary,
    backgroundColor: theme.background.primary,
  },

  // 主要按鈕（黃色）
  buttonPrimary: {
    backgroundColor: theme.primary,
    height: 48,
    borderRadius: theme.radius.xxl,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
  },

  buttonPrimaryText: {
    fontSize: theme.fontSize.lg,
    fontWeight: '600' as const,
    color: theme.background.primary,
  },

  // 次要按鈕（黑色）
  buttonSecondary: {
    backgroundColor: theme.secondary,
    height: 48,
    borderRadius: theme.radius.xxl,
    alignItems: 'center' as const,
    justifyContent: 'center' as const,
  },

  buttonSecondaryText: {
    fontSize: theme.fontSize.lg,
    fontWeight: '600' as const,
    color: theme.background.primary,
  },

  // 區塊
  section: {
    backgroundColor: theme.background.primary,
    padding: theme.spacing.lg,
    marginBottom: 1,
  },

  // 標籤
  label: {
    fontSize: theme.fontSize.md,
    color: theme.text.primary,
    marginBottom: theme.spacing.md,
  },

  // 提示文字
  hint: {
    fontSize: theme.fontSize.sm,
    color: theme.text.tertiary,
    marginTop: theme.spacing.sm,
  },

  // 錯誤文字
  error: {
    fontSize: theme.fontSize.sm,
    color: theme.status.error,
    marginTop: theme.spacing.sm,
  },
};

// ── v1 訂單畫面色票（App 專用，不與 Web 共用） ──────────────────────────────
// 值與 Web 維持一致以保留視覺,但由各平台各自擁有。
export const colors = {
  primary: '#FFC107',
  primaryDisabled: '#FFE082',
  primaryDeep: '#FF8F00',

  bgContent: '#F5F5F7',
  bgCard: '#FFFFFF',

  textPrimary: '#333333',
  textSecondary: '#666666',
  textTertiary: '#999999',
  textPlaceholder: '#BFBFBF',

  borderInput: '#D9D9D9',
  borderCard: '#EBEBEB',

  danger: '#F44336',

  statusOpen: '#FF9800',
  statusCompleted: '#4CAF50',
  statusCancelled: '#9E9E9E',

  typeBuy: '#4CAF50',
  typeSell: '#F44336',
} as const;

// 依訂單狀態取得對應色票（App）
export function statusColor(status: 'open' | 'completed' | 'cancelled'): string {
  switch (status) {
    case 'open':
      return colors.statusOpen;
    case 'completed':
      return colors.statusCompleted;
    case 'cancelled':
      return colors.statusCancelled;
  }
}

// 依訂單類型取得對應色票（App）
export function typeColor(type: 'buy' | 'sell'): string {
  return type === 'buy' ? colors.typeBuy : colors.typeSell;
}

export default theme;

