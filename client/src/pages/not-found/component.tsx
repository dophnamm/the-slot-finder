import { Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'

export function NotFoundPage() {
  const { t } = useTranslation()

  return (
    <PagePlaceholder title={t('Page not found')} description={t('There is no screen at this address.')}>
      <Button nativeButton={false} render={<Link to={ROUTES.dashboard} />}>
        {t('Back to dashboard')}
      </Button>
    </PagePlaceholder>
  )
}
