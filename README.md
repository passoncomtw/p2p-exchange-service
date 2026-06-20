# P2P Exchange Service

## 原型設計 (Claude design)

v1 原型(使用者端手機 + 後台桌面)設計連結:

https://claude.ai/design/p/2c7b1880-aaf0-4e7c-9bd1-496d0b4560af?file=P2P+Exchange+v1.dc.html

匯出的原始檔位於 `prototypes/P2P-Exchange-v1.html`(Claude design 元件格式 `.dc.html`,需搭配設計執行階段 `support.js` 才能渲染;完整可互動版請於上方連結檢視)。詳見 `prototypes/README.md`。

## Mobile App 發版流程 (iOS / Android)

### 前置條件

在 GitHub 倉庫的 Settings → Secrets and variables → Actions 設定以下 Secrets：

| Secret | 用途 |
|---|---|
| `EXPO_TOKEN` | Expo 帳號 Access Token（[expo.dev](https://expo.dev) → Access Tokens） |
| `EXPO_APPLE_SPECIFIC_PASSWORD` | Apple ID 專用密碼（用於 TestFlight 上傳） |
| `GOOGLE_SERVICE_ACCOUNT_KEY` | Google Play 服務帳號 JSON 完整內容（CI 注入，不 commit） |

Google Play 服務帳號：`expo-ci@ecoinwallet.iam.gserviceaccount.com`，金鑰檔案名稱 `ecoinwallet-6394ffc01290.json`（已加入 `.gitignore`，由 CI 從 Secret 注入）。

### 觸發方式

Workflow 名稱：`expo-build-submit`，僅支援手動觸發：

1. GitHub → Actions → `expo-build-submit` → Run workflow
2. 輸入語義化版本號，例如 `v1.0.0`
3. 選擇平台：`ios`、`android`、或 `all`

### 流程說明

```
輸入版本號 + 平台
    ↓
版本號格式驗證（必須符合 vX.Y.Z）
    ↓
安裝依賴（yarn install）
    ↓
寫入 app.json expo.version
    ↓
EAS Build（在 Expo 伺服器上編譯，--wait 等待完成）
    ↓
iOS  → Submit to TestFlight（EXPO_APPLE_SPECIFIC_PASSWORD）
Android → 注入 Google Service Account Key → Submit to Google Play Internal
```

選擇 `all` 時，iOS 與 Android 會在 Expo 伺服器上同時編譯，完成後依序 submit。

### EAS 設定

`frontend/app/eas.json`：

- `production` build：distribution 為 `store`，iOS/Android 均使用 `credentialsSource: remote`（EAS 管理簽名憑證）
- Android submit：track 為 `internal`（Google Play Internal Testing）
- iOS submit：`ascAppId: 6477443899`，`appleTeamId: W6K7F9HNX5`

### 安全注意事項

- Google Service Account Key **不可 commit**（`.gitignore` 已封鎖 `ecoinwallet-*.json`）
- 定期到 GCP Console → IAM → 服務帳號 → 金鑰，輪換金鑰並更新 GitHub Secret

---

## 預設帳號

執行 `go run backend/cmd/seed/main.go -f backend/etc/config.yaml` 建立以下帳號。

### 後台管理員

| 帳號 | 密碼 | 角色 |
|---|---|---|
| admin001 | admin@1234 | admin |

### App 使用者

| 帳號 | 密碼 |
|---|---|
| testdemo001 | a12345678 |
| testdemo002 | a12345678 |
| testdemo003 | a12345678 |
