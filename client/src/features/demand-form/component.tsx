import { useCurrentUser } from '@/shared/auth'
import { Section } from '@/shared/components/section'
import { TableLoading } from '@/shared/components/table-loading'
import { useExpertises } from '@/shared/expertises'
import { useTranslation } from '@/shared/i18n'
import { canEditProject, useProject } from '@/shared/programs'
import { Table, TableBody } from '@/shared/ui/table'
import { DemandEditor } from './components/demand-editor'
import { DEMAND_TABLE_COLUMNS } from './components/demand-editor/constants'
import type { DemandFormProps } from './types'
import { DEMAND_LOADING_ROWS } from './constants'
import { demandsKey } from './utils'

/** The project's demand: budget hours per expertise and minimum level, in one phase or the whole project. */
export function DemandForm({ projectId }: DemandFormProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: project, isPending } = useProject(projectId)
  const { data: expertises = [], isPending: expertisesLoading } = useExpertises()

  if (isPending)
    return (
      <Section
        title={t('Demand')}
        description={t(
          'Budget hours per expertise and minimum level. Without a phase, a demand spans the whole project.',
        )}
      >
        <Table>
          <TableBody>
            <TableLoading columns={DEMAND_TABLE_COLUMNS} rows={DEMAND_LOADING_ROWS} />
          </TableBody>
        </Table>
      </Section>
    )
  if (!project) return null

  return (
    <DemandEditor
      key={demandsKey(project.demands)}
      project={project}
      expertises={expertises}
      expertisesLoading={expertisesLoading}
      editable={canEditProject(user, project)}
    />
  )
}
