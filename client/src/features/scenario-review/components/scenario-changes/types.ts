import type { ScenarioChange, ScenarioProject } from '@/shared/scenarios'

export interface ScenarioChangesProps {
  changes: ScenarioChange[]
  /** The projects the scenario plans again; a change links to its project through them. */
  projects: ScenarioProject[]
}
