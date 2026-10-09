import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function ProgramsPage() {
  const { t } = useTranslation()

  return <PagePlaceholder title={t('Programs')} description={t('List of programs with status, timeline and the four progress bars.')} />
}
