import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function PeoplePage() {
  const { t } = useTranslation()

  return <PagePlaceholder title={t('People')} description={t('Expertise and level matrix, leave calendar and capacity.')} />
}
