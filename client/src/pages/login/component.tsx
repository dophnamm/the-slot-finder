import { LogIn } from 'lucide-react'
import { Navigate, useLocation } from 'react-router'
import { toast } from 'sonner'
import { signIn } from '@/shared/auth'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'
import { useAuthStore } from '@/shared/stores/auth-store'
import { Button } from '@/shared/ui/button'
import { returnPath } from './utils'

export function LoginPage() {
  const { t } = useTranslation()
  const location = useLocation()
  const status = useAuthStore((state) => state.status)
  const returnTo = returnPath(location.state)

  const handleSignInError = () => {
    toast.error(t('Could not reach the sign-in service.'))
  }
  const handleSignIn = () => {
    signIn(returnTo).catch(handleSignInError)
  }

  if (status === 'signed-in') return <Navigate to={returnTo} replace />

  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <div className="w-full max-w-sm">
        <PagePlaceholder title={t('Sign in')} description={t('Sign in to reach the dashboard or your own schedule.')}>
          <Button className="w-full" onClick={handleSignIn} disabled={status === 'loading'}>
            <LogIn />
            {t('Sign in with Keycloak')}
          </Button>
        </PagePlaceholder>
      </div>
    </div>
  )
}
