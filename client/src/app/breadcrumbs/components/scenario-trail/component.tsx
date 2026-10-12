import { ROUTES } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'
import { useScenario } from '@/shared/scenarios'
import { Trail } from '../trail'
import type { ScenarioTrailProps } from './types'

/** Scenarios › the scenario's name. */
export function ScenarioTrail({ scenarioId }: ScenarioTrailProps) {
  const { t } = useTranslation()
  const { data: scenario } = useScenario(scenarioId)

  return <Trail crumbs={[{ label: t('Scenarios'), to: ROUTES.scenarios }, { label: scenario?.name ?? t('Loading…') }]} />
}
