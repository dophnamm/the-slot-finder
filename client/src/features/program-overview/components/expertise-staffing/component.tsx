import { MetricBar } from '@/shared/components/metric-bar'
import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import type { ExpertiseStaffingProps } from './types'

/** Budget against planned hours per expertise across the program. */
export function ExpertiseStaffing({ rows }: ExpertiseStaffingProps) {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Staffing by expertise')}</CardTitle>
        <CardDescription>{t('Budget hours from the demand against hours planned in the baseline.')}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Expertise')}</TableHead>
              <TableHead className="text-right">{t('Budget')}</TableHead>
              <TableHead className="text-right">{t('Planned')}</TableHead>
              <TableHead className="text-right">{t('Gap')}</TableHead>
              <TableHead className="w-48">{t('Staffing')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {rows.map((row) => (
              <TableRow key={row.expertise_id}>
                <TableCell className="font-medium">{row.expertise}</TableCell>
                <TableCell className="text-right tabular-nums">{t('{hours} h', { hours: Math.round(row.budget_hours) })}</TableCell>
                <TableCell className="text-right tabular-nums">{t('{hours} h', { hours: Math.round(row.planned_hours) })}</TableCell>
                <TableCell className={cn('text-right tabular-nums', row.understaffed && 'text-destructive')}>
                  {t('{hours} h', { hours: Math.round(row.gap_hours) })}
                </TableCell>
                <TableCell>
                  <MetricBar compact label={t('Staffing')} value={row.staffing_percent} warning={row.understaffed} />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        {rows.length === 0 && <p className="py-6 text-center text-sm text-muted-foreground">{t('No demand yet.')}</p>}
      </CardContent>
    </Card>
  )
}
