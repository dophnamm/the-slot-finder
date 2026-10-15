import { useTranslation } from '@/shared/i18n'
import { SHIFT_PROJECT } from '@/shared/scenarios'
import { formatDateRange } from '@/utils/functions'
import type { AssumptionTextProps } from './types'
import { isLater, shiftSize } from './utils'

/** One assumption in words, e.g. "Checkout Revamp 2 weeks later" or "Alex Nguyen away Oct 26 – Nov 6, 8 hrs a day". */
export function AssumptionText({ assumption }: AssumptionTextProps) {
  const { t } = useTranslation()

  if (assumption.type === SHIFT_PROJECT) {
    const project = assumption.project || t('A project')
    const weeks = shiftSize(assumption)
    if (isLater(assumption)) {
      return (
        <span>
          {weeks === 1
            ? t('{project} 1 week later', { project })
            : t('{project} {weeks} weeks later', { project, weeks })}
        </span>
      )
    }
    return (
      <span>
        {weeks === 1
          ? t('{project} 1 week earlier', { project })
          : t('{project} {weeks} weeks earlier', { project, weeks })}
      </span>
    )
  }
  return (
    <span>
      {t('{person} away {dates}, {hours} hrs a day', {
        person: assumption.person || t('Someone'),
        dates: formatDateRange(assumption.from_date ?? '', assumption.to_date ?? ''),
        hours: assumption.hours_per_day ?? 0,
      })}
    </span>
  )
}
