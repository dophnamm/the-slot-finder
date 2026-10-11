import { useTranslation } from '@/shared/i18n'
import { PLAN_CHANGE_LABELS } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { TableCell, TableRow } from '@/shared/ui/table'
import { CHANGE_BADGE_VARIANT } from './constants'
import type { ChangeRowProps } from './types'

export function ChangeRow({ change }: ChangeRowProps) {
  const { t } = useTranslation()

  return (
    <TableRow>
      <TableCell className="font-medium">{change.person.name}</TableCell>
      <TableCell>{change.project}</TableCell>
      <TableCell>{change.expertise}</TableCell>
      <TableCell className="text-right">{change.before_hours}</TableCell>
      <TableCell className="text-right">{change.after_hours}</TableCell>
      <TableCell>
        <Badge variant={CHANGE_BADGE_VARIANT[change.change]}>{t(PLAN_CHANGE_LABELS[change.change])}</Badge>
      </TableCell>
    </TableRow>
  )
}
