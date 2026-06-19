# DESIGN.md Critique 報告 — p2p-exchange-service

**審查日期：** 2026-06-19
**審查範圍：** `DESIGN.md` v0.1.0 vs `frontend/web/src/` 所有 JSX/JS/CSS 檔案
**發現數量：** 14 項

---

## Pass 1: Structural（結構性審查）

### F1 — P2（中）：缺少 MUI Theme 中央化設定，整個設計系統靠 sx prop 散落維護

**描述：** DESIGN.md 第 9 節（Agent Prompt Guide）明確指出應使用 MUI v9 sx prop，但全站沒有 `createTheme()` 設定檔。所有色彩、字型、間距皆以字串 hardcode 在各元件的 sx prop 中，沒有任何 theme token 中樞。

**證據：**
```
frontend/web/src/ 下無任何 theme.js / theme.ts
main.jsx 無 ThemeProvider wrapper
```
每個元件都各自寫 `bgcolor: '#FFC107'`、`fontSize: '14px'` — 若設計系統更新，需要全文搜尋替換，風險極高。

**建議修正：** 建立 `src/theme/index.js`，使用 `createTheme()` 定義 palette / typography，在 `main.jsx` 以 `<ThemeProvider>` 包裹，元件改用 `theme.palette.primary.main` 等語義 token。

---

### F2 — P3（低）：DESIGN.md 缺少 z-index 層級定義

**描述：** Section 5（Layout）說明 Sidebar `z-index: 100`，但文件沒有完整的 z-index stacking table。程式碼中 Sidebar 使用 `zIndex: 100`，Header、Modal、Dropdown 的 z-index 完全沒有規範，未來加入 Modal/Toast/Tooltip 時層級衝突風險高。

**建議修正：** DESIGN.md 新增 `## 10. Z-Index Scale`，定義 `sidebar: 100 / header: 50 / modal: 1000 / toast: 1200` 等層級。

---

## Pass 2: Drift（漂移審查）

### F3 — P1（高）：`index.css` 使用完全不屬於設計系統的色彩變數集

**描述：** `src/index.css` 定義了一整套 CSS 自定義變數（`--accent: #aa3bff`、`--text: #6b6375` 等），這些顏色與 DESIGN.md Section 2 完全沒有交集。這是 Vite 初始化的模板殘留，但它設定了全域 `color: var(--text)`，會影響任何沒有被 MUI sx 覆蓋的文字元素。

**證據：**
```css
/* src/index.css line 2-9 */
:root {
  --text: #6b6375;      /* 不存在於 DESIGN.md */
  --accent: #aa3bff;    /* 紫色，與品牌黃 #FFC107 完全無關 */
  --bg: #fff;
  ...
  color-scheme: light dark;  /* 觸發系統 dark mode！ */
}
```
**建議修正：** 清除 `index.css` 中的設計無關 CSS 變數，僅保留 `body { margin: 0 }` 和 reset；將 CSS 變數重新定義為 DESIGN.md 的 token 值。

---

### F4 — P1（高）：`index.css` 啟用 `color-scheme: light dark`，但 DESIGN.md 完全未定義 dark mode

**描述：** `src/index.css` 設定 `color-scheme: light dark`，且有完整的 `@media (prefers-color-scheme: dark)` 規則覆蓋背景與文字色。但 DESIGN.md 完全沒有 dark mode token，元件中所有 hardcode 色值（`#F5F5F7`、`#333333` 等）在 dark mode 下不會切換。

**證據：**
```css
/* index.css */
@media (prefers-color-scheme: dark) {
  :root {
    --bg: #16171d;
    --text-h: #f3f4f6;
    ...
  }
}
```
而 `MainLayout` 仍 hardcode `bgcolor: '#F5F5F7'`，MUI sx 會優先覆蓋 CSS 變數，造成 dark mode 下白底 + 白字的可讀性災難。

**建議修正：** 二選一：（A）從 `index.css` 移除 `color-scheme` 和 dark media query，在 DESIGN.md 標注「不支援 dark mode」；（B）正式設計 dark mode token 並全套實作。

