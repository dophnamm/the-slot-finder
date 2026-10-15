import { UserX } from 'lucide-react'
import { EmptyState } from '@/shared/components/empty-state'
import { useTranslation } from '@/shared/i18n'
import { useAlternatives } from '@/shared/staffing'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { Skeleton } from '@/shared/ui/skeleton'
import { AlternativeRow } from '../alternative-row'
import { ALTERNATIVES_LOADING_ROWS } from './constants'
import type { SwapPersonDialogProps } from './types'

/** Lists who could take over an assignment, with their free hours; choosing one moves the hours to them. */
export function SwapPersonDialog({ draftId, assignment, busy, onChoose, onCancel }: SwapPersonDialogProps) {
  const { t } = useTranslation()
  const { data: alternatives = [], isPending } = useAlternatives(draftId, assignment.id)

  const handleOpenChange = (isOpen: boolean) => {
    if (!isOpen) onCancel()
  }

  return (
    <Dialog open onOpenChange={handleOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('Swap person')}</DialogTitle>
          <DialogDescription>
            {t('Moves the hours of {name} on {expertise} to someone from the program team.', {
              name: assignment.person.name,
              expertise: assignment.demand.expertise,
            })}
          </DialogDescription>
        </DialogHeader>
        <ul className="flex flex-col gap-2">
          {isPending &&
            Array.from({ length: ALTERNATIVES_LOADING_ROWS }, (_, index) => (
              <li key={index}>
                <Skeleton className="h-14" />
              </li>
            ))}
          {alternatives.map((alternative) => (
            <AlternativeRow key={alternative.person.id} alternative={alternative} busy={busy} onChoose={onChoose} />
          ))}
        </ul>
        {!isPending && alternatives.length === 0 && (
          <EmptyState
            framed
            icon={UserX}
            title={t('Nobody else qualifies.')}
            description={t('The program team has nobody else with this expertise at this level.')}
          />
        )}
        <DialogFooter>
          <Button type="button" variant="ghost" disabled={busy} onClick={onCancel}>
            {t('Cancel')}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
