import { useCallback, useEffect, useRef, useState } from 'react'
import { autoSignIn, signIn } from '@/shared/auth'
import type { SignInFailure } from './types'

/**
 * Sends a signed-out user straight to the Keycloak sign-in page, once per mount. `failure` says
 * why it stopped instead; `retry` tries again on the user's click.
 */
export function useSignInRedirect(signedOut: boolean, returnTo: string) {
  const started = useRef(false)
  const [failure, setFailure] = useState<SignInFailure | null>(null)

  useEffect(() => {
    if (!signedOut || started.current) return
    started.current = true
    autoSignIn(returnTo)
      .then((redirected) => {
        if (!redirected) setFailure('repeated')
      })
      .catch(() => setFailure('unreachable'))
  }, [signedOut, returnTo])

  const retry = useCallback(() => {
    setFailure(null)
    signIn(returnTo).catch(() => setFailure('unreachable'))
  }, [returnTo])

  return { failure, retry }
}
