import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import type { ScenarioChangesProps } from './types'
import { changeBadge, weekSpan } from './utils'

/** Who gains or loses hours, or works in other weeks, against the baseline. */
export function ScenarioChanges({ changes }: ScenarioChangesProps) {
  const { t } = useTranslation()

  return (
    <Section title={t('Changes against the baseline')} description={t('Each person and demand whose hours or weeks differ in this scenario.')}>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Person')}</TableHead>
            <TableHead>{t('Project')}</TableHead>
            <TableHead>{t('Expertise')}</TableHead>
            <TableHead>{t('Before')}</TableHead>
            <TableHead>{t('After')}</TableHead>
            <TableHead>{t('Change')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {changes.map((change) => (
            <TableRow key={`${change.person.id}:${change.demand_id}`}>
              <TableCell className="font-medium">{change.person.name}</TableCell>
              <TableCell>{change.project}</TableCell>
              <TableCell>{change.expertise}</TableCell>
              <TableCell>
                <div className="flex flex-col">
                  <span className="tabular-nums">{t('{hours} h', { hours: change.before_hours })}</span>
                  <span className="text-xs text-muted-foreground">{weekSpan(change.before_from, change.before_to)}</span>
                </div>
              </TableCell>
              <TableCell>
                <div className="flex flex-col">
                  <span className="tabular-nums">{t('{hours} h', { hours: change.after_hours })}</span>
                  <span className="text-xs text-muted-foreground">{weekSpan(change.after_from, change.after_to)}</span>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant={changeBadge(change).variant}>{t(changeBadge(change).label)}</Badge>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      {changes.length === 0 && <p className="py-6 text-center text-sm text-muted-foreground">{t('Nobody’s hours change in this scenario.')}</p>}
    </Section>
  )
}
