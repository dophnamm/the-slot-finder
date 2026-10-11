import { useState } from 'react'
import { toast } from 'sonner'
import { useCurrentUser } from '@/shared/auth'
import { ActionButton } from '@/shared/components/action-button'
import { ProgramFormDialog } from '@/shared/components/program-form-dialog'
import { StatusBadge } from '@/shared/components/status-badge'
import { Section } from '@/shared/components/section'
import {
  availableActions,
  canEdit,
  isOpenStatus,
  PROGRAM_ACTIONS,
  type ProgramAction,
  useProgramAction,
  weekSpan,
} from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { CONFIRMED_PROGRAM_ACTIONS } from './constants'
import type { ProgramHeaderProps } from './types'

export function ProgramHeader({ program }: ProgramHeaderProps) {
  const { t } = useTranslation()
  const { data: user } = useCurrentUser()
  const [editing, setEditing] = useState(false)
  const action = useProgramAction(program.id)
  const editable = canEdit(user, [program.lead])
  const actions = editable ? availableActions(PROGRAM_ACTIONS, program.status) : []

  const handleEdit = () => setEditing(true)
  const handleDone = () => toast.success(t('Program updated.'))
  const handleAction = (next: ProgramAction) => action.mutate(next, { onSuccess: handleDone })

  return (
    <>
      <Section
        variant="page"
        title={
          <>
            {program.name}
            <StatusBadge status={program.status} kind="program" />
          </>
        }
        description={t('{client} · lead: {lead} · {timeline}', {
          client: program.client,
          lead: program.lead?.name ?? t('No lead'),
          timeline: weekSpan(program.start_week, program.end_week),
        })}
        action={
          editable && (
            <>
              {isOpenStatus(program.status) && (
                <Button variant="outline" size="sm" onClick={handleEdit}>
                  {t('Edit')}
                </Button>
              )}
              {actions.map((name) => (
                <ActionButton
                  key={name}
                  action={name}
                  label={t(PROGRAM_ACTIONS[name].label)}
                  destructive={name === 'cancel'}
                  confirmText={CONFIRMED_PROGRAM_ACTIONS.includes(name) ? t('This cannot be undone. Open projects of a cancelled program are cancelled too. Continue?') : undefined}
                  disabled={action.isPending}
                  onAction={handleAction}
                />
              ))}
            </>
          )
        }
      />
      {editing && <ProgramFormDialog open={editing} program={program} onOpenChange={setEditing} />}
    </>
  )
}
