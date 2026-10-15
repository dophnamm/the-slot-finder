import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/ui/tooltip'
import { ICON_HINT_DELAY_MS } from './constants'
import type { IconHintProps } from './types'

/** Names an icon-only button for people who rest the pointer on it. */
export function IconHint({ label, children }: IconHintProps) {
  return (
    <Tooltip>
      <TooltipTrigger delay={ICON_HINT_DELAY_MS} render={children} />
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  )
}
