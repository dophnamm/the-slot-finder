import { AbsenceList, AddAbsence } from '@/features/absences'
import { useTranslation } from '@/shared/i18n'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import type { PersonAbsencesDialogProps } from './types'

export function PersonAbsencesDialog({ person, canApprove, onClose }: PersonAbsencesDialogProps) {
  const { t } = useTranslation()

  const handleOpenChange = (open: boolean) => {
    if (!open) onClose()
  }

  return (
    <Dialog open={person !== null} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-3xl">
        <DialogHeader>
          <DialogTitle>{t('Absences of {name}', { name: person?.name ?? '' })}</DialogTitle>
          <DialogDescription>
            {t('Approved absence always lowers capacity; planned absence does when the setting is on.')}
          </DialogDescription>
        </DialogHeader>
        {person && (
          <div className="flex flex-col gap-3">
            <div className="flex justify-end">
              <AddAbsence personId={person.id} canApprove={canApprove} />
            </div>
            <AbsenceList personId={person.id} canApprove={canApprove} />
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
