import { PROGRAM_STATUS_LABELS, PROJECT_STATUS_LABELS, type ProgramStatus, type ProjectStatus } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { STATUS_VARIANT } from './constants'
import type { StatusBadgeProps } from './types'

export function StatusBadge({ status, kind }: StatusBadgeProps) {
  const { t } = useTranslation()
  const label =
    kind === 'program' ? PROGRAM_STATUS_LABELS[status as ProgramStatus] : PROJECT_STATUS_LABELS[status as ProjectStatus]

  return <Badge variant={STATUS_VARIANT[status]}>{t(label)}</Badge>
}
