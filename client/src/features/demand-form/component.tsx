import { useCurrentUser } from '@/shared/auth'
import { useExpertises } from '@/shared/expertises'
import { canEditProject, useProject } from '@/shared/programs'
import { DemandEditor } from './components/demand-editor'
import type { DemandFormProps } from './types'
import { demandsKey } from './utils'

/** The project's demand: budget hours per expertise and minimum level, in one phase or the whole project. */
export function DemandForm({ projectId }: DemandFormProps) {
  const { data: user } = useCurrentUser()
  const { data: project } = useProject(projectId)
  const { data: expertises = [] } = useExpertises()

  if (!project) return null

  return <DemandEditor key={demandsKey(project.demands)} project={project} expertises={expertises} editable={canEditProject(user, project)} />
}
