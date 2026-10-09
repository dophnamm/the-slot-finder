import { create } from 'zustand'
import type { Role } from '@/resources/rbac'

export interface AuthUser {
  id: string
  name: string
  email: string
  role: Role
}

interface AuthState {
  token: string | null
  user: AuthUser | null
  setSession: (token: string, user: AuthUser) => void
  clearSession: () => void
}

/**
 * Client-side session state, kept in memory only (no persistence) until the
 * Keycloak login is wired in. Server data never goes here: it lives in React Query.
 */
export const useAuthStore = create<AuthState>()((set) => ({
  token: null,
  user: null,
  setSession: (token, user) => set({ token, user }),
  clearSession: () => set({ token: null, user: null }),
}))
