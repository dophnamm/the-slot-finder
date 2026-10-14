import { generatePath } from 'react-router'
import { PROGRAM_TABS, ROUTES, tabPath } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'
import { useProject } from '@/shared/programs'
import { Trail } from '../trail'
import type { ProjectTrailProps } from './types'

/** Programs › the project's program › its Projects tab › the project's name. */
export function ProjectTrail({ projectId }: ProjectTrailProps) {
  const { t } = useTranslation()
  const { data: project } = useProject(projectId)

  if (!project) return <Trail crumbs={[{ label: t('Programs'), to: ROUTES.programs }, { label: t('Loading…') }]} />
  const programPath = generatePath(ROUTES.programDetail, { id: project.program_id })
  return (
    <Trail
      crumbs={[
        { label: t('Programs'), to: ROUTES.programs },
        { label: project.program_name ?? t('Program'), to: programPath },
        { label: t('Projects'), to: tabPath(programPath, PROGRAM_TABS.projects) },
        { label: project.name },
      ]}
    />
  )
}
