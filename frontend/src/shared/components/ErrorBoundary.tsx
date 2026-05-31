import { Component, type ErrorInfo, type ReactNode } from 'react'

interface Props { children: ReactNode }
interface State { error: Error | null }

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null }

  static getDerivedStateFromError(error: Error): State {
    return { error }
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('[ErrorBoundary]', error, info.componentStack)
  }

  render() {
    if (this.state.error) {
      return (
        <div style={{
          minHeight: '100vh', background: '#0f1a2e', color: '#e2e8f0',
          display: 'flex', flexDirection: 'column', alignItems: 'center',
          justifyContent: 'center', gap: '1rem', padding: '2rem', textAlign: 'center',
        }}>
          <div style={{ fontSize: '3rem' }}>⚠</div>
          <h2 style={{ color: '#f87171', margin: 0 }}>Something went wrong</h2>
          <p style={{ color: '#7a9bb5', maxWidth: '420px', margin: 0 }}>
            {this.state.error.message}
          </p>
          <button
            onClick={() => window.location.reload()}
            style={{
              padding: '0.6rem 1.5rem', background: '#e2c97e', color: '#0f1a2e',
              border: 'none', borderRadius: '8px', fontWeight: 700, cursor: 'pointer',
            }}
          >
            Reload
          </button>
        </div>
      )
    }
    return this.props.children
  }
}
