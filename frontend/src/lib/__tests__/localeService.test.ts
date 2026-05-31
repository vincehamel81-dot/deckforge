import { describe, it, expect } from 'vitest'
import { getMergedNamespace } from '../localeService'

// These tests verify the pre-merge cascade strategy from ADR-002:
// each locale is deep-merged over the en-US base so t() calls are
// always O(1) lookups with zero runtime fallback logic.

describe('getMergedNamespace — cascade', () => {
  it('en-US returns the canonical namespace intact', () => {
    const result = getMergedNamespace('en-US', 'table')
    const shoe = result.shoe as Record<string, unknown>
    expect(shoe.refreshShoeTitle).toBeTypeOf('string')
    expect((shoe.refreshShoeTitle as string).length).toBeGreaterThan(0)
  })

  it('sparse locale never produces undefined for a key that exists in en-US', () => {
    // Tests the cascade invariant, not the current translation state.
    // Whether es-MX has its own refreshShoeTitle or inherits from en-US,
    // the merged result must be a non-empty string — never undefined.
    const esMX = getMergedNamespace('es-MX', 'table')
    const esShoe = esMX.shoe as Record<string, unknown>
    expect(esShoe.refreshShoeTitle).toBeTypeOf('string')
    expect((esShoe.refreshShoeTitle as string).length).toBeGreaterThan(0)
  })

  it('locale override replaces en-US value for translated keys', () => {
    const frCA = getMergedNamespace('fr-CA', 'table')
    const enUS = getMergedNamespace('en-US', 'table')
    const frShoe = frCA.shoe as Record<string, unknown>
    const enShoe = enUS.shoe as Record<string, unknown>
    // fr-CA has its own shoe.title — must differ from en-US
    expect(frShoe.title).not.toBe(enShoe.title)
    expect(frShoe.title).toBe('🃏 État du Sabot')
  })

  it('merge preserves all en-US keys in the merged locale', () => {
    const enUS = getMergedNamespace('en-US', 'table')
    const frCA = getMergedNamespace('fr-CA', 'table')
    // every top-level key in en-US must exist in fr-CA
    for (const key of Object.keys(enUS)) {
      expect(frCA).toHaveProperty(key)
    }
  })
})
