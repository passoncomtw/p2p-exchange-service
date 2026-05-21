# P2P Exchange — Backend

以 [go-zero](https://go-zero.dev) 框架建立的 REST API 服務。

---

## 目錄

- [專案結構](#專案結構)
- [前置需求](#前置需求)
- [快速啟動](#快速啟動)
- [常用指令](#常用指令)
- [架構說明](#架構說明)
- [新增 API 流程](#新增-api-流程)
- [Swagger UI](#swagger-ui)
- [設定檔說明](#設定檔說明)

---

## 專案結構

| 路徑 | 說明 |
|------|------|
| `p2p.go` | 程式進入點（main） |
| `p2p.api` | API 定義檔（go-zero DSL） |
| `Makefile` | 常用指令 |
| `go.mod` | Go 模組定義 |
| `etc/config.yaml` | 服務設定（Port、Host、JWT Secret 等） |
| `docs/swagger.yaml` | Swagger 規格文件 |
| `internal/config/` | 設定結構定義 |
| `internal/handler/` | HTTP Handler（解析請求、呼叫 Logic） |
| `internal/logic/` | 商業邏輯層 |
| `internal/svc/` | 服務依賴容器（DB、Cache 等注入此處） |
| `internal/types/` | Request / Response 結構（自動產生） |

---

## 前置需求

| 工具 | 版本 | 安裝方式 |
|------|------|----------|
| Go | >= 1.23 | https://go.dev/dl |
| goctl | >= 1.10 | `go install github.com/zeromicro/go-zero/tools/goctl@latest` |
| goctl-swagger | 最新 | `go install github.com/zeromicro/goctl-swagger@latest` |

---

## 快速啟動

1. 安裝依賴套件：`go mod tidy`
2. 啟動開發服務：`make dev`
3. 驗證服務是否正常：`curl http://localhost:8888/from/me`

服務啟動後會顯示監聽的 Host、Port 與 Swagger UI 網址。

---

## 常用指令

| 指令 | 說明 |
|------|------|
| `make dev` | 啟動開發服務 |
| `make swagger` | 從 `p2p.api` 產生 `docs/swagger.json` |
| `make build` | 編譯成 `bin/p2p` binary |
| `make clean` | 清除編譯產出 |

---

## 架構說明

go-zero 採用**三層架構**，每一層職責明確：

**Handler** — `internal/handler/`

負責 HTTP 層的工作：解析請求參數、呼叫 Logic、回傳結果。不應包含任何商業邏輯判斷。

**Logic** — `internal/logic/`

所有商業邏輯寫在這裡。透過 `ServiceContext` 存取資料庫等外部依賴。這是唯一需要開發者手動實作的層。

**ServiceContext** — `internal/svc/`

依賴容器，資料庫連線、Redis、第三方 SDK 都在此初始化，並注入至所有 Logic。

---

## 新增 API 流程

1. **在 `p2p.api` 定義新的 endpoint**：使用 go-zero DSL 語法描述 Request、Response 與路由
2. **執行 `goctl api go -api p2p.api -dir . --style goZero`**：自動產生 Handler、types，並更新路由註冊
3. **在對應的 Logic 實作商業邏輯**：填入 `internal/logic/` 下新增的檔案
4. **執行 `make swagger`**：更新 Swagger 文件

> Handler 和 types 是自動產生的，不需要手動修改。只有 Logic 需要自己實作。

---

## Swagger UI

服務啟動後，瀏覽器開啟 `http://localhost:8888/swagger`。

| 檔案 | 說明 |
|------|------|
| `docs/swagger.yaml` | 手動維護的 OpenAPI 規格 |
| `docs/swagger.json` | `make swagger` 自動產生 |

若兩者同時存在，服務優先使用 `swagger.json`。

---

## 設定檔說明

設定檔位於 `etc/config.yaml`，包含服務名稱、監聽介面、Port，以及 App / Backend 各自的 JWT Secret。

切換不同環境時，可建立多份設定檔（如 `etc/config-staging.yaml`），啟動時用 `-f` 參數指定。
