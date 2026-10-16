import { GitCompareArrows, Plus } from 'lucide-react'
import { useState } from 'react'
import { LoadError } from '@/shared/components/load-error'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { Term } from '@/shared/components/term'
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
  const { data, error, isPending, refetch } = useScenarios()
  const scenarios = data ?? []
  const [creating, setCreating] = useState(false)

  const handleNew = () => setCreating(true)
  const handleRetry = () => void refetch()

  return (
    <Section
      className="min-h-0 flex-1"
      title={t('Scenarios')}
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
      <Table scroll="both">
        <TableHeader>
          <TableRow>
            <TableHead>
              <Term term="scenario">{t('Scenario')}</Term>
            </TableHead>
            <TableHead>{t('Assumptions')}</TableHead>
            <TableHead>
              <Term term="baseline">{t('Baseline')}</Term>
            </TableHead>
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
          {error && !data && <LoadError columns={SCENARIO_TABLE_COLUMNS} error={error} onRetry={handleRetry} />}
          {!isPending && !error && scenarios.length === 0 && (
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
