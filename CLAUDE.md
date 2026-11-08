# P2P Exchange Service

P2P C2C 加密貨幣交易平台：用戶之間撮合買賣加密貨幣與法幣，平台擔任擔保方（escrow）。

## Agent 角色與工作範圍

- 角色：本專案的全端開發協作者，負責 Go 後端、React Native App、React 後台 Web 與共用 TypeScript 套件。
- 涉及資金的邏輯（錢包凍結/解凍、託管、出入金、訂單狀態機）一律視為高風險，須對照 `openspec/specs/` 與既有程式，不憑印象修改。
- 所有結論以實際程式碼與指令輸出為準；文件與程式碼不一致時以程式碼為準，並回報差異。
- 範圍外不主動動作：不 commit/push、不部署、不改 `k8s/` 與 `.github/workflows/`、不碰 `backend/*.creds`、`frontend/app/credentials*` 等機敏檔，除非使用者明確要求。
- 規格驅動：新需求先對照 `openspec/specs/`，流程為 `/opsx:propose` -> `/opsx:apply`。

## 技術棧（依實際 manifest）

| 層級 | 技術 |
|------|------|
| Backend | Go 1.25、go-zero v1.10.1、uber-fx（依賴注入）、gorilla/websocket |
| 資料 | PostgreSQL（pgx v5，Neon）、Redis（go-redis v9）、NATS JetStream |
| 外部整合 | ECPay（法幣入金 webhook）、Tron（鏈上掃描、熱錢包出金） |
| App | React Native 0.85、Expo SDK 56、React 19、TypeScript 6、Redux + Redux-Saga |
| Web（後台） | React 19、MUI 9、Vite、JSX、Redux Toolkit + Saga、Formik/Yup、i18next |
| 共用 | `shared/`（`@p2p/shared`）：型別、驗證、API 契約、i18n、設計 token，Node >= 22 |
| 部署 | Docker、`k8s/` manifest、GitHub Actions（`ci-build-push`、`db-migrate`、`expo-ios-testflight`） |

注意：`openspec/project.md` 的技術棧版本（Go 1.23、React 18、MUI v5）已落後，以各 manifest 為準。

## 目錄結構

```text
backend/
  cmd/v1-p2p-exchange-service/   唯一服務入口（legacy 已移除）
    internal/
      config/ constants/config.yaml   設定
      model/entity/                   資料表 entity
      repository/                     資料存取層（依領域分資料夾）
      service/                        商業邏輯層
      interfaces/{app,backend,v1}/    Request / Response 結構
      server/handlers/                HTTP handler
      wsserver/                       WebSocket
      swagger/                        產生的 spec
    swagger.api  Makefile             API DSL 與 swagger 產生
  internal/{errors,response,infra}    共用基礎
  pkg/                                ecpay、tron、mq、ws、notification、schedule 等
  cmd/migrate/                        獨立 migration 執行檔（Dockerfile.migrate、db-migrate.yml 使用）
  migrations/                         001-004 SQL + migrations.go
frontend/app/    Expo App（有自己的 CLAUDE.md / AGENTS.md）
frontend/web/    後台 Web
frontend/pkg/    notifications
shared/src/      api design domain dto i18n validation
openspec/        規格（specs/）與專案概覽
docs/            SPEC-v1、RUNBOOK-v1、HANDOFF-*、product_spec
k8s/             backend 與 local（nats、redis）
```

## 常用指令

```bash
# 本機依賴（Redis 6379、NATS 4222/8222，含 JetStream）
docker compose up -d

# 產生 Swagger（需 goctl 與 goctl-swagger、python3）
cd backend/cmd/v1-p2p-exchange-service && make swagger

# 後端測試（目前僅 16 個 _test.go）
cd backend && go test ./...

# Web
cd frontend/web && yarn dev        # 另有 yarn build

# App
cd frontend/app && yarn start      # android / ios / web 同名 script

# shared 型別檢查
cd shared && npm run typecheck
```

後端、Web、App 目前沒有 lint 或測試的統一 script，Web 與 App 有 `eslint.config.js` 但 package.json 未定義 lint script。

## 後端慣例

- 分層：handler -> service -> repository，依賴由 uber-fx 的 `module.go` 組裝；新增元件須在對應 `module.go` 註冊。
- 錯誤處理使用 `backend/internal/errors`，回應格式使用 `backend/internal/response`（`{ code, message, data }`）。
- 資料表異動只新增 `backend/migrations/` 的編號 SQL，不修改已套用的檔案。
- 設定範本在 `internal/constants/config.yaml`，機敏欄位由環境變數注入，不寫入程式碼。
- 錢包餘額變動須維持原子性並使用 Redis 分散式鎖，修改前先讀 `openspec/specs/core/wallet/` 與 `wallet/freeze/`。

## 前端慣例

- App：API 成功/失敗訊息統一透過 Redux `pushNotification`，不在 component 直接呼叫 `Alert.alert`（需使用者確認、帶導航、input validation 除外）。細節見 `frontend/app/CLAUDE.md`。
- 型別、驗證、i18n、設計 token 優先放 `shared/`，App 與 Web 共用，避免重複定義。
- 視覺依 `DESIGN.md` 設計系統與 `shared/src/design` token。

## 核心領域狀態機

- Order：`matched -> paid -> releasing -> completed`，另有 `cancelled`、`timeout`、`disputed`。
- Listing：`active <-> paused`，`active -> completed | cancelled`。
- 完整定義見 `openspec/project.md` 與 `openspec/specs/`。

## 已知事項

- `backend/migrations/002_seeds.sql` 的預設帳號密碼（`a12345678`）為刻意保留，由 `cmd/migrate` 與其他 migration 一併套用，不要排除。

## 語言與風格

- 回覆、文件、註解使用繁體中文；程式碼命名使用英文。
- 遵循 Clean Code，避免過度工程化；程式碼以自我說明為主。
- 未被要求就不建立文件；不使用 emoji。
