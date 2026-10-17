import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/shared/ui/tooltip'
import type { OverloadBadgeProps } from './types'

/** Marks an expertise whose people are over capacity for weeks in a row; the tooltip says from which week. */
export function OverloadBadge({ run }: OverloadBadgeProps) {
  const { t } = useTranslation()

  return (
    <Tooltip>
      <TooltipTrigger render={<Badge variant="red" tabIndex={0} />}>
        {t('Overloaded {count} weeks in a row', { count: run.weeks })}
      </TooltipTrigger>
      <TooltipContent>{t('From {week}, within the weeks shown', { week: run.from_week })}</TooltipContent>
    </Tooltip>
  )
}
