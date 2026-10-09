import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function IntakePage() {
  const { t } = useTranslation()

  return <PagePlaceholder title={t('Intake queue')} description={t('Incoming projects across programs, with “Propose squad”.')} />
}
