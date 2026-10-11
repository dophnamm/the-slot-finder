import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { ChangeRow } from './components/change-row'
import type { PlanChangesProps } from './types'

/** Who gains or loses hours against the baseline in force if the draft is accepted. */
export function PlanChanges({ changes }: PlanChangesProps) {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Changes against the baseline')}</CardTitle>
        <CardDescription>{t('Who gains or loses hours if this draft is accepted.')}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Person')}</TableHead>
              <TableHead>{t('Project')}</TableHead>
              <TableHead>{t('Expertise')}</TableHead>
              <TableHead className="text-right">{t('Before')}</TableHead>
              <TableHead className="text-right">{t('After')}</TableHead>
              <TableHead>{t('Change')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {changes.map((change) => (
              <ChangeRow key={`${change.person.id}:${change.demand_id}`} change={change} />
            ))}
          </TableBody>
        </Table>
        {changes.length === 0 && <p className="py-6 text-center text-sm text-muted-foreground">{t('No changes against the baseline.')}</p>}
      </CardContent>
    </Card>
  )
}
