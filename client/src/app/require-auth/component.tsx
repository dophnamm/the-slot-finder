import { Navigate, Outlet, useLocation } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useAuthStore } from '@/shared/stores/auth-store'
import { loginRedirectState } from './utils'

/** Renders the app only for a signed-in user; others go to the sign-in page. */
export function RequireAuth() {
  const status = useAuthStore((state) => state.status)
  const location = useLocation()

  if (status === 'loading') return null
  if (status === 'signed-out') return <Navigate to={ROUTES.login} replace state={loginRedirectState(location)} />
  return <Outlet />
}
