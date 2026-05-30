import usFlag from 'flag-icons/flags/4x3/us.svg'
import frFlag from 'flag-icons/flags/4x3/fr.svg'
import mxFlag from 'flag-icons/flags/4x3/mx.svg'
import { useLocale } from '../hooks/useLocale'
import type { Locale } from '../../lib/i18n'

// fr-CA uses the French flag (🇫🇷) by project convention — see ASSUMPTIONS A-008.
const FLAG_SRC: Record<Locale, string> = {
  'en-US': usFlag,
  'fr-CA': frFlag,
  'es-MX': mxFlag,
}

export function LocaleSwitcher() {
  const { locale, setLocale, supportedLocales } = useLocale()
  return (
    <div style={{ display: 'flex', gap: '0.2rem', alignItems: 'center' }}>
      {supportedLocales.map(loc => (
        <button
          key={loc}
          onClick={() => setLocale(loc)}
          title={loc}
          style={{
            background: 'none',
            border: locale === loc ? '1px solid #e2c97e' : '1px solid transparent',
            borderRadius: '3px',
            cursor: 'pointer',
            padding: '2px 3px',
            lineHeight: 0,
            opacity: locale === loc ? 1 : 0.45,
          }}
        >
          <img src={FLAG_SRC[loc]} alt={loc} width={20} height={15} style={{ display: 'block', borderRadius: '1px' }} />
        </button>
      ))}
    </div>
  )
}
