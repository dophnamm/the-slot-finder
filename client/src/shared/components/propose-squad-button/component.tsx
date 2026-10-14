import { Users } from 'lucide-react'
import { generatePath, useNavigate } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import { useTranslation } from '@/shared/i18n'
import { canStaff, type Draft, PLANNABLE_STATUSES, useGenerateDraft } from '@/shared/staffing'
import { Button } from '@/shared/ui/button'
import type { ProposeSquadButtonProps } from './types'

/** Asks the planning engine for a staffing draft for one project and opens it (planners and the program's manager). */
export function ProposeSquadButton({ projectId, status, programLeadId }: ProposeSquadButtonProps) {
  const { t } = useTranslation()
  const navigate = useNavigate()
  const { data: user } = useCurrentUser()
  const generate = useGenerateDraft()

  const handleOpen = (draft: Draft) => navigate(generatePath(ROUTES.staffingDraft, { id: draft.id }))
  const handleClick = () => generate.mutate([projectId], { onSuccess: handleOpen })

  if (!canStaff(user, [programLeadId]) || !PLANNABLE_STATUSES.includes(status)) return null

  return (
    <Button variant="outline" size="sm" disabled={generate.isPending} onClick={handleClick}>
      <Users />
      {generate.isPending ? t('Proposing…') : t('Propose squad')}
    </Button>
  )
}
