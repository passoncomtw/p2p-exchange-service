# Claude Code Handoff:App 交易流程(完整 P2P)

對象:在本機以 Claude Code 實作並即時驗證(可跑 Go 後端、migration、Expo)。
本文件是實作計畫 + API/狀態對應,不含程式碼;請依此實作並在本機測試。

---

## 0. 目標

在 App 底部新增「交易」分頁,顯示「除自己以外」的掛單;點擊看詳情;依掛單類型選擇買幣/賣幣(接單)。接單後該掛單對所有人顯示為「交易中」,參與雙方可看到訂單狀態並推進流程(付款 → 確認放行 → 完成)。沿用既有 `/app/listings` + `/app/orders` 後端(登入綁定)。

---

## 1. 架構前提(重要)

- 完整 P2P 沿用既有後端,**資料一律登入綁定**(JWT),不再用 `/v1` demo_user。
- 因此 App 的「掛單 / 我的掛單」要從 `/v1`(shared fetch client、demo_user)改為既有 `listingsApi`(`/app/listings`,httpClient 帶 token)。
- Web 後台仍透過同一張 `listings` 表(`/v1/admin/orders`)看得到 App 建立的掛單,跨平台可見性不變。
- 登入、persist 自動登入已完成(`V1Navigation` 閘門 + `redux-persist` 白名單 `auth`)。

### 需先驗證的兩個既有疑點

1. **回應信封層數**:既有 `listingsApi` / `p2pOrdersApi` 用 `response.data.data.data`(triple),代表後端疑似「雙層包裝」(`p2p.go` 的 `SetOkHandler` 疊加 handler 內的 `response.Success`)。請先 `curl` 一支 `/app/listings` 確認實際層數,並讓所有 App API 解析一致。(註:web 的 shared fetch client 目前假設「單層」,若後端是雙層,web 端解析需另修——不在本次範圍,但請記錄。)
2. **付款方式端點**:`bankCardsApi` 打的是 `/bankcards`,但後端實際路由是 `/app/payment-methods`(見 `backend/internal/handler/routes.go`)。請改用 `/app/payment-methods`,不要用 `/bankcards`。

---

## 2. 既有 API(可直接用)

`frontend/app/src/apis/`:

- `listingsApi`(`/app/listings`,JWT)
  - `list(params)` 全部掛單、`getById(id)`、`create(params)`、`mine(params)`、`cancel(id)`
- `p2pOrdersApi`(`/app/orders`,JWT)
  - `list(params)`、`getById(id)`、`create({listingId, cryptoAmount})`、`markPaid(id)`、`confirm(id)`、`cancel(id, reason)`、`dispute(id, reason)`

需新增:`paymentMethodsApi`(`/app/payment-methods`,JWT;鏡像 listingsApi 解析慣例)
- `list()` → `GET /app/payment-methods` → `{ list: PaymentMethodItem[] }`
- `create({type:'bank_transfer', bankName, accountName, accountNumber})` → `POST` → `{ id }`
- `remove(id)` → `DELETE /app/payment-methods/:id`

PaymentMethodItem:`{ id, type, bankName, accountName, accountNumber, isActive }`(後端 `types.go`)。

---

## 3. 資料模型對應

ListingItem(`interfaces/listing.ts`):`{ id:number, userId, type:'buy'|'sell', cryptoCurrency, fiatCurrency, totalAmount(掛單數量), remainingAmount, price, minOrderFiat, maxOrderFiat, paymentMethodId, status:'active'|'paused'|'completed'|'cancelled', createdAt }`

Order(`interfaces/order.ts`):`{ id, orderNo, listingId, listingType, sellerId, buyerId, cryptoAmount, price, fiatAmount, totalAmount, paymentMethodId, status:'matched'|'paid'|'releasing'|'completed'|'cancelled'|'timeout'|'disputed', paymentDeadline, paidAt, confirmedAt, completedAt, cancelledAt, cancelReason, createdAt }`

顯示對應:
- 掛單「總額(法幣)」= `price * totalAmount`(ListingItem 沒有法幣總額欄位,需自算)。
- 我的掛單狀態顯示:`active→待成交`、`completed→已完成`、`cancelled→已取消`、`paused→暫停`(v1 不用 paused)。

---

## 4. 後端行為(沿用,勿改;除非要排除自己需加 query)

`AppCreateOrder`(`backend/internal/logic/apporderlogic.go`)接單邏輯:
- 取 listing,須 `status='active'` 且 `remainingAmount >= cryptoAmount`。
- `fiatAmount = cryptoAmount * price`,需落在 `minOrderFiat..maxOrderFiat`。
- 角色指派:listing.type=`sell` → 接單者為 **buyer**,listing.user 為 seller;listing.type=`buy` → 接單者為 **seller**,listing.user 為 buyer。
- 付款方式:
  - listing 有 `paymentMethodId`(sell 掛單必有)→ 用之。
  - listing.type=`buy` 且無 paymentMethodId → 取「接單者(seller)」自己的第一個 active 付款方式;**沒有則回 400**。
