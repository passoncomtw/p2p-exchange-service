---
title: 丰盈钱包 Design System
version: 0.1.0
last_updated: 2026-06-04
source: frontend/web/src code extraction + 後台 Figma design (node-id 0-185)
platforms: [web-admin, mobile-app]
note: >
  Mobile app tokens adopt the same brand palette.
  React Native uses dp (density-independent pixels), not px.
  Web admin uses MUI v9 with @emotion/styled.
---

## 1. Brand

**Product name:** 丰盈钱包 (Fēngyíng Qiánbāo)
**Sub-brand (web admin):** 运营后台 / 後台管理系統
**Logo:** Custom SVG owl — two circular eyes on a yellow rounded-square background.
**Logo variants:**
- `logo.large`: 48×48dp, used in LoginScreen
- `logo.small`: 32×32dp, used in Sidebar header

**Brand voice:** Trustworthy, financial, neutral. Never playful in error states.

**i18n:** Default `zh-TW` (繁體中文). Second supported locale: `zh-CN` (简体中文).
Language selection persists to `localStorage` key `lang`.


## 2. Colors

All hex values are exact — do not approximate or substitute.

### Primary / Brand

| Token | Value | Usage |
|-------|-------|-------|
| `colors.primary` | `#FFC107` | CTA buttons, logo background, focus ring, sidebar tagline |
| `colors.primary-dark` | `#FFB300` | Button hover state |
| `colors.primary-disabled` | `#FFE082` | Button disabled state |
| `colors.primary-deep` | `#FF8F00` | Logo beak accent |

### Sidebar (Dark Theme)

| Token | Value | Usage |
|-------|-------|-------|
| `colors.sidebar-bg` | `#2C2C3C` | Sidebar background |
| `colors.sidebar-hover` | `#3A3A4E` | Nav item hover |
| `colors.sidebar-active` | `#1A1A28` | Nav item selected/active |
| `colors.sidebar-text` | `#CCCCCC` | Nav text (inactive) |
| `colors.sidebar-text-active` | `#FFFFFF` | Nav text (active) |

### Page Backgrounds

| Token | Value | Usage |
|-------|-------|-------|
| `colors.bg-login` | `#EBEDF2` | Login page full-screen background |
| `colors.bg-content` | `#F5F5F7` | Main content area background |
| `colors.bg-card` | `#FFFFFF` | Card / modal / input background |

### Neutral Text

| Token | Value | Usage |
|-------|-------|-------|
| `colors.text-primary` | `#333333` | Body text, input values, headings |
| `colors.text-secondary` | `#666666` | Labels, descriptions |
| `colors.text-tertiary` | `#999999` | Hints, icon colors |
| `colors.text-placeholder` | `#BFBFBF` | Input placeholder, password visibility icon |
| `colors.text-avatar` | `#9E9E9E` | Avatar default background |

### Borders & Dividers

| Token | Value | Usage |
|-------|-------|-------|
| `colors.border-input` | `#D9D9D9` | Form input default border |
| `colors.border-card` | `#EBEBEB` | Card border, header divider |
| `colors.border-select` | `#E0E0E0` | Header language select border |
| `colors.border-sidebar-logo` | `rgba(255,255,255,0.06)` | Sidebar logo section bottom border |

### Semantic

| Token | Value | Usage |
|-------|-------|-------|
| `colors.danger` | `#F44336` | Destructive actions (logout, error text) |
| `colors.status-active` | `#4CAF50` | Order/member active status badge |
| `colors.status-frozen` | `#FF9800` | Order frozen/timeout status badge |
| `colors.status-stopped` | `#9E9E9E` | Disabled status badge |


## 3. Typography

Font family: `system-ui, -apple-system, sans-serif` (web); `System` default (React Native).
Do not import custom web fonts unless explicitly approved.

### Scale

| Token | Size | Weight | Usage |
|-------|------|--------|-------|
| `typography.brand` | 18px | 700 | Login page brand name |
| `typography.page-title` | 16px | 600 | Page heading (Dashboard, list pages) |
| `typography.nav-item` | 14px | 400/600 | Sidebar top-level nav (600 when active) |
| `typography.body` | 13px | 400 | Input values, header username, sub-nav |
| `typography.label` | 12px | 400 | Form field labels, language select |
| `typography.caption` | 11px | 400 | Login subtitle, version string |
| `typography.micro` | 10px | 400 | Sidebar tagline |
| `typography.button` | 14px | 500 | CTA button text |

### Icon Sizes

| Context | Size |
|---------|------|
| Sidebar nav icons | 18px |
| Header / menu icons | 16px |
| Spinner (loading) | 18px |


