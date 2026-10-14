import { ClipboardList } from 'lucide-react'
import { MetricBar } from '@/shared/components/metric-bar'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import { STAFFING_LOADING_ROWS, STAFFING_TABLE_COLUMNS } from './constants'
import type { ExpertiseStaffingProps } from './types'

/** Budget against planned hours per expertise across the program. */
export function ExpertiseStaffing({ rows, isPending }: ExpertiseStaffingProps) {
  const { t } = useTranslation()

  return (
    <Section
      title={t('Staffing by expertise')}
      description={t('Budget hours from the demand against hours planned in the baseline.')}
    >
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
              <TableCell className="text-right tabular-nums">
                {t('{hours} h', { hours: Math.round(row.budget_hours) })}
              </TableCell>
              <TableCell className="text-right tabular-nums">
                {t('{hours} h', { hours: Math.round(row.planned_hours) })}
              </TableCell>
              <TableCell className={cn('text-right tabular-nums', row.understaffed && 'text-destructive')}>
                {t('{hours} h', { hours: Math.round(row.gap_hours) })}
              </TableCell>
              <TableCell>
                <MetricBar compact label={t('Staffing')} value={row.staffing_percent} warning={row.understaffed} />
              </TableCell>
            </TableRow>
          ))}
          {isPending && <TableLoading columns={STAFFING_TABLE_COLUMNS} rows={STAFFING_LOADING_ROWS} />}
          {!isPending && rows.length === 0 && (
            <TableEmpty columns={STAFFING_TABLE_COLUMNS} icon={ClipboardList} title={t('No demand yet.')} />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
