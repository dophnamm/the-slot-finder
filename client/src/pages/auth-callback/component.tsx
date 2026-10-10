import { Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { useCompleteSignIn } from './use-complete-sign-in'

export function AuthCallbackPage() {
  const { t } = useTranslation()
  const { failed } = useCompleteSignIn()

  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <div className="w-full max-w-sm">
        {failed ? (
          <PagePlaceholder title={t('Sign in')} description={t('Could not finish signing in.')}>
            <Button className="w-full" nativeButton={false} render={<Link to={ROUTES.login} replace />}>
              {t('Try again')}
            </Button>
          </PagePlaceholder>
        ) : (
          <PagePlaceholder title={t('Sign in')} description={t('Signing you in…')} />
        )}
      </div>
    </div>
  )
}
