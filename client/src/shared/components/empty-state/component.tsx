import { cn } from '@/utils/functions'
import type { EmptyStateProps } from './types'

/** What to show where a list has nothing yet: an icon, what is missing and, optionally, what to do. */
export function EmptyState({ icon: Icon, title, description, action, framed }: EmptyStateProps) {
  return (
    <div
      className={cn(
        'flex flex-col items-center gap-2 px-4 py-10 text-center whitespace-normal',
        framed && 'rounded-lg border',
      )}
    >
      <div className="flex size-10 items-center justify-center rounded-full bg-muted">
        <Icon className="size-5 text-muted-foreground" aria-hidden />
      </div>
      <p className="font-medium">{title}</p>
      {description && <p className="max-w-md text-sm text-muted-foreground">{description}</p>}
      {action}
    </div>
  )
}
