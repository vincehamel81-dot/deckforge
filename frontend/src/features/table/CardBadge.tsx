export const SUIT_SYMBOL: Record<string, string> = {
  HEARTS: '♥', SPADES: '♠', CLUBS: '♣', DIAMONDS: '♦',
}
export const SUIT_COLOR: Record<string, string> = {
  HEARTS: '#f87171', SPADES: '#e2e8f0', CLUBS: '#e2e8f0', DIAMONDS: '#f87171',
}
export const FACE_LABEL: Record<string, string> = {
  ACE: 'A', TWO: '2', THREE: '3', FOUR: '4', FIVE: '5', SIX: '6',
  SEVEN: '7', EIGHT: '8', NINE: '9', TEN: '10', JACK: 'J', QUEEN: 'Q', KING: 'K',
}

export function CardBadge({ suit, face }: { suit: string; face: string }) {
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
      background: '#0f1a2e', border: '1px solid #2d4a6a', borderRadius: '6px',
      padding: '0.3rem 0.5rem', margin: '0.2rem', fontSize: '0.9rem', fontWeight: 700,
      color: SUIT_COLOR[suit] ?? '#e2e8f0', minWidth: '2.8rem',
    }}>
      {FACE_LABEL[face] ?? face}{SUIT_SYMBOL[suit]}
    </span>
  )
}
