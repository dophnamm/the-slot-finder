import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { AbsenceForm } from '../absence-form'
import type { AddAbsenceProps } from './types'

/** The "Add absence" button for a card header; the form opens in a dialog and closes once saved. */
export function AddAbsence({ personId, canApprove }: AddAbsenceProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  const handleOpen = () => setOpen(true)
  const handleCreated = () => setOpen(false)

  return (
    <>
      <Button variant="outline" onClick={handleOpen}>
        <Plus />
        {t('Add absence')}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{t('Add absence')}</DialogTitle>
            <DialogDescription>
              {t('Approved absence always lowers capacity; planned absence does when the setting is on.')}
            </DialogDescription>
          </DialogHeader>
          <AbsenceForm personId={personId} canApprove={canApprove} onCreated={handleCreated} />
        </DialogContent>
      </Dialog>
    </>
  )
}
