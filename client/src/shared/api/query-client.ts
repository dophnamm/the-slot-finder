import { type Mutation, MutationCache, type Query, QueryCache, QueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { signOutLocally } from '@/shared/auth/session'
import { translate } from '@/shared/i18n'
import { QUERY_RETRY, QUERY_STALE_TIME_MS } from '@/utils/constants'
import { FORBIDDEN_STATUS, UNAUTHORIZED_STATUS } from './constants'
import type { ApiError } from './types'

declare module '@tanstack/react-query' {
  interface Register {
    // Every request goes through `http.ts`, which only rejects with ApiError.
    defaultError: ApiError
    // A query or mutation lists the error codes it shows itself, so the shared toast skips them.
    queryMeta: { quietErrorCodes?: readonly string[] }
    mutationMeta: { quietErrorCodes?: readonly string[] }
  }
}

/**
 * The one shared handler for failed queries and mutations. A 401 means the server no longer
 * accepts the token, so the session ends and the route guard sends the user to sign in.
 */
function notifyError(error: ApiError) {
  if (error.status === UNAUTHORIZED_STATUS) {
    toast.error(translate('Your session has ended. Sign in again.'))
    queryClient.clear()
    void signOutLocally()
    return
  }
  toast.error(error.message)
}

/** Like notifyError, except for the codes the query shows itself. */
function notifyQueryError(error: ApiError, query: Query<unknown, unknown, unknown>) {
  if (query.meta?.quietErrorCodes?.includes(error.code)) return
  notifyError(error)
}

/** Like notifyError, except for the codes the mutation shows itself. */
function notifyMutationError(
  error: ApiError,
  _variables: unknown,
  _context: unknown,
  mutation: Mutation<unknown, unknown, unknown>,
) {
  if (mutation.meta?.quietErrorCodes?.includes(error.code)) return
  notifyError(error)
}

/** Retries a failed query, except when the session has ended or the user may not see it. */
function shouldRetry(failureCount: number, error: ApiError) {
  return error.status !== UNAUTHORIZED_STATUS && error.status !== FORBIDDEN_STATUS && failureCount < QUERY_RETRY
}

export const queryClient = new QueryClient({
  queryCache: new QueryCache({ onError: notifyQueryError }),
  mutationCache: new MutationCache({ onError: notifyMutationError }),
  defaultOptions: {
    queries: {
      retry: shouldRetry,
      staleTime: QUERY_STALE_TIME_MS,
    },
  },
})
