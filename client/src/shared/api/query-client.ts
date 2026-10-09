import { MutationCache, QueryCache, QueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { QUERY_RETRY, QUERY_STALE_TIME_MS } from '@/utils/constants'
import type { ApiError } from './types'

declare module '@tanstack/react-query' {
  interface Register {
    // Every request goes through `http.ts`, which only rejects with ApiError.
    defaultError: ApiError
  }
}

/** The one shared handler for failed queries and mutations. */
function notifyError(error: ApiError) {
  toast.error(error.message)
}

export const queryClient = new QueryClient({
  queryCache: new QueryCache({ onError: notifyError }),
  mutationCache: new MutationCache({ onError: notifyError }),
  defaultOptions: {
    queries: {
      retry: QUERY_RETRY,
      staleTime: QUERY_STALE_TIME_MS,
    },
  },
})
