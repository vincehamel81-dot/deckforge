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

export function CardBadge({ suit, face, isNew }: { suit: string; face: string; isNew?: boolean }) {
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
      background: isNew ? '#1a2a4a' : '#0f1a2e',
      border: isNew ? '1px solid #e2c97e' : '1px solid #2d4a6a',
      boxShadow: isNew ? '0 0 6px #e2c97e55' : 'none',
      borderRadius: '6px', padding: '0.3rem 0.5rem', margin: '0.2rem',
      fontSize: '0.9rem', fontWeight: 700,
      color: SUIT_COLOR[suit] ?? '#e2e8f0', minWidth: '2.8rem',
      transition: 'box-shadow 0.3s ease',
    }}>
      {FACE_LABEL[face] ?? face}{SUIT_SYMBOL[suit]}
    </span>
  )
}
