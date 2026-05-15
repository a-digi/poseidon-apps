import { useState } from 'react';
import { HOST, setAdminToken } from './api';

interface AdminLoginProps {
  onAuthenticated: () => void;
}

export function AdminLogin({ onAuthenticated }: AdminLoginProps) {
  const [token, setToken] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (token.trim().length === 0) return;
    setSubmitting(true);
    setError(null);
    try {
      const r = await fetch(`${HOST}/plugins/coco-gg/api/games`, {
        headers: { Authorization: `Bearer ${token.trim()}` },
      });
      if (!r.ok) {
        setError(r.status === 401 ? 'Invalid admin token.' : `Login failed: ${r.statusText}`);
        return;
      }
      setAdminToken(token.trim());
      onAuthenticated();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Network error.');
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="min-h-screen bg-slate-50 flex items-center justify-center p-4">
      <form onSubmit={handleSubmit} className="bg-white border border-slate-200 rounded-lg shadow-sm p-6 w-full max-w-sm">
        <h1 className="text-lg font-bold text-slate-900 mb-1">Coco GG · Admin</h1>
        <p className="text-xs text-slate-500 mb-4">Enter the operator admin token to manage the server.</p>
        <label htmlFor="admin-token" className="block text-xs font-medium text-slate-700 mb-1">Admin token</label>
        <input
          id="admin-token"
          type="password"
          autoFocus
          value={token}
          onChange={(e) => setToken(e.target.value)}
          className="w-full rounded border border-slate-300 px-3 py-1.5 text-sm focus:outline-none focus:border-slate-500"
          placeholder="paste token"
        />
        {error !== null && <p className="mt-2 text-xs text-red-600">{error}</p>}
        <button
          type="submit"
          disabled={submitting || token.trim().length === 0}
          className="mt-3 w-full rounded bg-slate-900 px-3 py-1.5 text-xs font-medium text-white transition-colors hover:bg-slate-700 disabled:bg-slate-400"
        >
          {submitting ? 'Signing in…' : 'Sign in'}
        </button>
      </form>
    </div>
  );
}
