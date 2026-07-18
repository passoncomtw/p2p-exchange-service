// @p2p/shared — 共用核心匯出入口
// 型別 / DTO、驗證、狀態機、API client 契約、i18n、設計 token

// 領域型別
export * from './domain/order';
export * from './domain/status';

// 驗證
export * from './validation/order';

// DTO
export * from './dto/order';

// API client
export * from './api/client';

// i18n
export * from './i18n';

// 注意：設計 token 不跨平台共用。Web 定義於 frontend/web/src/theme，
// App 定義於 frontend/app/src/theme，各平台各自擁有。
