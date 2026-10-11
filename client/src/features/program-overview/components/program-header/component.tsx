import { useState } from 'react'
import { toast } from 'sonner'
import { useCurrentUser } from '@/shared/auth'
import { ProgramFormDialog } from '@/shared/components/program-form-dialog'
import { StatusBadge } from '@/shared/components/status-badge'
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
import { Card, CardAction, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { ProgramActionButton } from '../program-action-button'
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
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          {program.name}
          <StatusBadge status={program.status} kind="program" />
        </CardTitle>
        <CardDescription>
          {t('{client} · lead: {lead} · {timeline}', {
            client: program.client,
            lead: program.lead?.name ?? t('No lead'),
            timeline: weekSpan(program.start_week, program.end_week),
          })}
        </CardDescription>
        {editable && (
          <CardAction className="flex flex-wrap justify-end gap-2">
            {isOpenStatus(program.status) && (
              <Button variant="outline" size="sm" onClick={handleEdit}>
                {t('Edit')}
              </Button>
            )}
            {actions.map((name) => (
              <ProgramActionButton key={name} action={name} disabled={action.isPending} onAction={handleAction} />
            ))}
          </CardAction>
        )}
      </CardHeader>
      {editing && <ProgramFormDialog open={editing} program={program} onOpenChange={setEditing} />}
    </Card>
  )
}
