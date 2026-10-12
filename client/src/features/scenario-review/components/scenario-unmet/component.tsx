import { CircleCheck } from 'lucide-react'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { useTranslation } from '@/shared/i18n'
import { UNMET_REASON_LABELS } from '@/shared/staffing'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { UNMET_TABLE_COLUMNS } from './constants'
import type { ScenarioUnmetProps } from './types'

/** Demand hours nobody could take in this scenario, with the engine's reason. */
export function ScenarioUnmet({ unmet }: ScenarioUnmetProps) {
  const { t } = useTranslation()

  return (
    <Section title={t('Unmet demand')} description={t('Hours the planning engine could not place in this scenario.')}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Project')}</TableHead>
            <TableHead>{t('Expertise')}</TableHead>
            <TableHead className="text-right">{t('Hours')}</TableHead>
            <TableHead>{t('Reason')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {unmet.map((row) => (
            <TableRow key={row.demand_id}>
              <TableCell className="font-medium">{row.project}</TableCell>
              <TableCell>{row.expertise}</TableCell>
              <TableCell className="text-right text-destructive tabular-nums">
                {t('{hours} h', { hours: row.hours })}
              </TableCell>
              <TableCell>{t(UNMET_REASON_LABELS[row.reason])}</TableCell>
            </TableRow>
          ))}
          {unmet.length === 0 && (
            <TableEmpty columns={UNMET_TABLE_COLUMNS} icon={CircleCheck} title={t('Every demand is covered.')} />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