## 4. Spacing

Base unit: 4px. All spacing values are multiples of 4.

### Named Scale

| Token | Value | Usage |
|-------|-------|-------|
| `spacing.xs` | 4px | Label margin-bottom, scrollbar width |
| `spacing.sm` | 8px | Input vertical padding, header gap, button margin-top |
| `spacing.md` | 12px | Input horizontal padding |
| `spacing.lg` | 16px | Sidebar logo px, FormField gap, section gaps |
| `spacing.xl` | 20px | Sidebar nav item horizontal padding |
| `spacing.2xl` | 24px | Content area padding, header horizontal padding |
| `spacing.3xl` | 28px | Login card horizontal padding |
| `spacing.4xl` | 32px | Login card top padding |
| `spacing.sidebar-sub-indent` | 44px | Sub-nav item left padding (depth=1) |

### Component Dimensions

| Token | Value |
|-------|-------|
| `size.sidebar-width` | 160px |
| `size.header-height` | 56px |
| `size.sidebar-logo-height` | 56px |
| `size.nav-item-height` | 44px |
| `size.button-height` | 36px |
| `size.login-card-width` | 344px |
| `size.avatar` | 28px |
| `size.logo-large` | 48px |
| `size.logo-small` | 32px |
| `size.dropdown-min-width` | 140px |


## 5. Layout

### Web Admin

```
┌─────────────────────────────────────────────┐
│ Sidebar (160px fixed)  │ Header (56px)       │
│  Logo + nav            ├─────────────────────│
│                        │ Content (flex 1)    │
│                        │ padding: 24px       │
│                        │ bg: #F5F5F7         │
└────────────────────────┴─────────────────────┘
```

- Sidebar: `position: fixed`, `height: 100vh`, `z-index: 100`
- Content: `margin-left: 160px`, scrollable independently
- Header: full-width white bar, right-aligned user info + language select

### Mobile App (Expo / React Native)

- Navigation: Bottom Tab Navigator (Home, Trade, Orders, Profile)
- Screen width: device full width
- Safe area: always wrap in `SafeAreaView`
- Status bar: `StatusBar` with `barStyle: dark-content` on light screens
- List screens: `FlatList` with `keyExtractor`


## 6. Components

### Button — Primary CTA

```
bgcolor:        colors.primary (#FFC107)
color:          white
height:         36px
fontSize:       typography.button (14px/500)
borderRadius:   4px
boxShadow:      none
padding:        horizontal auto, fullWidth
:hover          bgcolor: colors.primary-dark (#FFB300)
:disabled       bgcolor: colors.primary-disabled (#FFE082), color: white
:loading        show CircularProgress size=18 color=white, hide label
```

### Input — Form Field

Custom (not MUI OutlinedInput floating label). Static label above, input below.

```
label:          typography.label (12px/400), color: colors.text-secondary
container:      border: 1px solid colors.border-input, borderRadius: 4px,
                px: spacing.md (12px), py: spacing.sm (8px), bgcolor: white
:focus-within   border-color: colors.primary (#FFC107)
input text:     typography.body (13px), color: colors.text-primary
placeholder:    color: colors.text-placeholder (#BFBFBF)
margin-bottom:  spacing.lg (16px) between fields
```

### Card — Login

```
width:          size.login-card-width (344px)
bgcolor:        colors.bg-card (white)
borderRadius:   4px
padding:        32px 28px 20px
boxShadow:      0 2px 8px rgba(0,0,0,0.08)
```

### Sidebar

```
width:          size.sidebar-width (160px)
bgcolor:        colors.sidebar-bg (#2C2C3C)
position:       fixed, left: 0, top: 0
height:         100vh
overflow-y:     auto (custom scrollbar: width 4px, thumb #444)
```

### Sidebar — Nav Item (Depth 0)

```
height:         44px
px:             20px
gap:            10px (icon + label)
fontSize:       14px / fontWeight: 400
color:          colors.sidebar-text (#CCCCCC)
:hover          bgcolor: colors.sidebar-hover (#3A3A4E)
.active         bgcolor: colors.sidebar-active (#1A1A28),
                color: colors.sidebar-text-active (#FFF),
                fontWeight: 600
```

### Sidebar — Nav Item (Depth 1, sub-menu)

```
px-left:        44px (simulates indent)
fontSize:       13px
inherits active/hover from parent rule
```

### Header

```
height:         56px
bgcolor:        white
border-bottom:  1px solid colors.border-card (#EBEBEB)
px:             24px
layout:         flex, align-items: center, justify-content: flex-end, gap: 16px
```

### Avatar

