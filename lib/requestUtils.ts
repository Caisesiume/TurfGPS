import { useRef, useCallback } from 'react';

/**
 * Hook to manage request race conditions
 * Returns a function that can be used to check if a request is still the latest
 */
export function useRequestManager() {
  const currentRequestRef = useRef<number>(0);

  const createRequest = useCallback(() => {
    const requestId = Date.now() + Math.random(); // More unique ID
    currentRequestRef.current = requestId;
    
    return {
      id: requestId,
      isLatest: () => currentRequestRef.current === requestId,
    };
  }, []);

  return { createRequest };
}

/**
 * Hook for debounced API calls with race condition protection
 */
export function useDebounceCallback<T extends (...args: unknown[]) => Promise<void>>(
  callback: T,
  delay: number
): T {
  const timeoutRef = useRef<NodeJS.Timeout>();
  
  // eslint-disable-next-line react-hooks/exhaustive-deps
  const debouncedCallback = useCallback(((...args: Parameters<T>) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    
    timeoutRef.current = setTimeout(() => {
      callback(...args);
    }, delay);
  }) as T, [callback, delay]);
  
  return debouncedCallback;
}
