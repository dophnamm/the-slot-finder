import { ArrowLeftRight } from 'lucide-react'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { PLAN_CHANGE_COLUMNS } from './constants'
import { ChangeRow } from './components/change-row'
import type { PlanChangesProps } from './types'

/** Who gains or loses hours against the baseline in force if the draft is accepted. */
export function PlanChanges({ changes }: PlanChangesProps) {
  const { t } = useTranslation()

  return (
    <Section
      title={t('Changes against the baseline')}
      description={t('Who gains or loses hours if this draft is accepted.')}
    >
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
          {changes.length === 0 && (
            <TableEmpty
              columns={PLAN_CHANGE_COLUMNS}
              icon={ArrowLeftRight}
              title={t('No changes against the baseline.')}
            />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
