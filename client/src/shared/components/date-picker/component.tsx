import { CalendarDays } from 'lucide-react'
import { useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Calendar } from '@/shared/ui/calendar'
import { Popover, PopoverContent, PopoverTrigger } from '@/shared/ui/popover'
import { cn, formatDate, parseDateValue, toDateValue } from '@/utils/functions'
import type { DatePickerProps } from './types'

/** A date field with the app's calendar in a popover, in place of the browser's own date picker. */
export function DatePicker({ id, name, value, onChange, min, disabled, 'aria-label': ariaLabel }: DatePickerProps) {
  const { t } = useTranslation()
  const [open, setOpen] = useState(false)
  const selected = parseDateValue(value)
  const earliest = min ? parseDateValue(min) : undefined

  const handleSelect = (date: Date | undefined) => {
    onChange(date ? toDateValue(date) : '', name)
    setOpen(false)
  }

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger
        render={
          <Button
            id={id}
            variant="outline"
            disabled={disabled}
            aria-label={ariaLabel}
            className={cn('w-full justify-start font-normal', !value && 'text-muted-foreground')}
          />
        }
      >
        <CalendarDays />
        {value ? formatDate(value) : t('Pick a date')}
      </PopoverTrigger>
      <PopoverContent className="w-auto p-0" align="start">
        <Calendar
          mode="single"
          weekStartsOn={1}
          selected={selected}
          defaultMonth={selected ?? earliest}
          onSelect={handleSelect}
          disabled={earliest ? { before: earliest } : undefined}
        />
      </PopoverContent>
    </Popover>
  )
}
