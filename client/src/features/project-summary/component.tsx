import { useState } from 'react'
import { generatePath } from 'react-router'
import { toast } from 'sonner'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import { ActionButton } from '@/shared/components/action-button'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { ProjectFormDialog } from '@/shared/components/project-form-dialog'
import { StatusBadge } from '@/shared/components/status-badge'
import { Section } from '@/shared/components/section'
import { TextLink } from '@/shared/components/text-link'
import {
  availableActions,
  canEdit,
  isOpenStatus,
  PROJECT_ACTIONS,
  type ProjectAction,
  useProject,
  useProjectAction,
} from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { formatDate } from '@/utils/functions'
import { Fact } from './components/fact'
import { CONFIRMED_PROJECT_ACTIONS } from './constants'
import type { ProjectSummaryProps } from './types'

/** Project header: status, risk, program, dates, points, lead, edit and status actions. */
export function ProjectSummary({ projectId }: ProjectSummaryProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const { data: project, isPending } = useProject(projectId)
  const action = useProjectAction(projectId)
  const [editing, setEditing] = useState(false)

  const handleEdit = () => setEditing(true)
  const handleDone = () => toast.success(t('Project updated.'))
  const handleAction = (next: ProjectAction) => action.mutate(next, { onSuccess: handleDone })

  if (isPending) return null
  if (!project)
    return <PagePlaceholder title={t('Project not found')} description={t('There is no project at this address.')} />

  const editable = canEdit(user, [project.lead, project.program_lead])
  const actions = editable ? availableActions(PROJECT_ACTIONS, project.status) : []

  return (
    <Section
      card
      title={
        <>
          {project.name}
          <StatusBadge status={project.status} kind="project" />
          {project.at_risk && <Badge variant="red">{t('At risk')}</Badge>}
        </>
      }
      description={t('Priority {priority}', { priority: project.priority })}
      action={
        editable && (
          <>
            {isOpenStatus(project.status) && (
              <Button variant="outline" size="sm" onClick={handleEdit}>
                {t('Edit')}
              </Button>
            )}
            {actions.map((name) => (
              <ActionButton
                key={name}
                action={name}
                label={t(PROJECT_ACTIONS[name].label)}
                destructive={name === 'cancel'}
                confirmText={
                  CONFIRMED_PROJECT_ACTIONS.includes(name) ? t('This cannot be undone. Continue?') : undefined
                }
                disabled={action.isPending}
                onAction={handleAction}
              />
            ))}
          </>
        )
      }
      variant="page"
      contentClassName="flex flex-col gap-4"
    >
      <dl className="grid gap-4 sm:grid-cols-3">
        <Fact label={t('Program')}>
          <TextLink to={generatePath(ROUTES.programDetail, { id: project.program_id })}>
            {project.program_name ?? t('Program')}
          </TextLink>
        </Fact>
        <Fact label={t('Lead')}>{project.lead?.name ?? t('No lead')}</Fact>
        <Fact label={t('Scope')}>
          {t('{done} of {scope} points done', { done: project.done_points, scope: project.scope_points })}
        </Fact>
        <Fact label={t('Earliest start')}>
          {t('{date} ({week})', { date: formatDate(project.start_earliest), week: project.start_week })}
        </Fact>
        <Fact label={t('Deadline')}>
          {t('{date} ({week})', { date: formatDate(project.deadline), week: project.deadline_week })}
        </Fact>
        <Fact label={t('Current phase')}>{project.current_phase?.name ?? t('None this week')}</Fact>
      </dl>
      {project.at_risk && (
        <p className="text-sm text-destructive">
          {t('The last phase ends after the deadline week. Move the deadline or shorten a phase.')}
        </p>
      )}
      {editing && (
        <ProjectFormDialog open={editing} programId={project.program_id} project={project} onOpenChange={setEditing} />
      )}
    </Section>
  )
}
