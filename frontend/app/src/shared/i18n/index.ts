// i18n 共用匯出：v1 僅 zh-TW，架構保留多語系擴充。
import { zhTW } from './zh-TW';

export const SUPPORTED_LOCALES = ['zh-TW'] as const;
export type Locale = (typeof SUPPORTED_LOCALES)[number];
export const DEFAULT_LOCALE: Locale = 'zh-TW';

export const messages: Record<Locale, typeof zhTW> = {
  'zh-TW': zhTW,
};

export { zhTW };
