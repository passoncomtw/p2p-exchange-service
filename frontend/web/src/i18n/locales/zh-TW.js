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
    },
    type: {
      buy: '買入',
      sell: '賣出',
    },
    status: {
      pending: '待付款',
      paid: '已付款',
      completed: '已完成',
      cancelled: '已取消',
      disputed: '申訴中',
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
