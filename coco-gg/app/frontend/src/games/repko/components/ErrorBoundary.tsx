import { Component, type ErrorInfo, type ReactNode } from 'react';

// React error boundaries are still class-only as of React 19; there is no hook
// equivalent. This is the one legitimate place in the Repko codebase for a
// class component.

interface Props {
  children: ReactNode;
  label: string;
}

interface State {
  error: Error | null;
  info: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null, info: null };

  static getDerivedStateFromError(error: Error): Partial<State> {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo): void {
    console.error('[repko ErrorBoundary]', this.props.label, error, info);
    this.setState({ info });
  }

  render(): ReactNode {
    if (this.state.error !== null) {
      return (
        <div className="m-4 p-4 rounded border border-red-300 bg-red-50 text-xs text-red-900 font-mono whitespace-pre-wrap">
          <div className="font-semibold mb-1">[{this.props.label}] crashed</div>
          <div className="mb-2">{this.state.error.message}</div>
          {this.state.info?.componentStack !== undefined && this.state.info.componentStack !== null && (
            <details>
              <summary className="cursor-pointer">component stack</summary>
              <pre className="mt-1 text-[10px] overflow-auto">{this.state.info.componentStack}</pre>
            </details>
          )}
        </div>
      );
    }
    return this.props.children;
  }
}
