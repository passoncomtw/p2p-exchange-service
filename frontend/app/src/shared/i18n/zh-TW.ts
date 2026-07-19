// v1 訂單相關繁體中文語系（Web 與 App 共用）
// v1 僅提供 zh-TW；架構保留未來擴充其他語系。
export const zhTW = {
  order: {
    // 導覽 / 頁面標題
    nav: {
      create: '掛單',
      myOrders: '我的掛單',
      admin: '訂單管理',
      trade: '交易',
      orders: '訂單',
      wallet: '錢包',
      market: '掛單',
      myListings: '我的掛單',
      matchedOrders: '媒合訂單',
      profile: '我',
    },
    pageTitle: {
      create: '建立掛單',
      myOrders: '我的掛單',
      admin: '後台訂單管理',
      detail: '訂單詳情',
      trade: '交易市場',
      listingDetail: '掛單詳情',
      orderDetail: '訂單詳情',
      addPayment: '新增收款帳戶',
    },

    // 類型
    type: {
      buy: '買幣',
      sell: '賣幣',
    },

    // 掛單表單(依類型切換語意)
    create: {
      sectionLabel: '我要',
      buyLabel: '買入 USDT',
      sellLabel: '賣出 USDT',
      buyHint: '您將以 TWD 購買 USDT，設定您願意支付的單價與數量',
      sellHint: '您將賣出 USDT 換取 TWD，設定您願意接受的單價與數量',
      priceLabel: '單價',
      priceBuyUnit: 'TWD / USDT',
      priceBuyPlaceholder: '每顆 USDT 您願意付多少 TWD',
      priceSellPlaceholder: '每顆 USDT 您願意收多少 TWD',
      quantityLabel: '數量',
      quantityUnit: 'USDT',
      quantityBuyPlaceholder: '想買入多少 USDT',
      quantitySellPlaceholder: '想賣出多少 USDT',
      totalBuy: '需支付',
      totalSell: '將收到',
      paymentMethodLabel: '收款帳戶',
      paymentMethodHint: '賣幣掛單必須選擇收款帳戶，買方付款後款項將轉入此帳戶',
      submitBuy: '建立買幣掛單',
      submitSell: '建立賣幣掛單',
    },

    // 掛單狀態
    status: {
      open: '待成交',
      completed: '已完成',
      cancelled: '已取消',
      all: '全部',
      active: '待成交',
      paused: '暫停',
    },

    // P2P 訂單狀態
    orderStatus: {
      matched: '交易中',
      paid: '已付款',
      releasing: '放行中',
      completed: '已完成',
      cancelled: '已取消',
      timeout: '已超時',
      disputed: '申訴中',
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
      remainingAmount: '可交易數量',
      fiatTotal: '法幣總額',
      tradeAmount: '交易數量',
      fiatEstimate: '預估法幣金額',
      orderNo: '訂單編號',
      role: '角色',
      counterparty: '對手方',
      paymentDeadline: '付款期限',
      bankName: '銀行名稱',
      accountName: '戶名',
      accountNumber: '帳號',
    },

    // 付款方式
    paymentMethod: {
      bank_transfer: '銀行轉帳',
      convenience_store: '超商代收',
    },

    // 登入(App)
    login: {
      brand: 'P2P 交易',
      subtitle: '登入以開始買賣 USDT',
      account: '帳號',
      password: '密碼',
      accountPlaceholder: '請輸入帳號',
      passwordPlaceholder: '請輸入密碼',
      submit: '登入',
      submitting: '登入中',
      demoHint: '買方帳號:testdemo001\n賣方帳號:testdemo002\n密碼:a12345678',
      logout: '登出',
      accountRequired: '請輸入帳號',
      passwordRequired: '請輸入密碼',
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
      retry: '重新載入',
      goCreate: '前往掛單',
      takeBuy: '買幣',
      takeSell: '賣幣',
      markPaid: '我已付款',
      confirmReceipt: '確認收款',
      cancelOrder: '取消訂單',
      dispute: '申訴',
      addPayment: '新增收款帳戶',
      max: '全部',
    },

    // 角色
    role: {
      asBuyer: '我是買方',
      asSeller: '我是賣方',
    },

    // 提示訊息
    message: {
      createSuccess: '掛單建立成功',
      cancelSuccess: '掛單已取消',
      completeSuccess: '訂單已標記完成',
      loadFailed: '載入失敗，請稍後再試',
      errorTitle: '載入失敗',
      errorBody: '無法取得掛單資料，請稍後再試。',
      submitFailed: '送出失敗，請稍後再試',
      emptyMine: '尚無掛單',
      emptyAdmin: '尚無訂單',
      cancelConfirm: '確定要取消此掛單嗎？',
      completeConfirm: '確定要將此訂單標記為已完成嗎？',
      takeOrderSuccess: '接單成功',
      takeOrderFailed: '接單失敗',
      markPaidSuccess: '已標記付款',
      markPaidConfirm: '確認已完成付款？',
      confirmReceiptSuccess: '已確認收款',
      confirmReceiptConfirm: '確認已收到款項？確認後將完成交易。',
      cancelOrderSuccess: '訂單已取消',
      cancelOrderConfirm: '確定要取消此訂單嗎？',
      disputeSuccess: '已提出申訴',
      disputeConfirm: '確定要對此訂單提出申訴嗎？',
      listingUnavailable: '此掛單已被接單或不可交易',
      noPaymentMethod: '您尚未新增收款帳戶，請先新增',
      emptyTrade: '目前沒有可交易的掛單',
      emptyOrders: '尚無訂單',
      addPaymentSuccess: '收款帳戶新增成功',
      hintMatchedBuyer: '請完成付款並點擊「我已付款」',
      hintMatchedSeller: '等待買方付款中',
      hintPaidBuyer: '已付款，等待賣方確認收款',
      hintPaidSeller: '買方已付款，請確認收款以完成交易',
      hintCompleted: '交易已完成',
      hintCancelled: '訂單已取消',
      hintDisputed: '申訴處理中，請等待客服介入',
      invalidAmount: '請輸入有效的交易數量（須大於 0 且不超過可交易數量）',
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
