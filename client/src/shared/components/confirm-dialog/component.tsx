import { useTranslation } from '@/shared/i18n'
import { useConfirmStore } from '@/shared/stores/confirm-store'
import { Button } from '@/shared/ui/button'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/shared/ui/dialog'

/** The one dialog behind `confirm()`: mounted once at the app root, it shows the question and answers it. */
export function ConfirmDialog() {
  const { t } = useTranslation()
  const request = useConfirmStore((state) => state.request)
  const settle = useConfirmStore((state) => state.settle)

  const handleConfirm = () => settle(true)
  const handleCancel = () => settle(false)
  const handleOpenChange = (open: boolean) => {
    if (!open) settle(false)
  }

  return (
    <Dialog open={request !== null} onOpenChange={handleOpenChange}>
      {request && (
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{request.title}</DialogTitle>
            {request.description && <DialogDescription>{request.description}</DialogDescription>}
          </DialogHeader>
          {request.warning && <p className="text-sm text-destructive">{request.warning}</p>}
          <DialogFooter>
            <Button type="button" variant="ghost" onClick={handleCancel}>
              {t('Cancel')}
            </Button>
            <Button type="button" variant={request.destructive ? 'destructive' : 'default'} onClick={handleConfirm}>
              {request.confirmLabel}
            </Button>
          </DialogFooter>
        </DialogContent>
      )}
    </Dialog>
  )
}
