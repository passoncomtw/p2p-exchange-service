# P2P Exchange v1 開發 Handoff 規格

對象:前端工程師(Web:React + MUI v9;App:React Native + StyleSheet)。
原則:以設計 token 為準、列出所有狀態、補齊邊界情況與無障礙。所有畫面文案走 i18n(v1 僅 zh-TW)。

涵蓋畫面:Web 掛單頁、Web 我的掛單、Web 後台訂單列表、Web 訂單詳情、App 掛單、App 我的掛單。

---

## 1. 設計 token(各平台各自擁有)

設計 token 不跨平台共用:Web 定義於 `frontend/web/src/theme/index.js`,App 定義於 `frontend/app/src/theme/index.ts`。下表為兩平台共同採用的值(刻意保持一致以維持視覺語言),但由各平台各自維護。`statusColor` / `typeColor` 兩個輔助函式於兩端各自實作。

### 色彩

| Token | 值 | 用途 |
|-------|----|------|
| `colors.primary` | #FFC107 | 主 CTA、focus ring |
| `colors.primaryDark` | #FFB300 | 按鈕 hover |
| `colors.primaryDisabled` | #FFE082 | 按鈕 disabled、付款方式選中底 |
| `colors.primaryDeep` | #FF8F00 | 標記完成確認動作 |
| `colors.bgContent` | #F5F5F7 | 頁面背景 |
| `colors.bgCard` | #FFFFFF | 卡片 / 輸入框背景 |
| `colors.textPrimary` | #333333 | 主要文字 |
| `colors.textSecondary` | #666666 | 標籤 / 次要文字 |
| `colors.textTertiary` | #999999 | 提示 / 空狀態 |
| `colors.textPlaceholder` | #BFBFBF | placeholder |
| `colors.borderInput` | #D9D9D9 | 輸入框邊框 |
| `colors.borderCard` | #EBEBEB | 卡片 / 表格邊框 |
| `colors.danger` | #F44336 | 取消、賣幣、錯誤 |
| `colors.statusOpen` | #FF9800 | 狀態:待成交 |
| `colors.statusCompleted` | #4CAF50 | 狀態:已完成 |
| `colors.statusCancelled` | #9E9E9E | 狀態:已取消 |
| `colors.typeBuy` | #4CAF50 | 類型:買幣 |
| `colors.typeSell` | #F44336 | 類型:賣幣 |

輔助函式:`statusColor(status)`、`typeColor(type)`。

### 字級 / 間距 / 圓角 / 尺寸

| 類別 | Token | 值 |
|------|-------|----|
| 字級 | `typography.body` | 13px / 400 |
| 字級 | `typography.pageTitle` | 16px / 600 |
| 字級 | `typography.button` | 14px / 500 |
| 字級 | `typography.label` | 12px / 400 |
| 間距 | `spacing.sm/md/lg/2xl` | 8 / 12 / 16 / 24 px |
| 圓角 | `radius.base` | 4px |
| 尺寸 | `sizes.buttonHeight` | 36px(Web)/ 44px(App 主按鈕) |
| 尺寸 | `sizes.sidebarWidth` | 160px |
| 尺寸 | `sizes.headerHeight` | 56px |

---

## 2. 共用元件

### TypeTag(類型標籤)

| 屬性 | 規格 |
|------|------|
| 變體 | buy / sell |
| 樣式 | 實心:背景 `typeColor(type)`、白字、12px/600、圓角 4、padding 8×2 |
| 文案 | `t('order.type.buy' | 'order.type.sell')` |
| Web | `frontend/web/src/components/OrderTags.jsx` |
| App | 內嵌於卡片 styles.typeTag |

### StatusTag(狀態標籤)

| 屬性 | 規格 |
|------|------|
| 變體 | open / completed / cancelled |
| 樣式 | 淡底 + 外框:文字與框 `statusColor`、底色 `statusColor` + 8% alpha、左側 6px 圓點、12px/500 |
| 文案 | `t('order.status.<status>')` |
| 與 TypeTag 區分 | 類型為實心、狀態為淡底+圓點 |

