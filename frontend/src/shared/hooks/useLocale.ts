import { useTranslation } from 'react-i18next'
import { SUPPORTED_LOCALES } from '../../lib/i18n'
import type { Locale } from '../../lib/i18n'

export function useLocale() {
  const { i18n } = useTranslation()
  return {
    locale: i18n.language as Locale,
    setLocale: (locale: Locale) => { void i18n.changeLanguage(locale) },
    supportedLocales: SUPPORTED_LOCALES,
  }
}
