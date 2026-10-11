import { activeWarnings, type Project } from '@/shared/programs'

/** The project's raised warnings; none while its metrics are not loaded. */
export function projectWarnings(project: Project) {
  return project.metrics ? activeWarnings(project.metrics) : []
}
