import { useQuery } from '@tanstack/react-query'
import { IDENTITY_ENDPOINTS } from '@/resources/endpoints'
import { request } from '@/shared/api/http'
import { useAuthStore } from '@/shared/stores/auth-store'
import { CURRENT_USER_QUERY_KEY } from './constants'
import { clearAutoSignIns } from './sign-in-attempts'
import type { CurrentUser } from './types'

async function getCurrentUser() {
  const user = await request<CurrentUser>({ url: IDENTITY_ENDPOINTS.me })
  // The server accepts the session, so automatic sign-ins are not going round in a loop.
  clearAutoSignIns()
  return user
}

/** The signed-in user from the server. Fetched once a token is present. */
export function useCurrentUser() {
  const signedIn = useAuthStore((state) => state.status === 'signed-in')
  return useQuery({ queryKey: CURRENT_USER_QUERY_KEY, queryFn: getCurrentUser, enabled: signedIn })
}
