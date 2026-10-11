import type { Project } from '@/shared/programs'

export interface PreviewProjectsProps {
  projects: Project[]
  isPending: boolean
  programId: string
}
