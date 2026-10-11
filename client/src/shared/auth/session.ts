import type { User } from 'oidc-client-ts'
import { useAuthStore } from '@/shared/stores/auth-store'
import type { SignInState } from './types'
import { userManager } from './user-manager'

function applyUser(user: User | null) {
  const { setToken, clearSession } = useAuthStore.getState()
  if (user && !user.expired) setToken(user.access_token)
  else clearSession()
}

const handleUserUnloaded = () => applyUser(null)

/**
 * Restores a saved session and keeps the auth store in step with the OIDC client: a silent
 * renewal swaps the token, and an expired or ended session clears it. Call it once at startup.
 */
export async function startAuthSession() {
  userManager.events.addUserLoaded(applyUser)
  userManager.events.addUserUnloaded(handleUserUnloaded)
  userManager.events.addAccessTokenExpired(() => applyUser(null))
  userManager.events.addSilentRenewError(() => void signOutLocally())
  applyUser(await userManager.getUser())
}

/** Sends the browser to the Keycloak sign-in page. */
export function signIn(returnTo: string) {
  const state: SignInState = { returnTo }
  return userManager.signinRedirect({ state })
}

let signInCompletion: Promise<SignInState | undefined> | null = null

/**
 * Finishes the code exchange on the callback page and returns where the user started. The
 * promise is shared, because React may run the page's effect twice and a code works only once.
 */
export function completeSignIn() {
  signInCompletion ??= userManager.signinRedirectCallback().then((user) => {
    applyUser(user)
    return user.state as SignInState | undefined
  })
  return signInCompletion
}

/**
 * Ends the Keycloak session and comes back to the Signed out page. The app stays as it is while
 * the browser leaves: clearing the session first would make RequireAuth start a sign-in that
 * races the sign-out. If Keycloak cannot be reached, the session ends in this tab only.
 */
export async function signOut() {
  userManager.events.removeUserUnloaded(handleUserUnloaded)
  try {
    await userManager.signoutRedirect()
  } catch {
    userManager.events.addUserUnloaded(handleUserUnloaded)
    await signOutLocally()
  }
}

/** Forgets the session in this tab only, e.g. after the server rejected the token. */
export async function signOutLocally() {
  await userManager.removeUser()
  useAuthStore.getState().clearSession()
}
