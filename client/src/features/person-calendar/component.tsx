import { useState } from 'react'
import { Section } from '@/shared/components/section'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { useSchedule } from '@/shared/people'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { currentMonth, formatIsoWeekday, formatMonth, monthIsoWeeks } from '@/utils/functions'
import { MonthNav } from './components/month-nav'
import { WeekDetail } from './components/week-detail'
import { WeekRow } from './components/week-row'
import { CALENDAR_LABEL_COLUMNS } from './constants'
import type { PersonCalendarProps } from './types'
import { daysByDate, findWeek, orderedWeekdays, weekCells } from './utils'

/** One person's month: absences and holidays by day, load against capacity and hours per project by week. */
export function PersonCalendar({ personId }: PersonCalendarProps) {
  const { t } = useTranslation()
  const [month, setMonth] = useState(currentMonth)
  const [selectedWeek, setSelectedWeek] = useState<string | null>(null)
  const weeks = monthIsoWeeks(month)
  const { data: schedule, isPending } = useSchedule(personId, weeks[0], weeks[weeks.length - 1])
  const weekdays = orderedWeekdays(schedule?.working_weekdays ?? [])
  const days = daysByDate(schedule?.days ?? [])
  const selected = findWeek(schedule?.weeks ?? [], selectedWeek)

  const handleMonthChange = (next: string) => {
    setMonth(next)
    setSelectedWeek(null)
  }

  return (
    <Section
      title={formatMonth(month)}
      description={t('Hours by week against capacity, with leave and holidays by day.')}
      action={<MonthNav month={month} onChange={handleMonthChange} />}
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Week')}</TableHead>
            {weekdays.map((weekday) => (
              <TableHead key={weekday}>{formatIsoWeekday(weekday)}</TableHead>
            ))}
            <TableHead className="text-right">{t('Load')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {schedule?.weeks.map((week) => (
            <WeekRow
              key={week.week}
              week={week}
              days={weekCells(week.week, weekdays, days)}
              month={month}
              selected={week.week === selectedWeek}
              onSelect={setSelectedWeek}
            />
          ))}
          {isPending && <TableLoading columns={weekdays.length + CALENDAR_LABEL_COLUMNS} />}
        </TableBody>
      </Table>
      {selected && <WeekDetail week={selected} />}
    </Section>
  )
}
