import type { BadgeVariant } from '@/shared/ui/badge'

/** Badge colour per status: new work is violet, planned blue, running green, paused amber, finished gray. */
export const STATUS_VARIANT: Record<string, BadgeVariant> = {
  open: 'blue',
  incoming: 'violet',
  planned: 'blue',
  active: 'green',
  pending: 'amber',
  on_hold: 'amber',
  closing: 'teal',
  closed: 'gray',
  done: 'gray',
  cancelled: 'red',
}
