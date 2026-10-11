import { Outlet, useLocation } from 'react-router'
import { AuthPanel } from '@/shared/components/auth-panel'
import { useTranslation } from '@/shared/i18n'
import { useAuthStore } from '@/shared/stores/auth-store'
import { Button } from '@/shared/ui/button'
import { useSignInRedirect } from './use-sign-in-redirect'
import { returnPath } from './utils'

/** Renders the app only for a signed-in user; others go straight to the Keycloak sign-in page. */
export function RequireAuth() {
  const { t } = useTranslation()
  const status = useAuthStore((state) => state.status)
  const location = useLocation()
  const { failed, retry } = useSignInRedirect(status === 'signed-out', returnPath(location))

  if (status === 'signed-in') return <Outlet />
  if (!failed) return null
  return (
    <AuthPanel title={t('Sign in')} description={t('Could not reach the sign-in service.')}>
      <Button className="w-full" onClick={retry}>
        {t('Try again')}
      </Button>
    </AuthPanel>
  )
}
