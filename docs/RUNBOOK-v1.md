# P2P Exchange v1 啟動與驗收手冊

本手冊說明如何在本機啟動 v1 原型(後端 + Web + App)並完成端到端驗收。
沙箱環境無法連線 Neon DB、也無法編譯 Go,故以下步驟需在你的機器執行。

---

## 0. 前置需求

- Node 22 LTS(專案已附 `.nvmrc`,可用 `nvm use`)
- Go >= 1.23(後端)
- 可連線的 PostgreSQL(既有 `backend/etc/config.yaml` 指向 Neon)

---

## 1. 後端:套用 migration 與 seed

```bash
cd backend

# 套用既有與 v1 migration(擇一方式)
# 方式 A:以 psql 逐一套用
psql "<DATABASE_DSN>" -f migrations/001_init.sql
psql "<DATABASE_DSN>" -f migrations/002_orders.sql
psql "<DATABASE_DSN>" -f migrations/003_v1.sql

# 植入 demo_user、trader_* 與樣本掛單(idempotent,可重複執行)
go run cmd/seed/main.go -f etc/config.yaml
```

`003_v1.sql` 只新增 `listings.payment_method_label` 欄位。seed 會建立 `demo_user` 與三位 `trader_*`,並植入 4 筆他人掛單(open / completed / cancelled)。

---

## 2. 後端:啟動服務

```bash
cd backend
go mod tidy
make dev          # 等同 go run p2p.go -f etc/config.yaml,監聽 :8888
```

冒煙測試:

```bash
# 建立一筆掛單
curl -X POST http://localhost:8888/v1/orders \
  -H 'Content-Type: application/json' \
  -d '{"type":"buy","asset":"USDT","fiat":"TWD","price":32,"quantity":10,"paymentMethod":"bank_transfer"}'

# 我的掛單
curl http://localhost:8888/v1/orders/mine

# 後台全部掛單
curl http://localhost:8888/v1/admin/orders

# 後台標記完成(將 :id 換成實際 id)
curl -X POST http://localhost:8888/v1/admin/orders/<id>/complete
```

回傳格式為 `{ "code": 0, "message": "success", "data": ... }`。

---

## 3. Web

```bash
cd frontend/web
npm install
npm run dev       # http://localhost:3000
```

`.env` 已設 `VITE_API_URL=http://localhost:8888`。

- `/` 掛單頁、`/my-orders` 我的掛單、`/admin` 後台、`/admin/orders/:id` 詳情。

---

## 4. App

```bash
cd frontend/app
yarn install      # 或 npm install
npx expo start
```

設定後端位址(擇一):

- iOS 模擬器:`EXPO_PUBLIC_API_BASE_URL=http://localhost:8888`
- Android 模擬器:`EXPO_PUBLIC_API_BASE_URL=http://10.0.2.2:8888`

可建立 `frontend/app/.env`(參考 `.env.example`)。

---

## 5. 端到端驗收

### Web
1. 於 `/` 選買/賣、填單價與數量、選付款方式,送出。空白或 ≤ 0 應顯示驗證錯誤。
2. 成功後導向 `/my-orders`,該筆狀態為「待成交(open)」。
3. 於「我的掛單」取消該筆,狀態變「已取消(cancelled)」。
4. 於 `/admin` 應看到全部訂單(含 seed 他人訂單);點入詳情確認資訊與狀態,與使用者端一致。
5. 對一筆 open 訂單按「標記完成」,狀態變「已完成(completed)」。

### App
1. 於「掛單」分頁建立訂單(同上驗證)。
2. 切到「我的掛單」分頁,看到該筆 open;下拉可重新整理。
3. 取消該筆 → cancelled。

### 跨平台
- App 建立的訂單,於 Web `/admin` 重新整理後應可見,且狀態正確。

---

## 6. 沙箱已完成的驗證

- `shared`:`tsc --noEmit` 通過;API client / 驗證 / 狀態機以 mock 後端執行測試全數通過。
- `frontend/web`:`vite build` 成功(含 shared TS 匯入)。
- `frontend/app`:六個新檔以 esbuild 語法檢查通過(無法在沙箱跑 Expo bundling)。
- 後端 Go 與實際 DB:沙箱無法連線/編譯,需依本手冊在本機驗證。
