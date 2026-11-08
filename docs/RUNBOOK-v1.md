# P2P Exchange v1 啟動與驗收手冊

本手冊說明如何在本機啟動 v1(後端 + 後台 Web + App)並完成基本驗收。

---

## 0. 前置需求

- Node 22 LTS(`shared/`、`frontend/app/` 已附 `.nvmrc`,可用 `nvm use`)
- Go >= 1.25(見 `backend/go.mod`)
- Docker(啟動本機 Redis、NATS)
- 可連線的 PostgreSQL(專案使用 Neon)

---

## 1. 本機依賴:Redis 與 NATS

```bash
docker compose up -d     # Redis :6379、NATS :4222(JetStream,監控 :8222)
```

NATS 為選用:`NATS_URL` 留空時 NATS 功能關閉,服務仍可啟動。

---

## 2. 後端:套用 migration

目前 migration 共 4 支,位於 `backend/migrations/`,需依序套用:

```bash
cd backend
psql "<DATABASE_DSN>" -f migrations/001_init.sql
psql "<DATABASE_DSN>" -f migrations/002_seeds.sql
psql "<DATABASE_DSN>" -f migrations/003_push_token.sql
psql "<DATABASE_DSN>" -f migrations/004_deposits_withdrawals.sql
```

| 檔案 | 內容 |
|------|------|
| `001_init.sql` | 核心資料表(app_users、backend_users、currencies、payment_methods、listings、orders、escrow_records、wallets、wallet_ledgers、order_status_logs) |
| `002_seeds.sql` | 開發/測試種子資料,冪等可重複執行 |
| `003_push_token.sql` | `app_users.expo_push_token` |
| `004_deposits_withdrawals.sql` | crypto_deposits、crypto_withdrawals、fiat_deposits、fiat_withdrawals |

種子帳號(密碼皆為 `a12345678`):

- 後台:`admin001`
- App:`testdemo001`、`testdemo002`、`testdemo003`、`seller001`、`buyer001`

---

## 3. 後端:啟動服務

必要環境變數(未設定會啟動失敗):`APP_JWT_ACCESS_SECRET`、`BACKEND_JWT_ACCESS_SECRET`、`DATABASE_DSN`、`REDIS_ADDR`。
其他選用變數(Redis 密碼、NATS、Tron、ECPay)見 `backend/cmd/v1-p2p-exchange-service/internal/config/config.go`。

```bash
cd backend
go mod tidy
cd cmd/v1-p2p-exchange-service
go run main.go -f internal/constants/config.yaml     # REST :8888、WebSocket :8889
```

冒煙測試:

```bash
# 健康檢查(檢查 DB;Redis/NATS 有設定時才納入)
curl http://localhost:8888/healthz

# 版本
curl http://localhost:8888/api/v1/version

# App 登入,取得 token
curl -X POST http://localhost:8888/app/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"testdemo001","password":"a12345678"}'

# 後台登入
curl -X POST http://localhost:8888/backend/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin001","password":"a12345678"}'
```

Swagger UI:`http://localhost:8888/swagger`。

各 API 的完整路由與請求格式以 `backend/cmd/v1-p2p-exchange-service/swagger.api` 為準。API 分為 `/app/*`(App)與 `/backend/*`(後台),`/v1/*` legacy 路由已移除。

---

## 4. 後台 Web

```bash
cd frontend/web
yarn install
yarn dev          # http://localhost:3000
```

`frontend/web/.env` 已設 `VITE_API_URL=http://localhost:8888`;`.env.production` 指向 `https://p2p-exchange-api.passon.tw`。

頁面路由:`/login`、`/`(Dashboard)、`/members`、`/orders`、`/fiat-withdrawals`、`/platform-wallet`。

---

## 5. App

```bash
cd frontend/app
yarn install
yarn start        # 等同 expo start
```

設定後端位址:複製 `frontend/app/.env.example` 為 `.env`,修改 `EXPO_PUBLIC_API_BASE_URL`(擇一):

- iOS 模擬器:`http://localhost:8888`
- Android 模擬器:`http://10.0.2.2:8888`
- 實體裝置/正式:`https://p2p-exchange-api.passon.tw`

---

## 6. 驗收

驗收流程請對照 `docs/SPEC-v1.md` 與 `openspec/specs/` 各領域規格,至少涵蓋:

1. 後台以 `admin001` 登入,可看到會員、訂單、法幣提領與平台錢包頁面。
2. App 以 `seller001` / `buyer001` 登入,可建立掛單、下單、付款、確認放行。
3. App 建立的訂單,於後台 `/orders` 重新整理後可見且狀態一致。
4. 後台對會員入金(`POST /backend/members/:id/deposit`)後,App 錢包餘額更新。

---

## 7. CI/CD 注意事項

- `db-migrate.yml` 在 `backend/migrations/**` 變更推送到 main 或手動觸發時,以 `backend/Dockerfile.migrate` 建置 `cmd/migrate`,並在 K8s 以 Job 執行(讀取 Secret `p2p-exchange-db` 的 `DATABASE_URL`)。已套用的檔案記錄在 `schema_migrations`,重複執行會略過。
- `002_seeds.sql` 的預設帳號與密碼為刻意保留,migrate 會與其他檔案一起套用,包含正式環境。
- 後端與 Web 映像由 `ci-build-push.yml` 建置並更新 gitops image tag。
- iOS 發版由 `expo-ios-testflight.yml` 觸發,流程見根目錄 `README.md`。
