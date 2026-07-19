export default {
  // 登入頁
  login: {
    brand: '豐盈錢包',
    subtitle: '後台管理系統',
    usernameLabel: '帳號',
    usernamePlaceholder: '請輸入帳號',
    passwordLabel: '登入密碼',
    passwordPlaceholder: '請輸入登入密碼',
    submit: '登入',
    version: 'V 1.0.0',
  },

  // 側邊欄
  sidebar: {
    brand: '豐盈錢包',
    tagline: '運營後台',
    menu: {
      members: '會員管理',
      memberList: '會員列表',
      orders: '訂單管理',
      finance: '財務管理',
      fiatWithdrawals: '法幣提領審核',
    },
  },

  // 頂部導航
  header: {
    logout: '退出登入',
  },

  // 頁面標題
  page: {
    dashboard: '首頁',
    memberList: '會員列表',
    orders: '訂單管理',
    fiatWithdrawals: '法幣提領審核',
  },

  // 會員列表
  member: {
    keyword: '關鍵字',
    keywordPlaceholder: '請輸入帳號或信箱',
    reset: '重置',
    filter: '篩選',
    loading: '載入中...',
    empty: '暫無資料',
    column: {
      username: '帳號',
      email: '電子郵件',
      createdAt: '建立時間',
      updatedAt: '更新時間',
    },
    pagination: {
      total: '共 {{total}} 條',
      page: '第 {{current}}/{{total}} 頁',
      pageSize: '每頁筆數：',
      items: '筆',
    },
  },

  // 訂單列表
  order: {
    keyword: '關鍵字',
    keywordPlaceholder: '請輸入訂單編號',
    statusLabel: '狀態',
    statusAll: '全部',
    reset: '重置',
    filter: '篩選',
    loading: '載入中...',
    empty: '暫無資料',
    column: {
      orderNo: '訂單編號',
      type: '類型',
      crypto: '數位貨幣',
      fiatAmount: '法幣金額',
      price: '單價',
      status: '狀態',
      createdAt: '建立時間',
      action: '操作',
    },
    type: {
      buy: '買入',
      sell: '賣出',
    },
    status: {
      matched: '待付款',
      pending: '待付款',
      paid: '已付款',
      completed: '已完成',
      cancelled: '已取消',
      timeout: '已逾時',
      disputed: '申訴中',
    },
    pagination: {
      total: '共 {{total}} 條',
      page: '第 {{current}}/{{total}} 頁',
      pageSize: '每頁筆數：',
      items: '筆',
    },
    detail: {
      title: '訂單詳情',
      basicInfo: '基本資訊',
      tradeInfo: '交易資訊',
      paymentDeadline: '付款截止',
      cancelledAt: '取消時間',
      cancelReason: '取消原因',
      fee: '手續費',
      view: '查看',
      close: '關閉',
    },
    dispute: {
      title: '申訴仲裁',
      hint: '此訂單處於申訴狀態。請選擇仲裁方式：「完成交易」將資產轉給買家；「退款取消」將資產退還給賣家。',
      arbitrate: '仲裁',
      selectAction: '請選擇仲裁方式',
      actionComplete: '完成交易（釋放給買家）',
      actionRefund: '退款取消（退還給賣家）',
      completeHint: '確認後，凍結的數位資產將轉入買家錢包，訂單標記為已完成。',
      reasonLabel: '取消原因',
      reasonPlaceholder: '請說明退款原因（例如：買家未完成付款）',
      reasonRequired: '退款取消時必須填寫原因',
      resolveFailed: '仲裁操作失敗，請稍後再試',
      submit: '確認仲裁',
    },
  },

  // 法幣提領審核
  fiatWithdrawal: {
    statusLabel: '狀態篩選',
    statusAll: '全部',
    reset: '重置',
    filter: '篩選',
    empty: '暫無資料',
    status: {
      pending: '待審核',
      approved: '已核准',
      rejected: '已拒絕',
    },
    column: {
      id: 'ID',
      userId: '用戶 ID',
      amount: '金額',
      bankCode: '銀行代碼',
      bankAccount: '銀行帳號',
      accountName: '戶名',
      status: '狀態',
      createdAt: '申請時間',
      action: '操作',
      rejectReason: '拒絕原因',
    },
    review: {
      title: '審核提領申請',
      applicantInfo: '申請人資訊',
      withdrawalInfo: '提領資訊',
      action: '審核操作',
      approve: '核准',
      reject: '拒絕',
      approveHint: '核准後，系統將安排轉帳至申請人銀行帳戶。',
      reasonLabel: '拒絕原因',
      reasonPlaceholder: '請填寫拒絕原因（必填）',
      reasonRequired: '拒絕時必須填寫原因',
      failed: '審核操作失敗，請稍後再試',
      submit: '確認提交',
      close: '關閉',
      view: '查看',
      actionChip: '審核',
      selectAction: '請選擇審核結果',
    },
    pagination: {
      total: '共 {{total}} 條',
      page: '第 {{current}}/{{total}} 頁',
      pageSize: '每頁筆數：',
      items: '筆',
    },
  },

  // 通用
  common: {
    welcome: '歡迎使用豐盈錢包運營後台',
  },
}
