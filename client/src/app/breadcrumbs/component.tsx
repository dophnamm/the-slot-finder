import { useMatches, useParams } from 'react-router'
import { useTranslation } from '@/shared/i18n'
import { ProgramTrail } from './components/program-trail'
import { ProjectTrail } from './components/project-trail'
import { Trail } from './components/trail'
import { STATIC_TRAILS } from './constants'
import { currentBreadcrumb } from './utils'

/** Where the user is, above the page: fixed texts, or the program and project names from the server. */
export function Breadcrumbs() {
  const { t } = useTranslation()
  const kind = currentBreadcrumb(useMatches())
  const { id = '' } = useParams()

  if (kind === 'programDetail') return <ProgramTrail programId={id} />
  if (kind === 'projectDetail') return <ProjectTrail projectId={id} />
  const trail = kind ? STATIC_TRAILS[kind] : undefined
  if (!trail) return null
  return <Trail crumbs={trail.map((crumb) => ({ label: t(crumb.label), to: crumb.to }))} />
}