### 主 CTA 按鈕

| 狀態 | 規格 |
|------|------|
| default | 背景 `primary`、白字、14px/500、高 36(Web)/44(App)、圓角 4、無陰影 |
| hover(Web) | 背景 `primaryDark` |
| disabled / submitting | 背景 `primaryDisabled`、停用點擊 |

### 表單輸入(數值)

| 狀態 | 規格 |
|------|------|
| default | 1px `borderInput`、圓角 4、padding 12×8、字 13 `textPrimary`、背景白 |
| focus(Web) | 框色 `primary` |
| error | 框色 `danger`,下方 12px `danger` 錯誤訊息 |
| placeholder | `textPlaceholder` |

---

## 3. Web — 掛單頁(`/`)

### Overview
使用者選買/賣、填單建立掛單。送出成功後導向我的掛單。對應 `screens/CreateOrderScreen`。

### Layout
頂部品牌列(56px)+ 置中內容(max-width 720,左右 padding 24)。表單置於卡片內,欄位間距 `lg`(16)。

### Components / 欄位
| 元件 | 變體 / 屬性 | 備註 |
|------|------------|------|
| ToggleButtonGroup | buy / sell,fullWidth | 選中:buy 綠 `typeBuy`、sell 紅 `danger`,白字 |
| Select 幣種 | options=ASSETS(USDT) | 下拉結構保留,v1 單一選項 |
| Select 法幣 | options=FIATS(TWD) | 同上 |
| TextField 單價 | type=number,min 0 | 驗證 > 0 |
| TextField 數量 | type=number,min 0 | 驗證 > 0 |
| Select 付款方式 | bank_transfer / convenience_store | |
| 總額列 | 唯讀,`price×quantity` | 即時計算,千分位 + 法幣 |
| 送出按鈕 | 主 CTA | submitting 時 disabled |

### States and Interactions
| 元素 | 狀態 | 行為 |
|------|------|------|
| 表單 | 提交且驗證失敗 | 各欄顯示 `t(messageKey)`,不送出 |
| 送出 | 進行中 | 按鈕 disabled |
| 送出 | 成功 | 頂部綠色 Snackbar「掛單建立成功」,600ms 後導向 `/my-orders` |
| 送出 | 失敗 | 紅色 Snackbar「送出失敗,請稍後再試」 |
| 總額 | 單價/數量變動 | 即時重算(calcTotalAmount,四捨五入至 2 位) |

驗證規則(shared `validateCreateOrder`):type 必選、asset/fiat 必選、price > 0、quantity > 0、paymentMethod 必選。

### Responsive
| 斷點 | 變化 |
|------|------|
| Desktop / Tablet | 卡片寬度受 max-width 720 限制,幣種/法幣與單價/數量各為兩欄並排 |
| Mobile(<600) | 維持兩欄(欄位短);如需可改為單欄堆疊(目前未強制) |

### Edge Cases
- 空白單價/數量:顯示驗證錯誤,不送出。
- 極大數值:總額以 `toLocaleString()` 顯示千分位。
- 慢速連線:送出期間按鈕 disabled,避免重複送出。

### Accessibility
- 每個輸入皆有 label(MUI label)。
- 買/賣 ToggleButton 具 selected 狀態。
- 錯誤訊息透過 helperText 呈現於對應欄位下方。

---

## 4. Web — 我的掛單(`/my-orders`)

### Overview
表格顯示 demo_user 自己的掛單;open 可取消。對應 `screens/MyOrdersScreen`。

### Components
表格欄:類型、幣種、單價、數量、總額、付款方式、狀態、建立時間、動作。
- 類型 → TypeTag;狀態 → StatusTag。
- 動作:當 `canCancel(status)`(僅 open)顯示「取消掛單」(danger 文字按鈕)。

