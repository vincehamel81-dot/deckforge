import { describe, it, expect } from 'vitest'
import { SUIT_SYMBOL, SUIT_COLOR, FACE_LABEL } from '../CardBadge'

// A deck has exactly 4 suits and 13 faces.
// These maps drive every card rendered in the UI — an incomplete map
// silently produces undefined, which shows as an empty badge.

const SUITS = ['HEARTS', 'SPADES', 'CLUBS', 'DIAMONDS'] as const
const FACES = ['ACE', 'TWO', 'THREE', 'FOUR', 'FIVE', 'SIX', 'SEVEN',
               'EIGHT', 'NINE', 'TEN', 'JACK', 'QUEEN', 'KING'] as const

describe('SUIT_SYMBOL', () => {
  it('covers all 4 suits', () => {
    SUITS.forEach(s => expect(SUIT_SYMBOL[s]).toBeDefined())
  })

  it('contains no undefined gaps', () => {
    expect(Object.keys(SUIT_SYMBOL)).toHaveLength(4)
  })
})

describe('SUIT_COLOR', () => {
  it('covers all 4 suits with a valid hex colour', () => {
    SUITS.forEach(s => expect(SUIT_COLOR[s]).toMatch(/^#[0-9a-fA-F]{6}$/))
  })

  it('reds and blacks use distinct colours', () => {
    expect(SUIT_COLOR.HEARTS).toBe(SUIT_COLOR.DIAMONDS)   // both red
    expect(SUIT_COLOR.SPADES).toBe(SUIT_COLOR.CLUBS)       // both light
    expect(SUIT_COLOR.HEARTS).not.toBe(SUIT_COLOR.SPADES)  // different families
  })
})

describe('FACE_LABEL', () => {
  it('covers all 13 card faces', () => {
    FACES.forEach(f => expect(FACE_LABEL[f]).toBeDefined())
  })

  it('has exactly 13 entries — no extras', () => {
    expect(Object.keys(FACE_LABEL)).toHaveLength(13)
  })

  it('maps face names to their display labels', () => {
    expect(FACE_LABEL.ACE).toBe('A')
    expect(FACE_LABEL.KING).toBe('K')
    expect(FACE_LABEL.TEN).toBe('10')
  })
})
