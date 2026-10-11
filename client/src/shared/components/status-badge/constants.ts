import type { badgeVariants } from '@/shared/ui/badge'
import type { VariantProps } from 'class-variance-authority'

type BadgeVariant = NonNullable<VariantProps<typeof badgeVariants>['variant']>

/** Badge look per status: running work stands out, finished work fades. */
export const STATUS_VARIANT: Record<string, BadgeVariant> = {
  open: 'default',
  incoming: 'secondary',
  planned: 'secondary',
  active: 'default',
  pending: 'outline',
  on_hold: 'outline',
  closing: 'secondary',
  closed: 'outline',
  done: 'outline',
  cancelled: 'destructive',
}
