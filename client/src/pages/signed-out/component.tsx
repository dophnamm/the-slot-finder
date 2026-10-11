import { Link, Navigate } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { AuthPanel } from '@/shared/components/auth-panel'
import { useTranslation } from '@/shared/i18n'
import { useAuthStore } from '@/shared/stores/auth-store'
import { Button } from '@/shared/ui/button'

/** Where Keycloak returns after sign-out. It stays here, so the app does not sign straight back in. */
export function SignedOutPage() {
  const { t } = useTranslation()
  const status = useAuthStore((state) => state.status)

  if (status === 'signed-in') return <Navigate to={ROUTES.dashboard} replace />
  return (
    <AuthPanel title={t('Signed out')} description={t('You have signed out.')}>
      <Button className="w-full" nativeButton={false} render={<Link to={ROUTES.dashboard} replace />}>
        {t('Sign in again')}
      </Button>
    </AuthPanel>
  )
}