### States and Interactions
| 元素 | 狀態 | 行為 |
|------|------|------|
| 取消按鈕 | 點擊 | 開啟確認對話框「確定要取消此掛單嗎?」 |
| 對話框 | 確認 | 呼叫 cancel → 重新載入 → 綠色 Snackbar「掛單已取消」 |
| 對話框 | 關閉 | 不動作 |
| 列表 | 載入失敗 | 紅色 Snackbar「載入失敗,請稍後再試」 |

### Edge Cases
- 空狀態:整列置中顯示「尚無掛單」(textTertiary,padding 32)。
- 僅 open 顯示取消;completed/cancelled 不顯示動作。

### Accessibility
- 表頭使用 `<th>`;動作按鈕有明確文案。
- 對話框聚焦管理由 MUI Dialog 處理。

---

## 5. Web — 後台訂單列表(`/admin`)

### Overview
讀後端全部訂單,可依狀態篩選,點列進詳情。沿用既有側邊欄(160px)+ Header(56px)版型。對應 `screens/AdminOrdersScreen`。

### Components
- Tabs 篩選:全部 / 待成交 / 已完成 / 已取消(對應 `all`、ORDER_STATUSES)。
- 表格欄:訂單編號、類型、幣種、單價、數量、總額、狀態、建立者、建立時間。
- 列 hover 可點,導向 `/admin/orders/:id`。

### States and Interactions
| 元素 | 狀態 | 行為 |
|------|------|------|
| Tab | 切換 | 以該狀態重新查詢(all 時不帶 status) |
| 列 | hover | MUI hover 高亮;游標 pointer |
| 列 | 點擊 | 導向詳情 |
| 列表 | 載入失敗 | 紅色 Snackbar |

### Edge Cases
- 空狀態:「尚無訂單」。
- 建立者欄顯示 username(seed 他人訂單與 demo_user 明顯區別)。

### Responsive
後台為桌面版型;表格於窄螢幕水平捲動(TableContainer)。

### Accessibility
- 側邊欄項目為可聚焦點擊區;狀態與使用者端一致(同一套 StatusTag)。

---

## 6. Web — 訂單詳情(`/admin/orders/:id`)

### Overview
顯示單筆完整資訊與正確狀態;open 可標記完成。對應 `screens/AdminOrderDetailScreen`。

### Components
- 標題列 + 「返回」按鈕(回 `/admin`)。
- 資訊卡:訂單編號、類型(TypeTag)、狀態(StatusTag)、幣種、法幣、單價、數量、總額、付款方式、建立者、建立時間,逐列以分隔線呈現。
- 「標記完成」按鈕:僅 `canComplete(status)`(open)顯示。

### States and Interactions
| 元素 | 狀態 | 行為 |
|------|------|------|
| 標記完成 | 點擊 | 確認對話框「確定要將此訂單標記為已完成嗎?」 |
| 對話框 | 確認 | 呼叫 complete → 重新載入 → 綠色 Snackbar「訂單已標記完成」 |
| 詳情 | 載入失敗 | 紅色 Snackbar |

### Edge Cases
- 非 open(completed/cancelled):不顯示「標記完成」。
- 找不到 id:後端回 404,前端顯示載入失敗訊息。

### Accessibility
- 確認動作文字按鈕;返回按鈕有明確文案。

---

## 7. App — 掛單(底部 tab「掛單」)

### Overview
React Native,使用者填單建立掛單;成功後切到「我的掛單」分頁。對應 `navigation/screens/V1CreateOrderScreen`。包在 `SafeAreaView` + `ScrollView`。

### Components / 欄位
| 元件 | 規格 |
|------|------|
| 買/賣 切換 | 兩個 TouchableOpacity,選中填 `typeColor`、白字 |
| 幣種 / 法幣 | 唯讀框(v1 單一值,下拉結構保留) |
| 單價 / 數量 | TextInput,keyboardType=decimal-pad;error 時框色 danger |
| 付款方式 | 兩個 TouchableOpacity,選中 `primaryDisabled` 底 + `primary` 框 |
| 總額 | 卡片列即時顯示 |
| 送出 | 主按鈕(高 44),submitting 時 `primaryDisabled` |

