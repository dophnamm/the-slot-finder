import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import type { ScheduleProject } from '@/shared/people'

/** The page of the project behind one line of the week. */
export function projectPath(project: ScheduleProject): string {
  return generatePath(ROUTES.projectDetail, { programId: project.program_id, id: project.project_id })
}
