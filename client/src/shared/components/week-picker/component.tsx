import { CalendarRange } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Calendar } from '@/shared/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/ui/popover'
import { cn, formatIsoWeekRange, isIsoWeek, toIsoWeek } from '@/utils/functions'
import type { WeekPickerProps } from './types'
import { weekDays } from './utils'

/** An ISO week field: a calendar with week numbers where any day picks its whole week. */
export function WeekPicker({ id, name, value, onChange, min, className, 'aria-label': ariaLabel }: WeekPickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const selected = weekDays(value)
  const earliest = min ? weekDays(min)?.from : undefined

  const handleSelect = (_range: unknown, day: Date) => {
    onChange(toIsoWeek(day), name)
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button id={id} variant="outline" aria-label={ariaLabel}
            className={cn('justify-start font-normal', !value && 'text-muted-foreground', className)} />
        }
      >
        <CalendarRange />
        {isIsoWeek(value) ? t('{week} · {range}', { week: value, range: formatIsoWeekRange(value) }) : t('Pick a week')}
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar mode="range" ISOWeek showWeekNumber selected={selected} defaultMonth={selected?.from ?? earliest}
          onSelect={handleSelect} disabled={earliest ? { before: earliest } : undefined} />
      </PopoverContent>
    </Popover>
  )
}