- 建立 order(status=`matched`)、escrow lock、扣減 listing.remainingAmount、寫 status log。
- 手續費 v1 全 0。

`AppMyListings`/`AppListListings`:目前**不排除自己**。交易分頁需「排除自己」→ 兩種做法擇一:
- 前端過濾:`listingsApi.list({status:'active'})` 後濾掉 `userId === currentUser.id`(最簡,建議)。
- 或後端加 `excludeSelf` query(較乾淨,需改 Go)。

訂單動作(`apporderlogic.go`):
- `markPaid`:buyer,`matched→paid`(設 paidAt)。
- `confirm`:seller,`paid→releasing→completed`(escrow release,設 confirmedAt/completedAt)。
- `cancel`:buyer 或 seller,**僅 matched 可取消** → `cancelled`(escrow refund)。
- `dispute`:buyer 或 seller,**僅 paid** → `disputed`。

> 注意:`AppGetOrder` 目前未驗證呼叫者是否為當事人;列表 `AppListOrders` 以 `role`+uid 篩選。詳情頁請以 `p2pOrdersApi.list({role})` 取得當事訂單,或直接 getById 後比對 buyerId/sellerId == currentUser.id 決定可執行動作。

---

## 5. 狀態流程與角色動作

```
接單(create order) ─▶ matched(交易中)
matched ──買方:已付款──▶ paid(已付款,待放行)
paid    ──賣方:確認收款──▶ completed(已完成)   // 後端一次推進 releasing→completed
matched ──買方/賣方:取消──▶ cancelled(已取消)
paid    ──買方/賣方:申訴──▶ disputed(申訴中)   // 後台 resolve(本次 App 不做)
(timeout 由系統,v1 不主動實作)
```

角色動作可見性(訂單詳情):
- 我是 **buyer** 且 status=`matched`:顯示「我已付款」(markPaid)、「取消訂單」。
- 我是 **seller** 且 status=`paid`:顯示「確認收款」(confirm)。
- 我是 buyer/seller 且 status=`matched`:可「取消訂單」。
- status=`paid` 且我是當事人:可「申訴」(dispute,選配)。
- `completed`/`cancelled`/`disputed`:只顯示狀態,無動作。

狀態色(沿用 App theme,可擴充):matched=待成交橙、paid=藍(資訊)、releasing=藍、completed=綠、cancelled=灰、timeout=灰、disputed=紅。

---

## 6. 導覽結構

`frontend/app/src/navigation/v1.tsx`:
- 將 `Main` 由「View(AppBar)+ Tab」改為 **native stack**:
  - `Tabs`(AppBar + Bottom Tabs)
  - `ListingDetail`(掛單詳情/接單)
  - `OrderDetail`(訂單詳情/狀態流程)
  - `AddPaymentMethod`(新增收款帳戶)
- 底部 Tabs(4 個):`掛單(create-outline)` / `交易(swap-horizontal-outline)` / `我的掛單(list-outline)` / `訂單(receipt-outline)`。
- 交易/訂單清單的列 → push 到對應詳情。

---

## 7. 畫面規格

### 7.1 交易(市場)TradeMarketScreen
- 資料:`listingsApi.list({ status:'active' })`,前端濾掉自己的 `userId`。
- UI:卡片列表(沿用我的掛單卡片樣式),每張顯示 類型標籤、單價、可交易數量(remainingAmount)、付款方式、建立者(可用 userId 或之後補 username)。
- 狀態:skeleton 載入、頁內錯誤+重新載入、空狀態圖示(沿用 V1MyOrdersScreen 樣式)。
- 互動:點卡片 → `ListingDetail`。
- 排序:`createdAt` desc(後端已排序)。

### 7.2 掛單詳情 / 接單 ListingDetailScreen
- 資料:`listingsApi.getById(id)`。
- 顯示完整掛單資訊(類型、單價、數量、總額、付款方式、建立時間)。
- 動作:依 listing.type 顯示主按鈕:
  - `sell` 掛單 → 「買幣」(接單者為 buyer)。
  - `buy` 掛單 → 「賣幣」(接單者為 seller;**需有收款帳戶**,沒有先導去 AddPaymentMethod)。
- 接單:`p2pOrdersApi.create({ listingId, cryptoAmount })`。
  - v1 採「整單接走」:`cryptoAmount = remainingAmount`(或讓使用者輸入,需落在 min/max 法幣區間)。
  - 成功 → 取得 order → push `OrderDetail`(該訂單),狀態 matched。
- 邊界:listing 已非 active(被接走)→ 顯示「已被接單/不可交易」並禁用按鈕(後端會回 400,前端也先判斷)。

### 7.3 訂單(我的交易)OrdersScreen
- 資料:`p2pOrdersApi.list()`(我參與的,buyer 或 seller)。可加狀態篩選 tabs(全部/交易中/已付款/已完成/已取消)。
- UI:卡片列表,顯示 orderNo、對手角色(我是買/賣)、單價、數量、法幣總額、狀態標籤、建立時間。
- skeleton/error/empty 同上。
- 點擊 → `OrderDetail`。

