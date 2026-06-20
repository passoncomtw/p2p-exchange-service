// v1 訂單相關繁體中文語系（Web 與 App 共用）
// v1 僅提供 zh-TW；架構保留未來擴充其他語系。
export const zhTW = {
  order: {
    // 導覽 / 頁面標題
    nav: {
      create: '掛單',
      myOrders: '我的掛單',
      admin: '訂單管理',
    },
    pageTitle: {
      create: '建立掛單',
      myOrders: '我的掛單',
      admin: '後台訂單管理',
      detail: '訂單詳情',
    },

    // 類型
    type: {
      buy: '買幣',
      sell: '賣幣',
    },

    // 狀態
    status: {
      open: '待成交',
      completed: '已完成',
      cancelled: '已取消',
      all: '全部',
    },

    // 表單欄位
    field: {
      type: '類型',
      asset: '幣種',
      fiat: '法幣',
      price: '單價',
      quantity: '數量',
      totalAmount: '總額',
      paymentMethod: '付款方式',
      status: '狀態',
      createdBy: '建立者',
      createdAt: '建立時間',
      orderId: '訂單編號',
    },

    // 付款方式
    paymentMethod: {
      bank_transfer: '銀行轉帳',
      convenience_store: '超商代碼',
    },

    // 動作
    action: {
      submit: '送出',
      cancel: '取消掛單',
      complete: '標記完成',
      back: '返回',
      viewDetail: '查看詳情',
      confirm: '確認',
      dismiss: '關閉',
    },

    // 提示訊息
    message: {
      createSuccess: '掛單建立成功',
      cancelSuccess: '掛單已取消',
      completeSuccess: '訂單已標記完成',
      loadFailed: '載入失敗，請稍後再試',
      submitFailed: '送出失敗，請稍後再試',
      emptyMine: '尚無掛單',
      emptyAdmin: '尚無訂單',
      cancelConfirm: '確定要取消此掛單嗎？',
      completeConfirm: '確定要將此訂單標記為已完成嗎？',
    },

    // 驗證錯誤（對應 validation/order.ts 的 messageKey）
    error: {
      typeRequired: '請選擇類型',
      assetRequired: '請選擇幣種',
      fiatRequired: '請選擇法幣',
      pricePositive: '單價必須大於 0',
      quantityPositive: '數量必須大於 0',
      paymentMethodRequired: '請選擇付款方式',
    },

    // 篩選
    filter: {
      status: '狀態篩選',
    },
  },
} as const;

export default zhTW;
