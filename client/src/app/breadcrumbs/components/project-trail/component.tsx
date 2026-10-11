import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'
import { useProject } from '@/shared/programs'
import { Trail } from '../trail'
import type { ProjectTrailProps } from './types'

/** Programs › the project's program › the project's name. */
export function ProjectTrail({ projectId }: ProjectTrailProps) {
  const { t } = useTranslation()
  const { data: project } = useProject(projectId)

  if (!project) return <Trail crumbs={[{ label: t('Programs'), to: ROUTES.programs }, { label: t('Loading…') }]} />
  return (
    <Trail
      crumbs={[
        { label: t('Programs'), to: ROUTES.programs },
        { label: project.program_name ?? t('Program'), to: generatePath(ROUTES.programDetail, { id: project.program_id }) },
        { label: project.name },
      ]}
    />
  )
}
