import { useTranslation } from '@/shared/i18n'
import { cn } from '@/utils/functions'
import {
  CELL_CLASSES,
  CURRENT_WEEK_CELL_CLASS,
  DEADLINE_CELL_CLASS,
  LABEL_COLUMN_CLASS,
  WEEK_COLUMN_CLASS,
} from '../../constants'
import type { TimelineRowProps } from './types'

export function TimelineRow({ row, weeks, deadlineWeek, currentWeek }: TimelineRowProps) {
  const { t } = useTranslation()

  return (
    <div className="flex items-center gap-px">
      <span className={cn(LABEL_COLUMN_CLASS, 'text-sm')}>{row.name || t('New phase')}</span>
      {row.cells.map((state, index) => (
        <span
          key={weeks[index]}
          data-current-week={weeks[index] === currentWeek || undefined}
          className={cn(
            WEEK_COLUMN_CLASS,
            'h-6 rounded-sm',
            CELL_CLASSES[state],
            weeks[index] === deadlineWeek && DEADLINE_CELL_CLASS,
            weeks[index] === currentWeek && CURRENT_WEEK_CELL_CLASS,
          )}
        />
      ))}
    </div>
  )
}