### 7.4 訂單詳情 OrderDetailScreen
- 資料:`p2pOrdersApi.getById(id)` + 由 currentUser.id 比對 buyerId/sellerId 判斷角色。
- 顯示:完整訂單欄位 + 狀態 + 對手付款方式(若 buyer 需付款,顯示賣方收款資訊:可由 paymentMethodId 取得;後端目前未在 order 回傳付款方式明細 → 若需顯示帳號,需後端補或另查 `/app/payment-methods`,請評估)。
- 動作:依第 5 節角色/狀態規則顯示按鈕,呼叫對應 API,成功後重新 getById 更新狀態。
- 動作確認:用 `Alert` 二次確認(已付款/確認收款/取消/申訴)。

### 7.5 新增收款帳戶 AddPaymentMethodScreen
- 表單:銀行名稱、戶名、帳號(type 固定 `bank_transfer`)。
- `paymentMethodsApi.create(...)` → 成功返回上一頁並可選用。
- 進入點:建立 sell 掛單時、接 buy 掛單時若無收款帳戶。

---

## 8. 掛單(建立)遷移到 /app/listings

`V1CreateOrderScreen` 改用 `listingsApi.create`:
- 參數:`{ type, price, totalAmount: quantity, minOrderFiat: 0, maxOrderFiat: price*quantity, paymentMethodId, cryptoCurrency:'USDT', fiatCurrency:'TWD', paymentTimeLimit:30 }`。
- `sell`:必須選一個收款帳戶(paymentMethodId);若無 → 導去 AddPaymentMethod。
- `buy`:paymentMethodId 可為 null。
- **付款方式 UI 調整**:legacy 後端付款方式只有 `bank_transfer`(payment_methods.type CHECK),**不支援「超商代收」**。建立掛單的付款方式改為「選擇收款銀行帳戶」(sell)。請移除/停用 convenience_store 選項(或保留 UI 但後端不接受 → 建議移除以免誤導)。
- 驗證沿用 shared `validateCreateOrder` 的 price/quantity > 0;type/asset/fiat 固定。

`V1MyOrdersScreen` 改用 `listingsApi.mine()`,狀態映射顯示(active→待成交…)。卡片可加「remainingAmount / 已成交」資訊(選配)。

---

## 9. i18n(新增鍵,放 shared `order.*` 或 app 端皆可)

- nav:`trade='交易'`、`orders='訂單'`
- 訂單狀態:`matched='交易中'`、`paid='已付款'`、`releasing='放行中'`、`completed='已完成'`、`cancelled='已取消'`、`timeout='已超時'`、`disputed='申訴中'`
- 動作:`takeBuy='買幣'`、`takeSell='賣幣'`、`markPaid='我已付款'`、`confirmReceipt='確認收款'`、`cancelOrder='取消訂單'`、`dispute='申訴'`、`addPayment='新增收款帳戶'`
- 角色:`asBuyer='我是買方'`、`asSeller='我是賣方'`
- 訊息:接單成功/失敗、付款成功、確認成功、取消成功、無收款帳戶提示等。

---

## 10. 本機測試清單

1. 後端:套用 migrations(001/002/003)、`go run cmd/seed`(已含 demo_user 與 trader_* + 樣本掛單)、`make dev`。
2. 確認回應信封層數(curl `/app/listings` 帶 testdemo001 token)。
3. App:設 `EXPO_PUBLIC_API_BASE_URL`,`npx expo start`。
4. 以 testdemo001 登入 → 建立一筆 sell 掛單(先新增收款帳戶)→ 我的掛單可見(active)。
5. 以 testdemo002 登入(另一裝置/模擬器)→ 交易分頁看到 testdemo001 的掛單(排除自己)→ 點詳情 → 買幣(接單)→ 訂單變 matched。
6. testdemo001 的「我的掛單」該筆 remainingAmount 減少;雙方在「訂單」分頁看到該訂單與狀態。
7. buyer(testdemo002)按「我已付款」→ paid;seller(testdemo001)按「確認收款」→ completed。
8. 另測 matched 取消 → cancelled;buy 掛單接單(賣幣)需接單者有收款帳戶。
9. 跨平台:Web 後台 `/admin` 仍看得到這些掛單。

---

## 11. 風險 / 待確認

- 回應信封層數(triple vs single)— 先 curl 驗證,統一所有 App API。
- `/bankcards` 應改 `/app/payment-methods`。
- 訂單詳情若要顯示賣方收款帳號,order 回傳未含付款方式明細,需後端補欄位或前端另查(評估後決定)。
- 「排除自己」採前端過濾(簡單)或後端 query(乾淨)。
- web 的 shared fetch client 信封假設與後端若不一致,需另案修(不在本次範圍)。
- 沙箱(Cowork)無法執行驗證,以上流程務必在本機實測。
