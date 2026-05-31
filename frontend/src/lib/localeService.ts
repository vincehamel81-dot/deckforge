import enCommon from '../locales/en-US/common.json'
import enAuth from '../locales/en-US/auth.json'
import enLobby from '../locales/en-US/lobby.json'
import enTable from '../locales/en-US/table.json'
import enErrors from '../locales/en-US/errors.json'

import frCommon from '../locales/fr-CA/common.json'
import frAuth from '../locales/fr-CA/auth.json'
import frLobby from '../locales/fr-CA/lobby.json'
import frTable from '../locales/fr-CA/table.json'
import frErrors from '../locales/fr-CA/errors.json'

import esMXCommon from '../locales/es-MX/common.json'
import esMXAuth from '../locales/es-MX/auth.json'
import esMXLobby from '../locales/es-MX/lobby.json'
import esMXTable from '../locales/es-MX/table.json'
import esMXErrors from '../locales/es-MX/errors.json'

type NestedRecord = Record<string, unknown>

const FALLBACK = 'en-US'
const SESSION_PREFIX = 'i18n:'
// In dev mode, Vite HMR can update JSON files without bumping the session key,
// producing stale merged namespaces. Skip the cache entirely in development so
// every page load uses freshly-merged locale data.
const USE_SESSION_CACHE = !import.meta.env.DEV

const SOURCE: Record<string, Record<string, NestedRecord>> = {
  'en-US': { common: enCommon, auth: enAuth, lobby: enLobby, table: enTable, errors: enErrors },
  'fr-CA': { common: frCommon, auth: frAuth, lobby: frLobby, table: frTable, errors: frErrors },
  'es-MX': { common: esMXCommon, auth: esMXAuth, lobby: esMXLobby, table: esMXTable, errors: esMXErrors },
}

// In-memory cache survives for the session without repeated JSON.parse calls.
const memoryCache = new Map<string, NestedRecord>()

function deepMerge(base: NestedRecord, override: NestedRecord): NestedRecord {
  const result = { ...base }
  for (const key of Object.keys(override)) {
    const ov = override[key]
    const bv = base[key]
    if (ov && typeof ov === 'object' && !Array.isArray(ov) && bv && typeof bv === 'object') {
      result[key] = deepMerge(bv as NestedRecord, ov as NestedRecord)
    } else {
      result[key] = ov
    }
  }
  return result
}

/**
 * Returns a pre-merged namespace object: target locale keys override the en-US base.
 * Result is cached in-memory and in sessionStorage so the merge runs once per locale+ns.
 * See ASSUMPTIONS A-005 for the long-term localStorage migration path.
 */
export function getMergedNamespace(locale: string, ns: string): NestedRecord {
  const cacheKey = `${locale}:${ns}`

  const mem = memoryCache.get(cacheKey)
  if (mem) return mem

  if (USE_SESSION_CACHE) {
    try {
      const stored = sessionStorage.getItem(SESSION_PREFIX + cacheKey)
      if (stored) {
        const parsed = JSON.parse(stored) as NestedRecord
        memoryCache.set(cacheKey, parsed)
        return parsed
      }
    } catch { /* sessionStorage unavailable in some environments */ }
  }

  const base = (SOURCE[FALLBACK]?.[ns] ?? {}) as NestedRecord
  const target = locale === FALLBACK ? {} : ((SOURCE[locale]?.[ns] ?? {}) as NestedRecord)
  const merged = deepMerge(base, target)

  memoryCache.set(cacheKey, merged)
  if (USE_SESSION_CACHE) {
    try {
      sessionStorage.setItem(SESSION_PREFIX + cacheKey, JSON.stringify(merged))
    } catch { /* storage full or unavailable */ }
  }

  return merged
}
