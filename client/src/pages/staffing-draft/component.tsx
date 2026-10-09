import { useParams } from 'react-router'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'

export function StaffingDraftPage() {
  const { id = '' } = useParams()
  const { t } = useTranslation()

  return <PagePlaceholder title={t('Staffing draft {id}', { id })} description={t('Review the assignment grid, reasons and diff; edit, lock and accept.')} />
}
