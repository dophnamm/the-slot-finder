import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { UNMET_REASON_LABELS } from '@/shared/staffing'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
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
              <TableCell className="text-right tabular-nums text-destructive">{t('{hours} h', { hours: row.hours })}</TableCell>
              <TableCell>{t(UNMET_REASON_LABELS[row.reason])}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {unmet.length === 0 && <p className="py-6 text-center text-sm text-muted-foreground">{t('Every demand is covered.')}</p>}
    </Section>
  )
}
