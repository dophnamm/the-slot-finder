import { Info } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/ui/tooltip'
import { TERM_DEFINITIONS } from './constants'
import type { TermProps } from './types'

/** A word of our jargon followed by an info icon; the definition shows on hover or keyboard focus. */
export function Term({ term, children }: TermProps) {
  const { t } = useTranslation()

  if (!term) return children

  return (
    <span className="inline-flex items-center gap-1">
      {children}
      <Tooltip>
        <TooltipTrigger
          aria-label={t('What does this mean?')}
          className="inline-flex cursor-help rounded-full text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring"
        >
          <Info className="size-3.5" />
        </TooltipTrigger>
        <TooltipContent>{t(TERM_DEFINITIONS[term])}</TooltipContent>
      </Tooltip>
    </span>
  )
}
