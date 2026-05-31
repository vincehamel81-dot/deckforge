import { describe, it, expect } from 'vitest'
import { getMergedNamespace } from '../localeService'

// All locales and namespaces that the app supports.
// If either list grows, the tests below automatically cover the new entries.
const NON_EN_LOCALES = ['fr-CA', 'es-MX'] as const
const NAMESPACES = ['common', 'auth', 'lobby', 'table', 'errors'] as const

/** Returns every dot-notation leaf path in a nested object.
 *  { shoe: { title: '…', refreshShoe: '↻' } } → ['shoe.title', 'shoe.refreshShoe'] */
function leafKeys(obj: Record<string, unknown>, prefix = ''): string[] {
  return Object.entries(obj).flatMap(([key, val]) => {
    const path = prefix ? `${prefix}.${key}` : key
    return val && typeof val === 'object' && !Array.isArray(val)
      ? leafKeys(val as Record<string, unknown>, path)
      : [path]
  })
}

function getByPath(obj: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>(
    (acc, key) => (acc && typeof acc === 'object' ? (acc as Record<string, unknown>)[key] : undefined),
    obj,
  )
}

// Core cascade invariant — parametric across every locale × every namespace.
// A single failing assertion prints the full path (e.g. "fr-CA/table/shoe.refreshShoeTitle")
// so the exact missing key is immediately visible.
//
// This test catches:
//   • cascade logic breaking (key becomes undefined)
//   • a new en-US key added without the merge propagating it
//   • a locale file accidentally setting a key to "" or null
//
// It does NOT care whether the value came from en-US fallback or a locale
// override — only that it is a non-empty string, which is always correct.
describe('cascade completeness — every locale × namespace', () => {
  for (const ns of NAMESPACES) {
    const enUS = getMergedNamespace('en-US', ns)
    const keys = leafKeys(enUS)

    for (const locale of NON_EN_LOCALES) {
      it(`${locale}/${ns} — all ${keys.length} leaf keys resolve to non-empty strings`, () => {
        const merged = getMergedNamespace(locale, ns)
        for (const path of keys) {
          const value = getByPath(merged as Record<string, unknown>, path)
          expect(value, `${locale}/${ns}/${path}`).toBeTypeOf('string')
          expect(
            (value as string).length,
            `${locale}/${ns}/${path} must not be empty`,
          ).toBeGreaterThan(0)
        }
      })
    }
  }
})

// Spot-check that locale overrides actually apply (not just en-US everywhere).
describe('locale override correctness', () => {
  it('fr-CA applies its own translations rather than falling through to en-US', () => {
    const frCA = getMergedNamespace('fr-CA', 'table')
    const enUS = getMergedNamespace('en-US', 'table')
    const frShoe = frCA.shoe as Record<string, unknown>
    const enShoe = enUS.shoe as Record<string, unknown>
    expect(frShoe.title).toBe('🃏 État du Sabot')   // fr-CA override
    expect(frShoe.title).not.toBe(enShoe.title)      // distinct from en-US value
  })
})
