import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import type { ScenarioProject } from '@/shared/scenarios'

/** The page of a project the scenario planned again. */
export function projectPath(project: ScenarioProject): string {
  return generatePath(ROUTES.projectDetail, { programId: project.program_id, id: project.id })
}
