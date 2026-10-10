import { create } from 'zustand'

/** `loading` until the saved session has been checked at startup. */
export type AuthStatus = 'loading' | 'signed-in' | 'signed-out'

interface AuthState {
  status: AuthStatus
  /** Keycloak access token sent as `Authorization: Bearer`. */
  token: string | null
  setToken: (token: string) => void
  clearSession: () => void
}

/**
 * Session state, kept in memory and filled by `startAuthSession` from the OIDC client.
 * The user's profile and role are server data: they come from `/me` through React Query.
 */
export const useAuthStore = create<AuthState>()((set) => ({
  status: 'loading',
  token: null,
  setToken: (token) => set({ status: 'signed-in', token }),
  clearSession: () => set({ status: 'signed-out', token: null }),
}))
