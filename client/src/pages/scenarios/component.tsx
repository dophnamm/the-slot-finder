import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function ScenariosPage() {
  const { t } = useTranslation()

  return <PagePlaceholder title={t('What-if scenarios')} description={t('Build scenarios (move a project, add leave) and view the diff.')} />
}
