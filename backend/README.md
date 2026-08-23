# P2P Exchange — Backend

以 [go-zero](https://go-zero.dev) 框架建立的 REST API 服務，採 Clean Architecture 分層。

> Legacy（原本以 `p2p.go`/`p2p.api` 為進入點的 go-zero 三層架構版本）已於遷移完成後移除，
> 目前 `cmd/v1-p2p-exchange-service` 是唯一的服務入口。

---

## 目錄

- [專案結構](#專案結構)
- [前置需求](#前置需求)
- [快速啟動](#快速啟動)
- [架構說明](#架構說明)
- [Swagger UI](#swagger-ui)
- [設定檔說明](#設定檔說明)

---

## 專案結構

| 路徑 | 說明 |
|------|------|
| `cmd/v1-p2p-exchange-service/main.go` | 程式進入點（main，`uber-fx` 依賴注入啟動） |
| `cmd/v1-p2p-exchange-service/swagger.api` | API 定義檔（go-zero DSL，僅用於產生 Swagger 文件） |
| `cmd/v1-p2p-exchange-service/Makefile` | Swagger 產生指令 |
| `cmd/v1-p2p-exchange-service/internal/constants/config.yaml` | 服務設定範本（機敏欄位由環境變數注入） |
| `cmd/v1-p2p-exchange-service/internal/config/` | 設定結構定義 |
| `cmd/v1-p2p-exchange-service/internal/model/entity/` | 資料表對應的 entity |
| `cmd/v1-p2p-exchange-service/internal/repository/` | 資料存取層 |
| `cmd/v1-p2p-exchange-service/internal/service/` | 商業邏輯層 |
| `cmd/v1-p2p-exchange-service/internal/interfaces/` | Request / Response 結構 |
| `cmd/v1-p2p-exchange-service/internal/server/handlers/` | HTTP Handler |
| `cmd/v1-p2p-exchange-service/internal/wsserver/` | WebSocket 伺服器 |
| `go.mod` | Go 模組定義（整個 backend 單一模組） |
| `migrations/` | 資料庫遷移檔（legacy 與 v1 共用同一個 DB） |
| `pkg/` | 跨服務共用套件（NATS/Redis 連線、ECPay、Tron、通知等） |
| `internal/errors`、`internal/response`、`internal/infra/rdb` | 共用基礎套件 |

---

## 前置需求

| 工具 | 版本 | 安裝方式 |
|------|------|----------|
| Go | >= 1.23 | https://go.dev/dl |
| goctl | >= 1.10 | `go install github.com/zeromicro/go-zero/tools/goctl@latest` |
| goctl-swagger | 最新 | `go install github.com/zeromicro/goctl-swagger@latest` |

---

## 快速啟動

```bash
go mod tidy
cd cmd/v1-p2p-exchange-service
go run main.go -f internal/constants/config.yaml
```

必要的機敏環境變數（未設定會啟動失敗，見 [設定檔說明](#設定檔說明)）：`APP_JWT_ACCESS_SECRET`、`BACKEND_JWT_ACCESS_SECRET`、`DATABASE_DSN`、`REDIS_ADDR`。

---

## 架構說明

Clean Architecture 四層，依賴方向由外而內：

**handlers** — `internal/server/handlers/`

HTTP 層：解析請求、呼叫 service、組裝回應。不含商業邏輯。

**service** — `internal/service/`

商業邏輯層，依賴 repository 介面（不依賴實作）。

**repository** — `internal/repository/`

資料存取層，封裝 SQL / Redis 操作，回傳 `model/entity` 的資料結構。

**model/entity** — `internal/model/entity/`

對應資料表的純資料結構。

---

## Swagger UI

服務啟動後，瀏覽器開啟 `http://localhost:8888/swagger`。

新增 API 時：
1. 在 `cmd/v1-p2p-exchange-service/swagger.api` 補上對應的 type 與 route 定義
2. 執行 `cd cmd/v1-p2p-exchange-service && make swagger` 重新產生 `internal/swagger/dist/spec.json`

> `swagger.api` 僅用於產生文件，不用於程式碼生成；handler/service/repository 皆為手動維護。

---

## 設定檔說明

設定檔位於 `cmd/v1-p2p-exchange-service/internal/constants/config.yaml`，機敏欄位（JWT Secret、DB DSN、Redis、NATS、Tron、ECPay 憑證）一律透過環境變數注入，不寫入設定檔。詳細環境變數清單見 `internal/config/config.go` 的欄位標籤。
