import { useHealth } from '@/features/health'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'

/** Shows whether the server API answers its health check. */
export function ServerStatus() {
  const { isPending, isSuccess } = useHealth()
  const { t } = useTranslation()

  if (isPending) return <Badge variant="outline">{t('Server: checking…')}</Badge>
  if (isSuccess) return <Badge>{t('Server: OK')}</Badge>
  return <Badge variant="destructive">{t('Server: unreachable')}</Badge>
}