---

### F5 — P2（中）：Login 頁版本文字使用了 `colors.text-placeholder`（#BFBFBF），但語意不正確

**描述：** `LoginScreen` 版本字串使用 `color: '#BFBFBF'`，而 DESIGN.md 定義 `#BFBFBF` 僅應用於 `colors.text-placeholder`（Input placeholder、password visibility icon）。版本字串在語意上應使用 `colors.text-tertiary`（#999999）。

**證據：**
```jsx
/* LoginScreen/index.jsx */
<Typography sx={{ fontSize: '11px', color: '#BFBFBF', mt: '16px' }}>
```
DESIGN.md Section 2: `colors.text-placeholder #BFBFBF — Input placeholder, password visibility icon`

**建議修正：** 改為 `color: '#999999'`（`colors.text-tertiary`）。

---

### F6 — P2（中）：Sidebar 品牌名稱 13px/700 無對應 token

**描述：** DESIGN.md Section 7（Do's and Don'ts）說明 bold(700) 可用於品牌名稱。但 Typography Scale（Section 3）中沒有為 sidebar brand label 定義對應 token，它既不符合 `typography.brand`（18px）也沒有其他 token（實際用了 13px/700）。

**證據：**
```jsx
/* Sidebar/index.jsx */
<Typography sx={{ fontSize: '13px', fontWeight: 700, color: '#FFF', lineHeight: 1.2 }}>
```
**建議修正：** DESIGN.md 新增 `typography.sidebar-brand: 13px / 700`，或確認此處應降為 `600`（`typography.nav-item` 的 active 字重）。

---

### F7 — P2（中）：DashboardScreen 中 `fontSize: '14px'` 用於一般內容文字，違反 body token

**描述：** DESIGN.md Section 3 明確規定 `typography.body = 13px`，14px 沒有對應 token（最接近的是 `typography.button = 14px/500`，但那僅用於按鈕）。

**證據：**
```jsx
/* DashboardScreen/index.jsx */
<Typography sx={{ fontSize: '14px', color: '#666' }}>
```
**建議修正：** 改為 `fontSize: '13px'` 符合 `typography.body`，或在 DESIGN.md 新增對應 token。

---

### F8 — P3（低）：語系選擇器的 `py` 值在 Header 與 LoginScreen 不一致

**描述：** DESIGN.md Section 6（Language Select）規定 `py: 5–6px`，程式碼中 Header 使用 `py: '5px'`，LoginScreen 使用 `py: '6px'`。允許範圍內但不一致，日後維護容易造成更多漂移。

**建議修正：** DESIGN.md 將 Language Select 的 py 明確定為單一值（建議 `6px`），並統一兩處程式碼。

---

## Pass 3: Accessibility（無障礙審查）

### F9 — P0（阻礙）：OwlLogo SVG 在兩個元件中均無 aria 屬性，且程式碼重複

**描述：** `LoginScreen` 和 `Sidebar` 各自實作了完全相同的 `OwlLogo` SVG 元件，但兩個版本均無 `aria-hidden="true"` 或 `role="img"` + `aria-label`。DESIGN.md Section 8 明確要求「所有 SVG icons 使用 `aria-hidden`」。Logo 作為裝飾性圖形，缺少 `aria-hidden` 會讓螢幕報讀器讀出 SVG 內部結構雜訊。

**證據：**
```jsx
/* LoginScreen/index.jsx & Sidebar/index.jsx — 各自重複實作 */
const OwlLogo = () => (
  <svg width="48" height="48" viewBox="0 0 48 48" fill="none" xmlns="...">
    {/* 無 aria-hidden="true" */}
```
**建議修正：**
1. 將 `OwlLogo` 抽取為共用元件（`src/components/OwlLogo.jsx`），消除重複
2. 加入 `aria-hidden="true"` 屬性

---

### F10 — P1（高）：密碼顯示切換按鈕（IconButton）缺少 aria-label

