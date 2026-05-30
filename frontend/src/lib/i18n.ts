import i18n from 'i18next'
import type { Resource } from 'i18next'
import { initReactI18next } from 'react-i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import { getMergedNamespace } from './localeService'

export const SUPPORTED_LOCALES = ['en-US', 'fr-CA', 'es-MX'] as const
export const NAMESPACES = ['common', 'auth', 'lobby', 'table', 'errors'] as const
export type Locale = typeof SUPPORTED_LOCALES[number]
export type Namespace = typeof NAMESPACES[number]

// Build the complete resources object synchronously at module load time.
// All locale files are bundled by Vite; no async loading phase needed.
// See ASSUMPTIONS A-007 for the lazy-loading upgrade path.
const resources: Resource = {}
for (const locale of SUPPORTED_LOCALES) {
  resources[locale] = {}
  for (const ns of NAMESPACES) {
    resources[locale][ns] = getMergedNamespace(locale, ns)
  }
}

i18n
  .use(LanguageDetector)
  .use(initReactI18next)
  .init({
    resources,
    fallbackLng: 'en-US',
    supportedLngs: [...SUPPORTED_LOCALES],
    defaultNS: 'common',
    ns: [...NAMESPACES],
    detection: {
      // Language preference stored in localStorage so it persists across sessions (ASSUMPTIONS A-005).
      order: ['localStorage'],
      caches: ['localStorage'],
      lookupLocalStorage: 'deckforge-locale',
    },
    interpolation: { escapeValue: false },
  })

export default i18n
