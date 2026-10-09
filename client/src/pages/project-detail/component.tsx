import { useParams } from 'react-router'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function ProjectDetailPage() {
  const { id = '' } = useParams()
  const { t } = useTranslation()

  return <PagePlaceholder title={t('Project {id}', { id })} description={t('Demand, timeline, assigned people, risks and scope points.')} />
}
