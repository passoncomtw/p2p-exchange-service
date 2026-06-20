// App 端 i18n 初始化（react-i18next）。
// v1 僅提供 zh-TW；語系訊息來自 shared 共用核心。
// 未來可改用 expo-localization 讀取裝置語系，目前固定 zh-TW。
import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { zhTW, DEFAULT_LOCALE } from '@shared';

i18n.use(initReactI18next).init({
  resources: {
    'zh-TW': { translation: zhTW },
  },
  lng: DEFAULT_LOCALE,
  fallbackLng: 'zh-TW',
  interpolation: {
    escapeValue: false,
  },
});

export default i18n;