**描述：** `LoginScreen` 的 `IconButton` 僅有圖示（Visibility / VisibilityOff icon），沒有 `aria-label`，螢幕報讀器無法告知使用者這個按鈕的功能。DESIGN.md Section 8：「interactive buttons require `aria-label` when icon-only」。

**證據：**
```jsx
/* LoginScreen/index.jsx */
<IconButton
  size="small"
  onClick={() => setShowPassword((v) => !v)}
  sx={{ color: '#BFBFBF', p: '2px' }}
>
  {showPassword ? <VisibilityOff /> : <Visibility />}
</IconButton>
```
**建議修正：** 加入 `aria-label={showPassword ? t('a11y.hidePassword') : t('a11y.showPassword')}`，並在 i18n 檔案中新增對應 key。

---

### F11 — P1（高）：語系切換後無 aria-live 通知

**描述：** DESIGN.md Section 8：「Language switcher: announce language change via `aria-live` region」。目前整個應用無任何 `aria-live` region，語系切換對螢幕報讀器使用者完全靜默。

**建議修正：** 在 `MainLayout` 或 `App.jsx` 加入 `aria-live="polite"` region，語系切換時設定公告文字。

---

### F12 — P1（高）：登入表單錯誤訊息缺少無障礙關聯

**描述：** DESIGN.md Section 8：「form errors must use `role="alert"` or `aria-describedby` linkage」。`LoginScreen` 的 `FormHelperText` 沒有 `role="alert"`，錯誤出現時螢幕報讀器不會自動播報。

**證據：**
```jsx
/* LoginScreen/index.jsx */
{error && (
  <FormHelperText error sx={{ mb: '8px', fontSize: '12px' }}>
    {error}
  </FormHelperText>
)}
```
**建議修正：** 加入 `role="alert"` 屬性：`<FormHelperText error role="alert" ...>`

---

### F13 — P2（中）：Loadable 全局 loading spinner 缺少 aria-label

**描述：** `Loadable.jsx` 的 `CircularProgress` 沒有 `aria-label`。DESIGN.md Section 8：「spinner must have `aria-label` per locale」。路由 lazy-load 時螢幕報讀器使用者不知道頁面正在載入。

**建議修正：** `<CircularProgress aria-label="載入中" />`

---

## Pass 4 & 5: Completeness / Consistency（完整性 & 一致性）

### F14 — P2（中）：`index.css` 的 `#root` 設定 `width: 1126px` + `text-align: center`，與後台全寬 Layout 規格衝突

**描述：** `index.css` 對 `#root` 設定了固定寬度與 `text-align: center`，這是 Vite 模板殘留，與 DESIGN.md Section 5（Web Admin Layout）規定的「Sidebar fixed 160px + Content flex 1」全寬後台版型不符。`text-align: center` 會影響所有未明確設定的子元素。

**證據：**
```css
/* index.css */
#root {
  width: 1126px;
  max-width: 100%;
  margin: 0 auto;
  text-align: center;
}
```
**建議修正：** `#root` 規則簡化為：
```css
#root {
  min-height: 100vh;
}
```

---

## 總結

| 嚴重度 | 數量 | 主要問題 |
|---|---|---|
| P0（阻礙） | 1 | SVG aria-hidden 缺失（F9） |
| P1（高） | 4 | index.css 色彩/dark mode 污染（F3、F4），密碼按鈕無 aria-label（F10），語系切換無通知（F11），錯誤訊息無 alert role（F12） |
| P2（中） | 6 | 缺 MUI Theme 中樞（F1），版本字色彩語意錯誤（F5），Sidebar 字重無 token（F6），body 字型漂移（F7），Loadable spinner 無 aria-label（F13），#root 全寬衝突（F14） |
| P3（低） | 3 | 缺 z-index 定義（F2），語系選擇 py 不一致（F8） |

**最高優先修復路徑：**
1. 清除 `index.css` 中的 Vite 模板殘留（F3、F4、F14）— 一次修改解決 3 個發現
2. 建立 MUI ThemeProvider 中樞（F1）
3. 補齊 accessibility 屬性：`aria-hidden` on OwlLogo、`aria-label` on IconButton、`role="alert"` on error（F9、F10、F12）
