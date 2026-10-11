import { useTranslation } from '@/shared/i18n'
import { TableCell, TableRow } from '@/shared/ui/table'
import type { DemandViewRowProps } from './types'

export function DemandViewRow({ demand, phase }: DemandViewRowProps) {
  const { t } = useTranslation()

  return (
    <TableRow>
      <TableCell className="font-medium">{demand.expertise}</TableCell>
      <TableCell>{phase ?? t('Whole project')}</TableCell>
      <TableCell>{t('Level {level}', { level: demand.min_level })}</TableCell>
      <TableCell>{demand.headcount}</TableCell>
      <TableCell>{demand.total_hours}</TableCell>
      <TableCell className="text-sm text-muted-foreground">{t('{start} – {end}', { start: demand.from_week, end: demand.to_week })}</TableCell>
    </TableRow>
  )
}
