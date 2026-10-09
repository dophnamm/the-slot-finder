import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'
import { ServerStatus } from './components/server-status'

export function DashboardPage() {
  const { t } = useTranslation()

  return (
    <PagePlaceholder
      title={t('Dashboard')}
      description={t('Person × week heatmap, with overload and deadline-risk alerts.')}
      action={<ServerStatus />}
    />
  )
}
