import { Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'

/** Shown in place of a page the signed-in user's role cannot open. */
export function NoAccessPage() {
  const { t } = useTranslation()

  return (
    <PagePlaceholder title={t('No access')} description={t('Your role cannot open this page.')}>
      <Button nativeButton={false} render={<Link to={ROUTES.dashboard} />}>
        {t('Go to your home page')}
      </Button>
    </PagePlaceholder>
  )
}
