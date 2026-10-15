import type { DraftProject } from '@/shared/staffing'

/** The programs of a draft's projects, each once, in project order. */
export function programNames(projects: DraftProject[]): string {
  return [...new Set(projects.map((project) => project.program))].join(', ')
}
