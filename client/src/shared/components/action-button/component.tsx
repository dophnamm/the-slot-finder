import { Button } from '@/shared/ui/button'
import type { ActionButtonProps } from './types'

/** A status action such as hold or cancel; irreversible ones ask for confirmation. */
export function ActionButton<A extends string>({
  action,
  label,
  destructive,
  confirmText,
  disabled,
  onAction,
}: ActionButtonProps<A>) {
  const handleClick = () => {
    if (!confirmText || window.confirm(confirmText)) onAction(action)
  }

  return (
    <Button variant={destructive ? 'destructive' : 'outline'} size="sm" disabled={disabled} onClick={handleClick}>
      {label}
    </Button>
  )
}
