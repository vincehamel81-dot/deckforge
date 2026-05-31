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

  it('sparse locale inherits missing keys from en-US', () => {
    // es-MX does not define shoe.refreshShoeTitle — must come from en-US base
    const enUS = getMergedNamespace('en-US', 'table')
    const esMX = getMergedNamespace('es-MX', 'table')
    const enShoe = enUS.shoe as Record<string, unknown>
    const esShoe = esMX.shoe as Record<string, unknown>
    expect(esShoe.refreshShoeTitle).toBe(enShoe.refreshShoeTitle)
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
