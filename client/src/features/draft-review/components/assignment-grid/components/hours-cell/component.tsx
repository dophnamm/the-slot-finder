import { Lock } from 'lucide-react'
import { type ChangeEvent, type KeyboardEvent, useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Input } from '@/shared/ui/input'
import { TableCell } from '@/shared/ui/table'
import { cn } from '@/utils/functions'
import { WEEK_COLUMN_CLASS } from '../../constants'
import { MAX_WEEK_HOURS } from './constants'
import type { HoursCellProps } from './types'

/** One week of an assignment: the hours as text, or an input that saves on blur or Enter. */
export function HoursCell({ week, hours, locked, inWindow, editable, onCommit }: HoursCellProps) {
  const { t } = useTranslation()
  const [value, setValue] = useState(hours ? String(hours) : '')

  const handleChange = (event: ChangeEvent<HTMLInputElement>) => setValue(event.target.value)
  const handleBlur = () => {
    const next = value.trim() === '' ? 0 : Number(value)
    if (Number.isNaN(next) || next < 0 || next > MAX_WEEK_HOURS || next === hours) {
      setValue(hours ? String(hours) : '')
      return
    }
    onCommit(week, next)
  }
  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter') event.currentTarget.blur()
  }

  if (!inWindow) return <TableCell className={cn(WEEK_COLUMN_CLASS, 'bg-muted/40')} />

  if (!editable) {
    return (
      <TableCell className={cn(WEEK_COLUMN_CLASS, 'text-sm', locked && 'font-medium text-primary')}>
        <span className="inline-flex items-center gap-1">
          {hours > 0 && hours}
          {locked && <Lock className="size-3" />}
        </span>
      </TableCell>
    )
  }

  return (
    <TableCell className={cn(WEEK_COLUMN_CLASS, 'p-1')}>
      <Input
        type="number"
        inputMode="decimal"
        min={0}
        max={MAX_WEEK_HOURS}
        step="any"
        value={value}
        aria-label={t('Hours in {week}', { week })}
        className={cn('mx-auto h-8 w-14 bg-card text-center', locked && 'border-primary')}
        onChange={handleChange}
        onBlur={handleBlur}
        onKeyDown={handleKeyDown}
      />
    </TableCell>
  )
}
