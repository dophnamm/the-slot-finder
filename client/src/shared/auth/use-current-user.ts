import { useQuery } from '@tanstack/react-query'
import { IDENTITY_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import { useAuthStore } from '@/shared/stores/auth-store'
import { CURRENT_USER_QUERY_KEY } from './constants'
import type { CurrentUser } from './types'

function getCurrentUser() {
  return request<CurrentUser>({ url: IDENTITY_ENDPOINTS.me })
}

/** The signed-in user from the server. Fetched once a token is present. */
export function useCurrentUser() {
  const signedIn = useAuthStore((state) => state.status === 'signed-in')
  return useQuery({ queryKey: CURRENT_USER_QUERY_KEY, queryFn: getCurrentUser, enabled: signedIn })
}
