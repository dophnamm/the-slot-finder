import { RefreshCw } from 'lucide-react'
import { generatePath, Link, useNavigate } from 'react-router'
import { toast } from 'sonner'
import { can } from '@/resources/rbac'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import { EventSummary } from '@/shared/components/event-summary'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import {
  type Draft,
  PLAN_STATUS_BADGE, PLAN_STATUS_LABELS,
  PUBLISHED_PLAN_STATUS,
  useAcceptDraft,
  useGenerateDraft,
  useRejectDraft,
} from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { formatDateTime } from '@/utils/functions'
import type { DraftHeaderProps } from './types'

/** Draft status, the baseline it was built on, the projects it covers, and accept or reject. */
export function DraftHeader({ draft, editable }: DraftHeaderProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: user } = useCurrentUser()
  const accept = useAcceptDraft(draft.id)
  const reject = useRejectDraft(draft.id)
  const generate = useGenerateDraft()
  const busy = accept.isPending || reject.isPending || generate.isPending
  const canRegenerate = draft.stale && can(user?.role, 'staffing-drafts:manage')

  const handleAccepted = () => toast.success(t('Draft accepted. It is now the baseline.'))
  const handleRejected = () => toast.success(t('Draft rejected.'))
  const handleOpen = (next: Draft) => navigate(generatePath(ROUTES.staffingDraft, { id: next.id }))
  const handleAccept = () => {
    if (!window.confirm(t('Accept this staffing draft? It becomes the baseline everyone plans against.'))) return
    const confirmStretch = draft.has_stretch
    if (confirmStretch && !window.confirm(t('This draft has stretch assignments: people below the minimum level or without the skill. Accept them anyway?'))) return
    accept.mutate({ base_version: draft.base_version, confirm_stretch: confirmStretch }, { onSuccess: handleAccepted })
  }
  const handleReject = () => {
    if (!window.confirm(t('Reject this staffing draft?'))) return
    reject.mutate(undefined, { onSuccess: handleRejected })
  }
  const handleRegenerate = () => generate.mutate(draft.projects.map((project) => project.id), { onSuccess: handleOpen })

  return (
    <Section
      card
      title={
        <>
          {t('Staffing draft')}
          <Badge variant={PLAN_STATUS_BADGE[draft.status]}>{t(PLAN_STATUS_LABELS[draft.status])}</Badge>
          {draft.auto && <Badge variant="violet">{t('Re-plan')}</Badge>}
          {draft.has_stretch && <Badge variant="red">{t('Has stretch')}</Badge>}
        </>
      }
      description={
        draft.status === PUBLISHED_PLAN_STATUS
          ? t('Accepted as baseline version {version}.', { version: draft.base_version })
          : t('Built on baseline version {version}.', { version: draft.base_version })
      }
      action={
        editable &&
        !draft.stale && (
          <>
            <Button variant="outline" size="sm" disabled={busy} onClick={handleReject}>
              {t('Reject')}
            </Button>
            <Button size="sm" disabled={busy} onClick={handleAccept}>
              {t('Accept draft')}
            </Button>
          </>
        )
      }
      variant="page"
      contentClassName="flex flex-col gap-3"
    >
      <div className="flex flex-wrap gap-x-4 gap-y-1 text-sm">
        <span className="text-muted-foreground">{t('Covers')}</span>
        {draft.projects.map((project) => (
          <Link key={project.id} className="font-medium hover:underline" to={generatePath(ROUTES.projectDetail, { id: project.id })}>
            {t('{project} ({program})', { project: project.name, program: project.program })}
          </Link>
        ))}
      </div>
      {draft.superseded_by && (
        <div className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-warning p-3">
          <p className="text-sm">{t('A change produced a newer version of this draft.')}</p>
          <Button variant="outline" size="sm" nativeButton={false} render={<Link to={generatePath(ROUTES.staffingDraft, { id: draft.superseded_by })} />}>
            {t('Open the latest version')}
          </Button>
        </div>
      )}
      {draft.auto && (
        <div className="flex flex-col gap-1 text-sm">
          <span className="text-muted-foreground">{t('Triggered by')}</span>
          <ul className="flex flex-col gap-1">
            {draft.triggered_by.map((event) => (
              <li key={event.id} className="flex flex-wrap gap-x-2">
                <span className="whitespace-nowrap text-muted-foreground">{formatDateTime(event.occurred_at)}</span>
                <EventSummary event={event} />
              </li>
            ))}
          </ul>
        </div>
      )}
      {draft.stale && (
        <div className="flex flex-wrap items-center justify-between gap-2 rounded-md border border-destructive/40 p-3">
          <p className="text-sm text-destructive">
            {t('The baseline changed to version {version} since this draft was made, so it can no longer be accepted.', {
              version: draft.current_version,
            })}
          </p>
          {canRegenerate && (
            <Button variant="outline" size="sm" disabled={busy} onClick={handleRegenerate}>
              <RefreshCw />
              {t('Generate a new draft')}
            </Button>
          )}
        </div>
      )}
    </Section>
  )
}
