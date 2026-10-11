import { PROGRAM_ACTIONS } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { CONFIRMED_ACTIONS } from './constants'
import type { ProgramActionButtonProps } from './types'

export function ProgramActionButton({ action, disabled, onAction }: ProgramActionButtonProps) {
  const { t } = useTranslation()
  const confirmed = CONFIRMED_ACTIONS.includes(action)

  const handleClick = () => {
    if (!confirmed || window.confirm(t('This cannot be undone. Open projects of a cancelled program are cancelled too. Continue?'))) {
      onAction(action)
    }
  }

  return (
    <Button variant={action === 'cancel' ? 'destructive' : 'outline'} size="sm" disabled={disabled} onClick={handleClick}>
      {t(PROGRAM_ACTIONS[action].label)}
    </Button>
  )
}
