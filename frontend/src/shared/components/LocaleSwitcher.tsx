import { useLocale } from '../hooks/useLocale'
import type { Locale } from '../../lib/i18n'

const FLAG: Record<Locale, string> = {
  'en-US': '🇺🇸',
  'fr-CA': '🇫🇷',
  'es-MX': '🇲🇽',
}

export function LocaleSwitcher() {
  const { locale, setLocale, supportedLocales } = useLocale()
  return (
    <div style={{ display: 'flex', gap: '0.15rem', alignItems: 'center' }}>
      {supportedLocales.map(loc => (
        <button
          key={loc}
          onClick={() => setLocale(loc)}
          title={loc}
          style={{
            background: 'none',
            border: locale === loc ? '1px solid #e2c97e' : '1px solid transparent',
            borderRadius: '4px',
            cursor: 'pointer',
            padding: '0.1rem 0.2rem',
            fontSize: '1rem',
            opacity: locale === loc ? 1 : 0.45,
            lineHeight: 1,
          }}
        >
          {FLAG[loc]}
        </button>
      ))}
    </div>
  )
}
