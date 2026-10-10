import type { Role } from '@/resources/rbac'

/** The signed-in user as the server reports it (`GET /me`). */
export interface CurrentUser {
  id: string
  username: string
  name: string
  email: string
  /** Empty when the account holds no application role. */
  role: Role | ''
  /** The user's person record; null when the account has none. */
  person_id: string | null
}

/** Router state given to the sign-in page: the page the user asked for. */
export interface LoginRedirectState {
  from: string
}

/** Carried through the Keycloak round trip so the user lands where they started. */
export interface SignInState {
  returnTo: string
}
