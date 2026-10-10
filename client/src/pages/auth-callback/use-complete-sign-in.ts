import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { completeSignIn } from '@/shared/auth'

/** Finishes the Keycloak round trip and goes back to where the user started. */
export function useCompleteSignIn() {
  const navigate = useNavigate()
  const [failed, setFailed] = useState(false)

  useEffect(() => {
    completeSignIn()
      .then((state) => navigate(state?.returnTo ?? ROUTES.dashboard, { replace: true }))
      .catch(() => setFailed(true))
  }, [navigate])

  return { failed }
}
