import { useTranslation } from '@/shared/i18n'
import { shortfall, UNMET_REASON_LABELS } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { TableCell, TableRow } from '@/shared/ui/table'
import type { CoverageRowProps } from './types'

export function CoverageRow({ demand }: CoverageRowProps) {
  const { t } = useTranslation()
  const missing = shortfall(demand)

  return (
    <TableRow>
      <TableCell className="font-medium">{demand.project}</TableCell>
      <TableCell>{t('{expertise} (level {level}+)', { expertise: demand.expertise, level: demand.min_level })}</TableCell>
      <TableCell>{demand.phase ?? t('Whole project')}</TableCell>
      <TableCell className="text-sm text-muted-foreground">{t('{start} – {end}', { start: demand.from_week, end: demand.to_week })}</TableCell>
      <TableCell className="text-right">{demand.hours}</TableCell>
      <TableCell className="text-right">{demand.assigned_hours}</TableCell>
      <TableCell>
        {missing === 0 ? (
          <Badge variant="secondary">{t('Covered')}</Badge>
        ) : (
          <div className="flex flex-col gap-1">
            <Badge variant="destructive">{t('Short {hours} h', { hours: missing })}</Badge>
            {demand.unmet_reason && <span className="text-xs text-muted-foreground">{t(UNMET_REASON_LABELS[demand.unmet_reason])}</span>}
          </div>
        )}
      </TableCell>
    </TableRow>
  )
}
