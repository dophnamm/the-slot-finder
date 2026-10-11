import type { BadgeVariant } from '@/shared/ui/badge'
import { EXPERTISE_BADGES, UNKNOWN_EXPERTISE_BADGE } from './constants'
import { useExpertises } from './use-expertises'

/** One colour per expertise, the same on every screen: picked by its place in the catalogue. */
export function useExpertiseBadge() {
  const { data: expertises = [] } = useExpertises()
  return (expertiseId: string): BadgeVariant => {
    const index = expertises.findIndex((expertise) => expertise.id === expertiseId)
    return index < 0 ? UNKNOWN_EXPERTISE_BADGE : EXPERTISE_BADGES[index % EXPERTISE_BADGES.length]
  }
}
