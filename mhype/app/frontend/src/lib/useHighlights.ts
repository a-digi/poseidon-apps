import { useState, useRef, useCallback, useEffect } from 'react';
import { getHighlights } from '../api';
import type { Highlights } from '../api';

export interface UseHighlights {
  data: Highlights | null;
  loading: boolean;
  error: string | null;
  refetch: () => void;
}

export function useHighlights(): UseHighlights {
  const [data, setData] = useState<Highlights | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const reqRef = useRef(0);

  const refetch = useCallback(() => {
    const id = ++reqRef.current;
    setLoading(true);
    getHighlights(10)
      .then(result => {
        if (reqRef.current !== id) return;
        setData(result);
        setError(null);
      })
      .catch(err => {
        if (reqRef.current !== id) return;
        setError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (reqRef.current === id) setLoading(false);
      });
  }, []);

  useEffect(() => {
    refetch();
  }, [refetch]);

  return { data, loading, error, refetch };
}