```
size:           28×28px
fontSize:       13px / fontWeight: 600
bgcolor:        colors.text-avatar (#9E9E9E)
content:        first character of username, uppercase
```

### Status Badge (Order / Member)

```
active:         color colors.status-active, small text label
frozen/timeout: color colors.status-frozen
stopped:        color colors.status-stopped
fontSize:       12px
```

### Language Select (Header + Login)

```
fontSize:       12px
border:         colors.border-select or colors.border-input
py:             5–6px, px: 10px
options:        繁體中文 (zh-TW), 简体中文 (zh-CN)
```


## 7. Do's and Don'ts

### Colors

- **DO** use `colors.primary` (#FFC107) for the single primary CTA on each screen.
- **DO** use `colors.danger` (#F44336) only for destructive/irreversible actions (logout, delete, cancel order).
- **DON'T** use the primary yellow for informational text — it fails WCAG contrast on white.
- **DON'T** invent new grays. Use the defined neutral scale.

### Typography

- **DO** use `typography.page-title` (16px/600) for every page's `<h1>` equivalent.
- **DO** use `typography.body` (13px) for table cell content and sidebar sub-items.
- **DON'T** go below 10px for any visible text.
- **DON'T** use bold (700) outside brand name and logo context.

### Spacing

- **DO** use multiples of 4px for all spacing.
- **DON'T** add random padding values (e.g., 7px, 15px, 22px).

### Sidebar

- **DO** keep sidebar fixed at 160px — do not make it collapsible without design approval.
- **DON'T** add more than 2 levels of nav depth.

### Buttons

- **DO** use `fullWidth` for login/auth page CTAs.
- **DON'T** place two primary (#FFC107) buttons side by side on the same screen.

### i18n

- **DO** always use `t('key')` — never hardcode Chinese strings in JSX.
- **DON'T** fall back to zh-CN strings in zh-TW context or vice versa.

### React Native (App)

- **DO** use `StyleSheet.create()` for all RN styles.
- **DO** use `dp` equivalents — 1dp ≈ 1px in the design tokens above.
- **DON'T** use `px` units in RN StyleSheet (RN ignores unit suffixes).
- **DON'T** use `position: fixed` in RN — use navigation patterns instead.


## 8. Accessibility

- All interactive elements must have a minimum touch target of **44×44dp** (nav items already comply).
- Color contrast:
  - Primary button (`#FFC107` on white): fails WCAG AA for text — use white text on button (#FFC107 bg passes for large text ≥18px).
  - Sidebar text (`#CCCCCC` on `#2C2C3C`): contrast ratio ≈ 5.4:1 — passes AA.
  - Active sidebar text (`#FFF` on `#1A1A28`): contrast ratio ≈ 15:1 — passes AAA.
- Screen readers: all SVG icons in Sidebar use `aria-hidden`; interactive buttons require `aria-label` when icon-only.
- Language switcher: announce language change via `aria-live` region.
- Error states: form errors must use `role="alert"` or `aria-describedby` linkage.
- Loading states: spinner must have `aria-label="載入中"` / `aria-label="加载中"` per locale.


## 9. Agent Prompt Guide

When you (Claude / any AI agent) are working on UI in this project, follow these rules:

### Before writing any UI code

1. Read this `DESIGN.md` in full.
2. Identify which token categories the task touches (button? input? layout? color?).
3. Grep the existing code to confirm the token value is already in use — do not introduce new values.

### Token usage rules

- **Colors**: Only use hex values from Section 2. If a color you need is not listed, add it to Section 2 first and get user approval.
- **Spacing**: All spacing must be a multiple of 4px (see Section 4 named scale).
- **Typography**: All font sizes must match an entry in Section 3.

### Platform check

- If writing for `frontend/web/src/`: use MUI v9 `sx` prop for all styles. System props (non-`sx`) are forbidden on `Box`.
- If writing for `frontend/app/`: use `StyleSheet.create()`, no px units, wrap in `SafeAreaView`.

### When tokens are missing

If the design requires a value not in this file, do NOT invent one. Instead:
1. Log it under `## Open Token Questions` in `findings.md`.
2. Ask the user: "DESIGN.md does not define `<token>`. What value should I use?"

### i18n check

Every visible string in UI must be in `src/i18n/locales/zh-TW.js` and `src/i18n/locales/zh-CN.js` before the component is written. Use `useTranslation()` hook.

### Order of UI implementation

1. Read relevant Figma node (if accessible) or existing component for reference.
2. Check `findings.md` for any active `## Design Context` block.
3. Map the design to tokens from this file.
4. Write the component. Flag any deviation in a comment: `// DESIGN.md deviation: <reason>`.
