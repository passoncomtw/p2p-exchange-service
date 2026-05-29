import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import zhTW from './locales/zh-TW'
import zhCN from './locales/zh-CN'

// 從 localStorage 讀取，預設繁體中文
const savedLang = localStorage.getItem('lang') || 'zh-TW'

i18n.use(initReactI18next).init({
  resources: {
    'zh-TW': { translation: zhTW },
    'zh-CN': { translation: zhCN },
  },
  lng: savedLang,
  fallbackLng: 'zh-TW',
  interpolation: {
    escapeValue: false,
  },
})

export const changeLanguage = (lang) => {
  i18n.changeLanguage(lang)
  localStorage.setItem('lang', lang)
}

export const SUPPORTED_LANGS = [
  { code: 'zh-TW', label: '繁體中文' },
  { code: 'zh-CN', label: '简体中文' },
]

export default i18n
