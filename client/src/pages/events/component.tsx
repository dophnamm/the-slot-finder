import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function EventsPage() {
  const { t } = useTranslation()

  return <PagePlaceholder title={t('Change history')} description={t('Change history and re-plan status.')} />
}
