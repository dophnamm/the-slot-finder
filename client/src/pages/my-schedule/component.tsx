import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function MySchedulePage() {
  const { t } = useTranslation()

  return <PagePlaceholder title={t('My schedule')} description={t('Your projects by week, and a place to report leave.')} />
}
