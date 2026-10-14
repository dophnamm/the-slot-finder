import { Plus } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle } from '@/shared/ui/dialog'
import { HolidayForm } from '../holiday-form'

/** The "Add holiday" button for the card header; the form opens in a dialog and closes once saved. */
export function AddHoliday() {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)

  const handleOpen = () => setOpen(true)
  const handleCreated = () => setOpen(false)

  return (
    <>
      <Button variant="outline" onClick={handleOpen}>
        <Plus />
        {t('Add holiday')}
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{t('Add holiday')}</DialogTitle>
            <DialogDescription>{t('Holidays lower everyone’s capacity on those days.')}</DialogDescription>
          </DialogHeader>
          <HolidayForm onCreated={handleCreated} />
        </DialogContent>
      </Dialog>
    </>
  )
}
