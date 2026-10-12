import { GitCompareArrows, Plus } from 'lucide-react'
import { useState } from 'react'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { useScenarios } from '@/shared/scenarios'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { SCENARIO_TABLE_COLUMNS } from './constants'
import { NewScenarioDialog } from './components/new-scenario-dialog'
import { ScenarioRow } from './components/scenario-row'

/** The planner's what-ifs: each one's assumptions, how much it changes the plan, and whether it is stale. */
export function Scenarios() {
  const { t } = useTranslation()
  const { data: scenarios = [], isPending } = useScenarios()
  const [creating, setCreating] = useState(false)

  const handleNew = () => setCreating(true)

  return (
    <Section
      title={t('What-if scenarios')}
      description={t(
        'Try a change, such as a project slipping or someone on leave, and see what it does to the plan. Nothing real changes.',
      )}
      action={
        <Button onClick={handleNew}>
          <Plus />
          {t('New scenario')}
        </Button>
      }
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Scenario')}</TableHead>
            <TableHead>{t('Assumptions')}</TableHead>
            <TableHead>{t('Baseline')}</TableHead>
            <TableHead className="text-right">{t('People changed')}</TableHead>
            <TableHead className="text-right">{t('Hours moved')}</TableHead>
            <TableHead className="text-right">{t('Unmet')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {scenarios.map((scenario) => (
            <ScenarioRow key={scenario.id} scenario={scenario} />
          ))}
          {isPending && <TableLoading columns={SCENARIO_TABLE_COLUMNS} />}
          {!isPending && scenarios.length === 0 && (
            <TableEmpty
              columns={SCENARIO_TABLE_COLUMNS}
              icon={GitCompareArrows}
              title={t('No scenarios yet. Create one to try a change.')}
            />
          )}
        </TableBody>
      </Table>
      {creating && <NewScenarioDialog open={creating} onOpenChange={setCreating} />}
    </Section>
  )
}