### States and Interactions
| 元素 | 狀態 | 行為 |
|------|------|------|
| 送出 | 驗證失敗 | 對應欄位下顯示錯誤(shared 同一份規則) |
| 送出 | 成功 | `Alert` 顯示「掛單建立成功」,清空單價/數量,navigate('MyOrders') |
| 送出 | 失敗 | `Alert` 顯示「送出失敗,請稍後再試」 |

### Edge Cases
- 鍵盤:數值鍵盤;ScrollView 確保小螢幕可捲動。
- Android 連本機後端需用 `10.0.2.2:8888`(見 RUNBOOK)。

### Accessibility
- 切換與付款方式按鈕設 `accessibilityRole="button"` 與 `accessibilityState.selected`。
- 送出按鈕 `accessibilityRole="button"`。

---

## 8. App — 我的掛單(底部 tab「我的掛單」)

### Overview
卡片式列表(FlatList)顯示自己訂單與狀態;open 可取消;下拉重新整理。對應 `navigation/screens/V1MyOrdersScreen`。

### Components
- 卡片:頂部 TypeTag + StatusTag;逐列顯示單價、數量、總額(粗體)、付款方式、建立時間。
- open 時顯示「取消掛單」按鈕(danger 外框)。

### States and Interactions
| 元素 | 狀態 | 行為 |
|------|------|------|
| 分頁 | 取得焦點 | `useFocusEffect` 重新載入(確保剛建立的訂單即時出現) |
| 下拉 | 釋放 | RefreshControl 重新載入 |
| 取消 | 點擊 | `Alert` 確認(取消/確認);確認後呼叫 cancel → 重新載入 |
| 載入失敗 | — | `Alert`「載入失敗,請稍後再試」 |

### Edge Cases
- 空狀態:置中「尚無掛單」(textTertiary)。
- 僅 open 顯示取消。

### Accessibility
- `keyExtractor` 使用 order.id;取消按鈕 `accessibilityRole="button"`。

---

## 9. 動效 / Motion

| 元素 | 觸發 | 動畫 | 時長 |
|------|------|------|------|
| Snackbar(Web) | 成功/失敗 | 淡入,2000ms 後自動關閉 | MUI 預設 |
| Dialog(Web) | 開關 | MUI 預設淡入/縮放 | MUI 預設 |
| Tab 切換(App) | 點擊 | React Navigation 預設 | 預設 |
| 導向我的掛單 | 建立成功 | Web 600ms 後導航;App 即時 navigate | — |

v1 不額外引入自訂動畫(KISS)。

---

## 10. 共同無障礙原則

- 對比:主按鈕一律白字於 `primary` 底;狀態色僅作標示並輔以文字,不以顏色為唯一資訊。
- 表單:所有輸入有 label;錯誤訊息與欄位關聯。
- 觸控目標:App 互動元件高度 ≥ 40,主按鈕 44。
- 文案:全數走 i18n,zh-TW;不寫死字串。

---

## 11. 實作對應檔案

| 畫面 | 檔案 |
|------|------|
| Web 掛單 | `frontend/web/src/screens/CreateOrderScreen/index.jsx` |
| Web 我的掛單 | `frontend/web/src/screens/MyOrdersScreen/index.jsx` |
| Web 後台列表 | `frontend/web/src/screens/AdminOrdersScreen/index.jsx` |
| Web 詳情 | `frontend/web/src/screens/AdminOrderDetailScreen/index.jsx` |
| Web 標籤元件 | `frontend/web/src/components/OrderTags.jsx` |
| App 掛單 | `frontend/app/src/navigation/screens/V1CreateOrderScreen/index.tsx` |
| App 我的掛單 | `frontend/app/src/navigation/screens/V1MyOrdersScreen/index.tsx` |
| Web 設計 token | `frontend/web/src/theme/index.js` |
| App 設計 token | `frontend/app/src/theme/index.ts` |
| 共用 驗證 / 狀態機 / i18n / API(不含 token) | `shared/src/**` |
