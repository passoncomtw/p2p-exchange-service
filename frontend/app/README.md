# E幣錢包 App 本地打包指南

## 前置需求

| 平台 | 必要工具 |
|------|---------|
| iOS | macOS + Xcode 16+ + Apple 開發者帳號 |
| Android | JDK 17+ + Android SDK（Android Studio） |

```bash
# 安裝 EAS CLI
npm install -g eas-cli

# 登入 Expo 帳號
eas login

# 進入 app 目錄
cd frontend/app
```

---

## Development Build（含 expo-dev-client）

開發階段使用，支援熱更新與 Expo Dev Tools。

```bash
# iOS Simulator（產出 .app，直接安裝到模擬器）
eas build --platform ios --profile development --local

# Android 實機（產出 debug .apk）
eas build --platform android --profile development --local
```

> `--local` 在本機執行，不佔用 EAS 雲端 build 次數。

---

## Preview Build（實機測試）

測試版本，不走 App Store / Play Store，直接安裝到實機。

```bash
# iOS 實機（產出 .ipa，需 Ad Hoc 或 Enterprise 憑證）
eas build --platform ios --profile preview --local

# Android 實機（產出 .apk，可直接 sideload）
eas build --platform android --profile preview --local
```

---

## 安裝到裝置

### iOS .ipa

```bash
# 使用 xcrun simctl 安裝到模擬器（.app 格式）
xcrun simctl install booted <path-to.app>

# 使用 ideviceinstaller 安裝到實機（需先安裝）
brew install ideviceinstaller
ideviceinstaller --install <path-to.ipa>

# 或透過 Xcode Devices 視窗拖曳安裝
```

### Android .apk

```bash
# 透過 adb 安裝（需啟用 USB 偵錯）
adb install <path-to.apk>
```

---

## 常用選項

| 選項 | 說明 |
|------|------|
| `--local` | 在本機執行 build，不上傳至 EAS 雲端 |
| `--platform ios` | 僅建置 iOS |
| `--platform android` | 僅建置 Android |
| `--profile <name>` | 指定 eas.json 中的 build profile |
| `--output <path>` | 指定產出檔案路徑 |

---

## EAS.json Profile 說明

| Profile | 用途 | iOS 產出 | Android 產出 |
|---------|------|---------|-------------|
| `development` | 開發（含 dev client） | .app（模擬器） | debug .apk |
| `preview` | 實機測試 | .ipa（Ad Hoc） | .apk |
| `production` | 上架 App Store / Play Store | .ipa | .aab |
| `production-apk` | 上架但產出 APK | — | .apk |
