export default {
  // 登录页
  login: {
    brand: '丰盈钱包',
    subtitle: '后台管理系统',
    usernameLabel: '账号',
    usernamePlaceholder: '请输入账号',
    passwordLabel: '登录密码',
    passwordPlaceholder: '请输入登录密码',
    submit: '登录',
    version: 'V 1.0.0',
  },

  // 侧边栏
  sidebar: {
    brand: '丰盈钱包',
    tagline: '运营后台',
    menu: {
      members: '会员管理',
      memberList: '会员列表',
      orders: '订单管理',
    },
  },

  // 顶部导航
  header: {
    logout: '退出登录',
  },

  // 页面标题
  page: {
    dashboard: '首页',
    memberList: '会员列表',
    orders: '订单管理',
  },

  // 会员列表
  member: {
    keyword: '关键字',
    keywordPlaceholder: '请输入账号或邮箱',
    reset: '重置',
    filter: '筛选',
    loading: '加载中...',
    empty: '暂无数据',
    column: {
      username: '账号',
      email: '电子邮件',
      createdAt: '创建时间',
      updatedAt: '更新时间',
    },
    pagination: {
      total: '共 {{total}} 条',
      page: '第 {{current}}/{{total}} 页',
      pageSize: '每页笔数：',
      items: '笔',
    },
  },

  // 订单列表
  order: {
    keyword: '关键字',
    keywordPlaceholder: '请输入订单编号',
    statusLabel: '状态',
    statusAll: '全部',
    reset: '重置',
    filter: '筛选',
    loading: '加载中...',
    empty: '暂无数据',
    column: {
      orderNo: '订单编号',
      type: '类型',
      crypto: '数字货币',
      fiatAmount: '法币金额',
      price: '单价',
      status: '状态',
      createdAt: '创建时间',
      action: '操作',
    },
    type: {
      buy: '买入',
      sell: '卖出',
    },
    status: {
      matched: '待付款',
      pending: '待付款',
      paid: '已付款',
      completed: '已完成',
      cancelled: '已取消',
      timeout: '已超时',
      disputed: '申诉中',
    },
    pagination: {
      total: '共 {{total}} 条',
      page: '第 {{current}}/{{total}} 页',
      pageSize: '每页笔数：',
      items: '笔',
    },
    detail: {
      title: '订单详情',
      basicInfo: '基本信息',
      tradeInfo: '交易信息',
      paymentDeadline: '付款截止',
      cancelledAt: '取消时间',
      cancelReason: '取消原因',
      fee: '手续费',
      view: '查看',
      close: '关闭',
    },
    dispute: {
      title: '申诉仲裁',
      hint: '此订单处于申诉状态。请选择仲裁方式：「完成交易」将资产转给买家；「退款取消」将资产退还给卖家。',
      arbitrate: '仲裁',
      selectAction: '请选择仲裁方式',
      actionComplete: '完成交易（释放给买家）',
      actionRefund: '退款取消（退还给卖家）',
      completeHint: '确认后，冻结的数字资产将转入买家钱包，订单标记为已完成。',
      reasonLabel: '取消原因',
      reasonPlaceholder: '请说明退款原因（例如：买家未完成付款）',
      reasonRequired: '退款取消时必须填写原因',
      resolveFailed: '仲裁操作失败，请稍后再试',
      submit: '确认仲裁',
    },
  },

  // 通用
  common: {
    welcome: '欢迎使用丰盈钱包运营后台',
  },
}
