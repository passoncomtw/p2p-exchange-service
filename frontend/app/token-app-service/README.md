# Token App Service

基於 Expo 和 React Navigation 的 React Native 應用程式，整合 Redux Toolkit、Redux Saga 進行狀態管理和非同步流程控制。

## 📋 目錄

- [功能特性](#功能特性)
- [技術架構](#技術架構)
- [快速開始](#快速開始)
- [專案結構](#專案結構)
- [Redux 狀態管理](#redux-狀態管理)
- [登入登出功能](#登入登出功能)
- [使用 Expo Redux DevTools Plugin 除錯](#使用-expo-redux-devtools-plugin-除錯)
  - [快速設定](#快速設定)
  - [使用方式](#使用方式)
  - [測試流程](#測試流程)
- [開發指南](#開發指南)
- [App Icon 設定](#app-icon-設定)

---

## 功能特性

### ✨ 核心功能

- ✅ **React Navigation** - 使用 Native Stack 和 Bottom Tab 導航
- ✅ **Redux Toolkit** - 現代化的 Redux 狀態管理
- ✅ **Redux Saga** - 處理複雜的非同步流程
- ✅ **TypeScript** - 完整的類型安全
- ✅ **認證系統** - 登入/登出功能，條件式導航
- ✅ **Redux DevTools** - 支援 React Native Debugger 除錯

### 🎨 UI/UX 特性

- 🌓 **主題支援** - 基於系統外觀的深色/淺色主題
- 📱 **跨平台** - iOS、Android、Web 支援
- 🔗 **深層連結** - 自動配置深層連結和 URL 處理
- 🎯 **Edge-to-Edge** - Android 邊緣到邊緣配置

---

## 技術架構

### 主要技術棧

```
React Native (0.81.5)
├── Expo (~54.0.1)
├── React (19.1.0)
├── TypeScript (~5.9.2)
└── Navigation
    ├── @react-navigation/native (^7.1.8)
    ├── @react-navigation/native-stack (^7.3.16)
    └── @react-navigation/bottom-tabs (^7.4.0)

Redux 生態系統
├── @reduxjs/toolkit (^2.10.1)
├── react-redux (^9.2.0)
├── redux-saga (^1.4.2)
└── redux-thunk (^3.1.0)
```

### 架構設計

```
src/
├── navigation/
│   ├── index.tsx               # 條件式導航（根據驗證狀態）
│   ├── screens/                # 畫面組件
│   │   ├── PublicScreen/       # 公開頁面
│   │   ├── LoginScreen/        # 登入頁面
│   │   ├── HomeScreen/         # 主頁（已驗證）
│   │   └── ExploreScreen/      # 探索頁面（已驗證）
│   └── store/                  # Redux Store
│       ├── configureStore.ts   # Store 配置
│       ├── hooks.ts            # Typed hooks
│       ├── actions/            # Saga action creators
│       ├── sagas/              # Redux Saga
│       └── slices/             # Redux slices (reducers)
├── components/                 # 共用組件
├── constants/                  # 常數定義
└── hooks/                      # 自訂 hooks
```

---

## 快速開始

### 前置需求

- Node.js (建議 v18 或 v20)
- Yarn 或 npm
- Expo CLI
- iOS Simulator (macOS) 或 Android Studio

### 安裝

```bash
# 克隆專案
cd /Users/tomaslin/Projects/passontw-app-services/cmd/token-app-service

# 安裝依賴
yarn install
# 或
npm install
```

### 環境變數設定

在專案根目錄建立 `.env` 檔案：

```bash
# API Base URL
EXPO_PUBLIC_API_BASE_URL=https://token-app-api.passon.tw

# Log Level (DEBUG, INFO, WARN, ERROR, NONE)
# DEBUG: 顯示所有日誌（開發環境推薦）
# INFO: 顯示一般資訊、警告和錯誤
# WARN: 只顯示警告和錯誤
# ERROR: 只顯示錯誤
# NONE: 不顯示任何日誌
EXPO_PUBLIC_LOG_LEVEL=DEBUG
```

**日誌級別說明：**

| 級別 | 顯示內容 | 適用場景 |
|------|----------|----------|
| `DEBUG` | 所有日誌（debug, info, warn, error） | 開發環境（預設） |
| `INFO` | 一般資訊、警告、錯誤 | 測試環境 |
| `WARN` | 警告和錯誤 | 預發佈環境 |
| `ERROR` | 僅錯誤 | 生產環境 |
| `NONE` | 不顯示任何日誌 | 效能測試 |

### 運行專案

```bash
# 啟動開發伺服器
  npx expo start

# 或清除快取後啟動
npx expo start -c

# 在 iOS 模擬器中運行
npx expo start --ios
# 或按 'i'

# 在 Android 模擬器中運行
npx expo start --android
# 或按 'a'

# 在網頁瀏覽器中運行
npx expo start --web
# 或按 'w'
```

---

## 專案結構

### 導航結構

```
Navigation (條件式)
├─ isAuthenticated = false
│  └─ PublicStack
│     ├─ PublicScreen        (公開頁面)
│     └─ LoginScreen         (登入頁面)
│
└─ isAuthenticated = true
   └─ RootStack
      ├─ HomeTabs            (底部 Tab 導航)
      │  ├─ HomeScreen       (主頁)
      │  └─ ExploreScreen    (探索)
      └─ NotFoundScreen      (404)
```

### Redux Store 結構

```typescript
RootState
└─ auth
   ├─ isAuthenticated: boolean    // 驗證狀態
   ├─ user: User | null           // 使用者資料
   ├─ loading: boolean            // 載入狀態
   └─ error: string | null        // 錯誤訊息
```

---

## Redux 狀態管理

### Store 配置

專案使用 Redux Toolkit 配置 Store，整合了：

- **Redux Thunk** - 簡單的非同步邏輯
- **Redux Saga** - 複雜的非同步流程控制
- **Redux DevTools** - 開發環境下的除錯工具

```typescript
// src/navigation/store/configureStore.ts
export const store = configureStore({
  reducer: {
    auth: authReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      thunk: true,
      serializableCheck: {...},
    }).concat(sagaMiddleware),
  devTools: __DEV__, // 啟用 Redux DevTools
});
```

### Typed Hooks

使用型別安全的 hooks：

```typescript
// 使用
import { useAppDispatch, useAppSelector } from '@/navigation/store/hooks';

function MyComponent() {
  const dispatch = useAppDispatch();
  const user = useAppSelector((state) => state.auth.user);
  
  // ...
}
```

### Redux Saga

處理非同步邏輯：

```typescript
// Saga Watcher
function* watchAuthSagas() {
  yield takeLatest(AUTH_SAGA_ACTIONS.LOGIN_REQUEST, loginSaga);
  yield takeLatest(AUTH_SAGA_ACTIONS.LOGOUT_REQUEST, logoutSaga);
}

// Saga Worker
function* loginSaga(action) {
  try {
    yield put(loginStart());
    const user = yield call(loginApi, action.payload);
    yield put(loginSuccess(user));
  } catch (error) {
    yield put(loginFailure(error.message));
  }
}
```

---

## 登入登出功能

### 測試帳號

```
使用者名稱: demo
密碼: password
```

### 登入流程

1. **PublicScreen** → 點擊「前往登入」
2. **LoginScreen** → 輸入帳號密碼
3. Dispatch `loginRequest` action
4. Redux Saga 處理登入邏輯：
   - dispatch `loginStart` (loading = true)
   - 呼叫 API (模擬 1 秒延遲)
   - dispatch `loginSuccess` (isAuthenticated = true)
5. Navigation 偵測狀態變化
6. 自動切換到 **RootStack** (HomeTabs)

### 登出流程

1. **HomeScreen** → 點擊「登出」按鈕
2. Dispatch `logoutRequest` action
3. Redux Saga 處理登出邏輯：
   - 清除 localStorage token
   - 延遲 300ms
   - dispatch `logout` (isAuthenticated = false)
4. Navigation 偵測狀態變化
5. 自動切換回 **PublicStack**

### Actions 流程圖

```
登入:
auth/loginRequest (Saga Action)
  └─ auth/loginStart
  └─ [API Call]
  └─ auth/loginSuccess

登出:
auth/logoutRequest (Saga Action)
  └─ [清除 Token]
  └─ auth/logout
```

---

## 使用 Expo Redux DevTools Plugin 除錯

### 為什麼使用 Expo Redux DevTools Plugin？

根據 [Expo 官方文檔](https://docs.expo.dev/debugging/devtools-plugins/#redux)，Expo 提供了原生的 Redux DevTools Plugin，無需安裝額外的應用程式。

#### 與 React Native Debugger 的比較

| 特性 | Expo Redux DevTools | React Native Debugger |
|------|---------------------|----------------------|
| 安裝方式 | ✅ npm/yarn 套件 | ❌ 需下載獨立應用程式 |
| 啟動方式 | ✅ 終端內建（`shift + m`） | ❌ 需另外啟動應用程式 |
| Port 設定 | ✅ 自動配置 | ❌ 需手動設定 Port |
| Expo 整合 | ✅ 原生支援 | ⚠️ 需額外配置 |
| 開發體驗 | ✅ 無縫整合 | ⚠️ 需切換視窗 |

**結論**：對於 Expo 專案，**Expo Redux DevTools Plugin 是更好的選擇**！

---

### 快速設定

#### 📋 前置作業

專案已配置好 Expo Redux DevTools Plugin，只需安裝依賴即可使用。

#### 步驟 1: 安裝依賴

```bash
# 使用 yarn (推薦)
yarn install

# 或使用 npm
npm install
```

這會安裝 `redux-devtools-expo-dev-plugin` 套件（已添加到 package.json）。

#### 步驟 2: 配置檢查 ✅

##### package.json

已包含：
```json
{
  "dependencies": {
    "redux-devtools-expo-dev-plugin": "^0.3.0"
  }
}
```

##### configureStore.ts

已正確配置：
```typescript
import devToolsEnhancer from 'redux-devtools-expo-dev-plugin';

export const store = configureStore({
  reducer: { auth: authReducer },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({...}).concat(sagaMiddleware),
  devTools: false, // 禁用內建的 devTools
  enhancers: (getDefaultEnhancers) => 
    getDefaultEnhancers().concat(devToolsEnhancer()), // 使用 Expo DevTools
});
```

參考：[Expo 官方文檔](https://docs.expo.dev/debugging/devtools-plugins/#redux)

---

### 使用方式

#### 啟動專案

```bash
# 清除快取並啟動
npx expo start -c

# 在 iOS 模擬器中運行
npx expo start --ios

# 在 Android 模擬器中運行
npx expo start --android
```

#### 開啟 Redux DevTools

在終端按 `shift + m`，會顯示：

```
┌─────────────────────────────────────┐
│  More tools                         │
│  ─────────────────────────────      │
│  › redux-devtools-expo-dev-plugin   │ ← 選這個！
│    @dev-plugins/react-navigation    │
└─────────────────────────────────────┘
```

選擇後，會在**瀏覽器**中自動開啟 Redux DevTools 介面！

---

### 測試流程

#### 測試帳號

```
使用者名稱: demo
密碼: password
```

#### 完整測試步驟

1. **安裝並啟動**
   ```bash
   yarn install
   npx expo start --ios
   ```

2. **開啟 DevTools**
   - 在終端按 `shift + m`
   - 選擇 `redux-devtools-expo-dev-plugin`

3. **執行登入流程**
   - PublicScreen → 點擊「前往登入」
   - LoginScreen → 輸入 `demo` / `password`
   - 點擊「登入」按鈕

4. **觀察 Redux Actions**
   在 Redux DevTools 中會看到：
   ```
   @@INIT
   auth/loginRequest          ← Saga action
   auth/loginStart            ← loading = true
   ⏳ (1秒延遲 - 模擬 API)
   auth/loginSuccess          ← isAuthenticated = true
   ```

5. **檢查 State 變化**
   ```json
   {
     "auth": {
       "isAuthenticated": true,    // ← 變化
       "user": {                    // ← 新增
         "id": "1",
         "name": "Demo User"
       },
       "loading": false,
       "error": null
     }
   }
   ```

6. **測試登出**
   - HomeScreen → 點擊「登出」按鈕
   - 觀察 State 恢復初始狀態

---

### 常見問題

#### Q: 按 shift + m 沒看到 Redux DevTools？

**A:** 
1. 確認已執行 `yarn install`
2. 重新啟動：`npx expo start -c`
3. 確認終端沒有錯誤訊息

#### Q: Redux DevTools 顯示空白？

**A:**
1. 確認 `devTools: false` 已設定（在 configureStore.ts）
2. 確認 `devToolsEnhancer()` 已添加到 enhancers
3. 重新啟動專案並清除快取

#### Q: 生產環境會包含 DevTools 嗎？

**A:** 不會！`devToolsEnhancer()` 只在開發環境運作，生產版本會自動移除。

---

### 快速開始

#### 1. 安裝套件

```bash
# 使用 yarn
yarn install

# 或使用 npm
npm install
```

#### 2. 配置已完成 ✅

專案已經正確配置了 Expo Redux DevTools Plugin：

```typescript
// src/navigation/store/configureStore.ts
import devToolsEnhancer from 'redux-devtools-expo-dev-plugin';

export const store = configureStore({
  reducer: { auth: authReducer },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({...}).concat(sagaMiddleware),
  devTools: false, // 禁用內建的 devTools
  enhancers: (getDefaultEnhancers) => 
    getDefaultEnhancers().concat(devToolsEnhancer()), // 使用 Expo DevTools
});
```

#### 3. 啟動專案

```bash
# 清除快取並啟動
npx expo start -c

# 在 iOS 模擬器中運行
npx expo start --ios

# 在 Android 模擬器中運行
npx expo start --android
```

#### 4. 開啟 Redux DevTools

在終端中按 `shift + m` 開啟 Dev Tools 選單，選擇：

```
┌─────────────────────────────────────┐
│  More tools                         │
│  ─────────────────────────────      │
│  › redux-devtools-expo-dev-plugin   │ ← 選擇這個
│    @dev-plugins/react-navigation    │
│    @dev-plugins/react-query         │
└─────────────────────────────────────┘
```

選擇後會自動在**瀏覽器**中開啟 Redux DevTools 介面！

### 使用 Redux DevTools

#### 查看 Redux State

1. 在 Redux 面板中，點擊 **"State"** 標籤
2. 查看完整的 State 樹：

```json
{
  "auth": {
    "isAuthenticated": false,
    "user": null,
    "loading": false,
    "error": null
  }
}
```

#### 追蹤 Actions

執行登入操作後，在 Actions 列表中會看到：

```
@@INIT
auth/loginRequest          ← Saga action
auth/loginStart            ← loading = true
⏳ (1秒 API 延遲)
auth/loginSuccess          ← isAuthenticated = true
```

#### 時間旅行除錯

1. 點擊任一過去的 Action
2. App 的 State 會立即回到那個時間點
3. UI 也會相應更新

這對於：
- 🐛 追蹤 Bug 發生的時間點
- 🔄 重現特定的狀態
- 🧪 測試不同的狀態組合

### 測試登入登出流程

#### 完整測試場景

```bash
# 1. 啟動 React Native Debugger (Port: 19000)
open -a "React Native Debugger"

# 2. 啟動專案
npx expo start -c
npx expo start --ios

# 3. 開啟 Debug 模式 (Cmd + D)

# 4. 執行登入流程
# - PublicScreen → 前往登入
# - LoginScreen → 輸入 demo / password
# - 觀察 Redux DevTools 中的 Actions

# 5. 執行登出流程
# - HomeScreen → 點擊登出
# - 觀察 State 變化
```

#### 觀察 State 變化

**初始狀態:**
```json
{ "auth": { "isAuthenticated": false, "user": null } }
```

**登入中:**
```json
{ "auth": { "isAuthenticated": false, "loading": true } }
```

**登入成功:**
```json
{
  "auth": {
    "isAuthenticated": true,
    "user": { "id": "1", "name": "Demo User" },
    "loading": false
  }
}
```

### 常見問題排除

#### 問題 1: "Waiting for connection"

**解決方法:**
1. 確認 Port 設定為 19000
2. 在 App 中開啟 Debug 模式
3. 重新啟動 Expo: `npx expo start -c`

#### 問題 2: Redux DevTools 空白

**檢查項目:**
1. 確認 `devTools: __DEV__` 已設定
2. 確認 `<Provider store={store}>` 已包裹 App
3. 重新啟動 React Native Debugger

#### 問題 3: 看不到 Actions

**解決方法:**
在元件中添加日誌確認 dispatch：
```typescript
const handleLogin = () => {
  console.log('Dispatching loginRequest');
  dispatch(loginRequest({ account: 'demo', password: 'password' }));
};
```

#### 問題 4: App 卡頓

**原因:** Remote Debugging 會降低效能（正常現象）

**解決方法:**
- 僅在需要時開啟 Debug 模式
- 除錯完畢後關閉 Remote Debugging

### 進階功能

#### 1. 手動 Dispatch Actions

在 Redux DevTools 底部輸入：
```javascript
{
  "type": "auth/loginRequest",
  "payload": {
    "account": "test",
    "password": "test123"
  }
}
```
按 `Cmd + Enter` dispatch！

#### 2. 查看 Network 請求

1. 點擊：`Debugger` → `Enable Network Inspect`
2. 切換到 `Network` 標籤
3. 查看所有 HTTP 請求和回應

#### 3. React DevTools

在上半部分：
- 查看元件樹
- 檢查 props 和 state
- 查看 hooks 值
- 即時修改 props 測試

#### 4. Console 除錯

在 Console 標籤：
- 查看所有 `console.log` 輸出
- 執行 JavaScript 程式碼
- 查看錯誤堆疊

### 快捷鍵

#### React Native Debugger

| 快捷鍵 | 功能 |
|--------|------|
| `Cmd/Ctrl + T` | 更改 Port |
| `Cmd/Ctrl + R` | 重新載入 App |
| `Cmd/Ctrl + K` | 清除 Console |
| `Cmd/Ctrl + [` | 上一個 Action |
| `Cmd/Ctrl + ]` | 下一個 Action |

#### 模擬器

| 快捷鍵 | 功能 | 平台 |
|--------|------|------|
| `Cmd + D` | 開啟 Dev Menu | iOS |
| `Cmd + M` | 開啟 Dev Menu | Android (macOS) |
| `Ctrl + M` | 開啟 Dev Menu | Android (Windows) |
| `Cmd + R` | 重新載入 | iOS |
| `R + R` | 重新載入 | Android |

---

## App Icon 設定

### 圖標檔案位置

所有 App 圖標存放於 `src/assets/images/` 目錄：

```
src/assets/images/
├── icon.png              # 主要圖標 (1024x1024) - App Store / 通用
├── ios-icon.png          # iOS 專用圖標 (1024x1024)
├── android-icon.png      # Android Play Store 圖標 (512x512)
├── adaptive-icon.png     # Android 自適應圖標前景 (192x192)
├── favicon.png           # Web favicon
└── splash-icon.png       # 啟動畫面圖標
```

### app.json 配置

```json
{
  "expo": {
    "name": "E幣錢包",
    "icon": "./src/assets/images/icon.png",
    "ios": {
      "supportsTablet": true,
      "bundleIdentifier": "com.passon.ecoinwallet",
      "icon": "./src/assets/images/ios-icon.png",
      "buildNumber": "1"
    },
    "android": {
      "package": "com.passon.ecoinwallet",
      "versionCode": 1,
      "adaptiveIcon": {
        "foregroundImage": "./src/assets/images/adaptive-icon.png",
        "backgroundColor": "#7B68C8"
      },
      "icon": "./src/assets/images/android-icon.png"
    },
    "plugins": [
      [
        "expo-splash-screen",
        {
          "backgroundColor": "#7B68C8",
          "image": "./src/assets/images/icon.png",
          "imageWidth": 200
        }
      ]
    ]
  }
}
```

### 圖標規格說明

| 圖標類型 | 檔案名稱 | 尺寸 | 用途 |
|----------|----------|------|------|
| 主圖標 | `icon.png` | 1024x1024 | App Store / 通用 |
| iOS 圖標 | `ios-icon.png` | 1024x1024 | iOS App Store |
| Android 圖標 | `android-icon.png` | 512x512 | Google Play Store |
| 自適應圖標 | `adaptive-icon.png` | 192x192 | Android 自適應圖標前景 |
| Favicon | `favicon.png` | 48x48 | Web 瀏覽器標籤 |

### Android 自適應圖標

Android 8.0+ 支援自適應圖標，由前景和背景組成：

- **前景圖片** (`adaptive-icon.png`): 主要圖案，192x192 PNG
- **背景顏色** (`backgroundColor`): `#7B68C8` (紫色，與 App 主題一致)

系統會根據裝置製造商的設定，將圖標裁切成不同形狀（圓形、方形、圓角方形等）。

### 啟動畫面 (Splash Screen)

```json
{
  "plugins": [
    [
      "expo-splash-screen",
      {
        "backgroundColor": "#7B68C8",
        "image": "./src/assets/images/icon.png",
        "imageWidth": 200
      }
    ]
  ]
}
```

- **背景顏色**: 紫色 `#7B68C8`，與錢包頁面主題一致
- **圖標寬度**: 200px，居中顯示

### 更換圖標步驟

1. **準備圖標檔案**
   - 確保圖標為 PNG 格式
   - iOS/主圖標: 1024x1024px
   - Android Play Store: 512x512px
   - Android 自適應圖標: 192x192px

2. **替換檔案**
   ```bash
   # 複製新圖標到專案
   cp /path/to/new-icon.png src/assets/images/icon.png
   cp /path/to/new-ios-icon.png src/assets/images/ios-icon.png
   cp /path/to/new-android-icon.png src/assets/images/android-icon.png
   cp /path/to/new-adaptive-icon.png src/assets/images/adaptive-icon.png
   ```

3. **清除快取並重建**
   ```bash
   # 清除 Expo 快取
   npx expo start -c
   
   # 預覽原生專案（檢查圖標）
   npx expo prebuild --clean
   ```

### 打包時的圖標處理

#### EAS Build (推薦)

```bash
# 打包 iOS
eas build --platform ios

# 打包 Android
eas build --platform android

# 打包全部平台
eas build --platform all
```

EAS Build 會自動：
- 根據 `app.json` 配置生成所有必要尺寸的圖標
- iOS: 生成 AppIcon.appiconset 中的所有尺寸
- Android: 生成 mipmap-* 資料夾中的所有尺寸

#### 本地開發 (Expo Go)

在 Expo Go 中開發時，圖標不會顯示為自訂圖標（會顯示 Expo 圖標）。要看到自訂圖標，需要進行 Development Build 或 Production Build。

### 常見問題

#### Q: 圖標更換後沒有更新？

**A:** 
1. 清除 Metro bundler 快取：`npx expo start -c`
2. 如果是原生打包，執行：`npx expo prebuild --clean`
3. 刪除模擬器中的 App 後重新安裝

#### Q: Android 圖標顯示白邊？

**A:** 確保 `adaptive-icon.png` 的主要內容在安全區域內（圖片中心 66% 區域），避免被裁切。

#### Q: iOS 圖標有透明背景？

**A:** iOS 不支援透明背景的圖標。確保 `ios-icon.png` 有實心背景色。

---

## 開發指南

### 程式碼規範

本專案遵循以下開發原則：

#### Clean Code 核心理念

1. **需求先行原則** - 確保開發有明確目標
2. **單一職責原則** - 每個模組只負責一個功能
3. **開放封閉原則** - 對擴展開放，對修改封閉
4. **DRY 原則** - 避免重複程式碼
5. **YAGNI 原則** - 不實作不需要的功能

#### 架構設計

- **分層架構** - 清晰的層次劃分
- **模組化設計** - 高內聚、低耦合
- **型別安全** - 完整的 TypeScript 支援

### 新增功能

#### 新增 Redux Slice

```typescript
// 1. 創建 slice
// src/navigation/store/slices/newSlice.ts
import { createSlice } from '@reduxjs/toolkit';

const newSlice = createSlice({
  name: 'newFeature',
  initialState: {...},
  reducers: {...},
});

export default newSlice.reducer;

// 2. 添加到 store
// src/navigation/store/configureStore.ts
export const store = configureStore({
  reducer: {
    auth: authReducer,
    newFeature: newFeatureReducer, // 新增
  },
  // ...
});

// 3. 更新類型
export type RootState = ReturnType<typeof store.getState>;
```

#### 新增 Saga

```typescript
// 1. 創建 saga
// src/navigation/store/sagas/newSaga.ts
function* newSaga(action) {
  try {
    yield put(startAction());
    const result = yield call(api, action.payload);
    yield put(successAction(result));
  } catch (error) {
    yield put(failureAction(error));
  }
}

export function* watchNewSaga() {
  yield takeLatest('NEW_ACTION', newSaga);
}

// 2. 添加到 root saga
// src/navigation/store/sagas/index.ts
export default function* rootSaga() {
  yield all([
    fork(watchAuthSagas),
    fork(watchNewSaga), // 新增
  ]);
}
```

#### 新增畫面

```typescript
// 1. 創建畫面組件
// src/navigation/screens/NewScreen/index.tsx
export default function NewScreen() {
  return <View>...</View>;
}

// 2. 添加到導航
// src/navigation/index.tsx
const RootStack = createNativeStackNavigator({
  screens: {
    HomeTabs: {...},
    NewScreen: { // 新增
      screen: NewScreen,
      options: { title: 'New' },
    },
  },
});
```

### 測試

```bash
# 運行 linter
npx eslint src/

# 清除快取
npx expo start -c

# 重置專案
npm run reset-project
```

### 部署

```bash
# 建立 iOS 版本
eas build --platform ios

# 建立 Android 版本
eas build --platform android

# 建立兩個平台
eas build --platform all
```

---

## 資源連結

### 官方文檔

- [React Navigation](https://reactnavigation.org/)
- [Expo](https://docs.expo.dev/)
- [Redux Toolkit](https://redux-toolkit.js.org/)
- [Redux Saga](https://redux-saga.js.org/)
- [React Native Debugger](https://github.com/jhen0409/react-native-debugger)

### 相關工具

- [React Native](https://reactnative.dev/)
- [TypeScript](https://www.typescriptlang.org/)
- [Redux DevTools](https://github.com/reduxjs/redux-devtools)

---

## 授權

本專案使用的 Demo 資源來自 [lucide.dev](https://lucide.dev/)

---

## 總結

### ✅ 專案特色

1. **現代化技術棧** - Expo + React Navigation + Redux Toolkit + Redux Saga
2. **完整的認證系統** - 登入/登出功能，條件式導航
3. **優秀的開發體驗** - Redux DevTools 支援，時間旅行除錯
4. **型別安全** - 完整的 TypeScript 支援
5. **Clean Code** - 遵循最佳實踐和開發原則

### 🚀 立即開始

```bash
# 安裝依賴
yarn install

# 啟動專案
npx expo start

# 開始除錯
# 1. 安裝 React Native Debugger
# 2. 設定 Port: 19000
# 3. 開啟 Debug 模式 (Cmd + D)
# 4. 開始開發！
```

### 📞 需要幫助？

- 查看上方的[常見問題排除](#常見問題排除)
- 參考[開發指南](#開發指南)
- 閱讀[官方文檔](#資源連結)

Happy Coding! 🎉
