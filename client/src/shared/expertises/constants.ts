import type { BadgeVariant } from '@/shared/ui/badge'

/** Badge colours handed out to expertises in catalogue order. */
export const EXPERTISE_BADGES: readonly BadgeVariant[] = ['blue', 'violet', 'teal', 'pink', 'amber', 'green']

/** Badge colour of an expertise missing from the catalogue. */
export const UNKNOWN_EXPERTISE_BADGE: BadgeVariant = 'gray'
