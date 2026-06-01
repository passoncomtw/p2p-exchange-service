/**
 * Store 相關的 Interface 定義
 */

/**
 * 用戶錢包資訊
 */
export interface UserWallet {
  status: number;
  usefulBalance: number;
  guaranteedBalance: number;
  freezeBalance: number;
}

/**
 * 推薦人資訊
 */
export interface ReferralUser {
  id: number;
  account: string;
  email: string;
  name: string;
  type: number;
}

/**
 * 銀行資訊
 */
export interface Bank {
  id: number;
  bankCode: string;
  bankName: string;
}

/**
 * 銀行卡資料結構
 */
export interface BankCard {
  id: number;
  bankId: number;
  cardNumber: string;
  name: string;
  branchName: string; // 必需，銀行卡必定有分行名稱
  status: number;
  createdAt: string;
  bank: Bank;
}

/**
 * 用戶資訊
 */
export interface User {
  id: number;
  type: number;
  account: string;
  name: string;
  email: string;
  createAt: string;
  referralCode: string;
  referralUser?: ReferralUser;
  wallet: UserWallet;
  bankCards?: BankCard[]; // 登入時會返回銀行卡列表
}

/**
 * Auth State
 */
export interface AuthState {
  isAuthenticated: boolean;
  user: User | null;
  accessToken: string | null;
  expireIn: number | null;
  loading: boolean;
  error: string | null;
}

/**
 * 登入憑證
 */
export interface LoginCredentials {
  account: string;
  password: string;
  notificationToken?: string;
}

/**
 * 註冊憑證
 */
export interface RegisterCredentials {
  nickname: string;
  account: string;
  email: string;
  password: string;
  transactionPassword: string;
  referralCode?: string;
  notificationToken?: string;
}

/**
 * 登入響應數據
 */
export interface LoginData {
  access_token: string;
  expireIn: number;
  user: User;
}

/**
 * 註冊響應數據
 */
export interface RegisterData {
  id: number;
  type: number;
  account: string;
  name: string;
  email: string;
  createAt: string;
  referralCode: string;
  wallet: UserWallet;
}

/**
 * API 回應格式
 */
export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
  code: string;
}

