import { Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { AuthPanel } from '@/shared/components/auth-panel'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { useCompleteSignIn } from './use-complete-sign-in'

export function AuthCallbackPage() {
  const { t } = useTranslation()
  const { failed } = useCompleteSignIn()

  if (!failed) return <AuthPanel title={t('Sign in')} description={t('Signing you in…')} />
  return (
    <AuthPanel title={t('Sign in')} description={t('Could not finish signing in.')}>
      <Button className="w-full" nativeButton={false} render={<Link to={ROUTES.dashboard} replace />}>
        {t('Try again')}
      </Button>
    </AuthPanel>
  )
}
