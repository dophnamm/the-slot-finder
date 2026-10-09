import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function LoginPage() {
  const { t } = useTranslation()

  return (
    <div className="flex min-h-svh items-center justify-center p-6">
      <div className="w-full max-w-sm">
        <PagePlaceholder
          title={t('Log in')}
          description={t('Sign in to reach the dashboard or your own schedule.')}
        />
      </div>
    </div>
  )
}
