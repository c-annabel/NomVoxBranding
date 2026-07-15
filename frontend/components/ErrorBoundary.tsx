"use client";

import React from "react";

interface Props {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

/**
 * ErrorBoundary — wraps the app and catches unhandled React render errors.
 * Shows a NomVox-styled recovery screen instead of a blank white page.
 * 
 * Usage: wrap <HomeClient /> in <ErrorBoundary> in page.tsx or layout.tsx.
 * Class component is required because React error boundaries must be class-based.
 */
export default class ErrorBoundary extends React.Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, info: React.ErrorInfo) {
    // Log to console — visible in browser DevTools and Vercel function logs
    console.error("[NomVox ErrorBoundary]", error, info.componentStack);
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null });
  };

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) return this.props.fallback;

      return (
        <div
          className="min-h-screen flex items-center justify-center px-6"
          style={{ background: "var(--color-void)" }}
        >
          <div
            className="rounded-2xl border p-10 max-w-lg w-full text-center"
            style={{
              background: "rgba(18,22,42,0.97)",
              borderColor: "rgba(248,113,113,0.30)",
              boxShadow: "0 0 40px rgba(248,113,113,0.10)",
            }}
          >
            {/* Icon */}
            <div
              className="w-14 h-14 rounded-full flex items-center justify-center mx-auto mb-6 text-2xl"
              style={{ background: "rgba(248,113,113,0.14)", border: "1px solid rgba(248,113,113,0.30)" }}
            >
              ⚠
            </div>

            <p
              className="text-xs font-black uppercase tracking-widest mb-2"
              style={{ color: "#f87171" }}
            >
              NomVox encountered an error
            </p>
            <h2
              className="text-2xl font-black mb-3"
              style={{ color: "var(--color-text-primary)" }}
            >
              Something went wrong
            </h2>
            <p className="text-sm mb-2" style={{ color: "var(--color-text-secondary)" }}>
              The interface crashed unexpectedly. Your session data is not lost.
            </p>

            {/* Error detail — collapsible */}
            {this.state.error && (
              <details className="text-left mt-4 mb-6">
                <summary
                  className="text-xs font-bold cursor-pointer select-none"
                  style={{ color: "var(--color-text-hint)" }}
                >
                  Technical details ▸
                </summary>
                <pre
                  className="mt-2 text-xs rounded-lg px-3 py-2 overflow-x-auto"
                  style={{
                    background: "rgba(0,0,0,0.35)",
                    color: "#f87171",
                    border: "1px solid rgba(248,113,113,0.20)",
                    whiteSpace: "pre-wrap",
                    wordBreak: "break-word",
                  }}
                >
                  {this.state.error.message}
                </pre>
              </details>
            )}

            <div className="flex flex-col sm:flex-row gap-3 justify-center">
              <button
                type="button"
                onClick={this.handleReset}
                className="px-6 py-3 rounded-lg text-sm font-black transition-all"
                style={{
                  background: "linear-gradient(135deg,rgba(109,40,217,0.85) 0%,rgba(30,90,180,0.80) 100%)",
                  border: "1px solid rgba(139,92,246,0.40)",
                  color: "#fff",
                }}
                onMouseEnter={e => {
                  e.currentTarget.style.background = "linear-gradient(135deg,rgba(139,92,246,0.90) 0%,rgba(34,150,220,0.85) 100%)";
                }}
                onMouseLeave={e => {
                  e.currentTarget.style.background = "linear-gradient(135deg,rgba(109,40,217,0.85) 0%,rgba(30,90,180,0.80) 100%)";
                }}
              >
                ↺ Try again
              </button>
              <button
                type="button"
                onClick={() => { window.location.reload(); }}
                className="px-6 py-3 rounded-lg text-sm font-bold transition-all"
                style={{
                  border: "1px solid rgba(255,255,255,0.14)",
                  color: "var(--color-text-secondary)",
                  background: "transparent",
                }}
                onMouseEnter={e => { e.currentTarget.style.borderColor = "rgba(255,255,255,0.30)"; }}
                onMouseLeave={e => { e.currentTarget.style.borderColor = "rgba(255,255,255,0.14)"; }}
              >
                ⟳ Reload page
              </button>
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
