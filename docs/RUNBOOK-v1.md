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

種子帳號(註解所寫密碼為 `a12345678`,但 `admin001` 的雜湊不符,見 6.1):

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

以下以 curl 驗證主流程,路由與欄位以 `swagger.api` 為準。`<...>` 為前一步回應中的值。

```bash
B=http://localhost:8888
J='Content-Type: application/json'
```

### 6.1 登入

```bash
# 後台:回應的 token 在 data.token
curl -s -X POST $B/backend/auth/login -H "$J" -d '{"username":"admin001","password":"a12345678"}'

# App:回應的 token 在 data.access_token
curl -s -X POST $B/app/auth/login -H "$J" -d '{"username":"seller001","password":"a12345678"}'
curl -s -X POST $B/app/auth/login -H "$J" -d '{"username":"buyer001","password":"a12345678"}'
```

之後的請求帶 `-H "Authorization: Bearer <token>"`,分別記為 `ADMIN`、`SELLER`、`BUYER`。

注意:`002_seeds.sql` 內 `admin001` 的雜湊與註解所寫的 `a12345678` 不符,後台登入會回 401(App 種子帳號正常)。本機驗證時可先重設:

```bash
psql "<DATABASE_DSN>" -c "create extension if not exists pgcrypto; update backend_users set password_hash=crypt('a12345678',gen_salt('bf',10)) where username='admin001';"
```

### 6.2 賣方建立收款方式與掛單

```bash
curl -s -X POST $B/app/payment-methods -H "$J" -H "Authorization: Bearer $SELLER" \
  -d '{"type":"bank_transfer","bankName":"Test Bank","accountName":"seller","accountNumber":"1234567890"}'
# 取得 data.id 為 <paymentMethodId>

curl -s -X POST $B/app/listings -H "$J" -H "Authorization: Bearer $SELLER" \
  -d '{"type":"sell","cryptoCurrency":"USDT","fiatCurrency":"TWD","totalAmount":100,"price":32,"minOrderFiat":100,"maxOrderFiat":3200,"paymentTimeLimit":15,"paymentMethodId":<paymentMethodId>}'
# 取得 data.id 為 <listingId>
```

預期:`code` 為 0;`GET /app/wallets` 顯示賣方 USDT 的凍結餘額增加 100。

### 6.3 買方下單、付款,賣方確認放行

```bash
curl -s -X POST $B/app/orders -H "$J" -H "Authorization: Bearer $BUYER" -d '{"listingId":<listingId>,"cryptoAmount":10}'
# 取得 data.id 為 <orderId>

curl -s -X PUT $B/app/orders/<orderId>/pay     -H "Authorization: Bearer $BUYER"
curl -s -X PUT $B/app/orders/<orderId>/confirm -H "Authorization: Bearer $SELLER"
curl -s $B/app/orders/<orderId> -H "Authorization: Bearer $BUYER"
```

預期:訂單 `status` 為 `completed`;買方 USDT 可用餘額增加 10,賣方凍結餘額減少 10。

### 6.4 後台檢視與入金

```bash
curl -s "$B/backend/orders" -H "Authorization: Bearer $ADMIN"                 # 看得到上述訂單
curl -s "$B/backend/members?keyword=buyer001" -H "Authorization: Bearer $ADMIN"   # 取得會員 <memberId>
curl -s -X POST $B/backend/members/<memberId>/deposit -H "$J" -H "Authorization: Bearer $ADMIN" \
  -d '{"currency":"USDT","amount":"50"}'
```

預期:訂單列表含該筆訂單且狀態一致;入金回應的 `available_balance` 增加 50。

### 6.5 介面對照

- 後台 Web(`http://localhost:3000`)以 `admin001` 登入後,Dashboard、`/members`、`/orders` 可看到上述資料。
- App 以 `seller001` / `buyer001` 登入後,可看到相同的掛單、訂單與錢包餘額。

---

## 7. CI/CD 注意事項

- `db-migrate.yml` 在 `backend/migrations/**` 變更推送到 main 或手動觸發時,以 `backend/Dockerfile.migrate` 建置 `cmd/migrate`,並在 K8s 以 Job 執行(讀取 Secret `p2p-exchange-db` 的 `DATABASE_URL`)。已套用的檔案記錄在 `schema_migrations`,重複執行會略過。
- `002_seeds.sql` 的預設帳號與密碼為刻意保留,migrate 會與其他檔案一起套用,包含正式環境。
- 後端與 Web 映像由 `ci-build-push.yml` 建置並更新 gitops image tag。
- iOS 發版由 `expo-ios-testflight.yml` 觸發,流程見根目錄 `README.md`。
