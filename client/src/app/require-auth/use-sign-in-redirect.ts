import { useCallback, useEffect, useRef, useState } from 'react'
import { signIn } from '@/shared/auth'

/**
 * Sends a signed-out user straight to the Keycloak sign-in page, once per mount. `failed` is true
 * when Keycloak could not be reached; `retry` tries again.
 */
export function useSignInRedirect(signedOut: boolean, returnTo: string) {
  const started = useRef(false)
  const [failed, setFailed] = useState(false)

  const redirect = useCallback(() => {
    started.current = true
    signIn(returnTo).catch(() => setFailed(true))
  }, [returnTo])

  const retry = useCallback(() => {
    setFailed(false)
    redirect()
  }, [redirect])

  useEffect(() => {
    if (signedOut && !started.current) redirect()
  }, [signedOut, redirect])

  return { failed, retry }
}
