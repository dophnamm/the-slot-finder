import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/shared/ui/card'
import { Progress, ProgressLabel, ProgressValue } from '@/shared/ui/progress'
import { MEMBER_INITIALS, PROGRESS, TIMELINE } from '../../mock-data'
import type { ProgramHeaderProps } from './types'

export function ProgramHeader({ programId }: ProgramHeaderProps) {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          {t('Program {id}', { id: programId })}
          <Badge variant="secondary">{t('Open')}</Badge>
        </CardTitle>
        <CardDescription>
          {t('Timeline: {start} – {end} ({dates})', {
            start: TIMELINE.start,
            end: TIMELINE.end,
            dates: TIMELINE.dates,
          })}
          {' · '}
          {t('Next milestone: {name}, {week}', {
            name: TIMELINE.milestone,
            week: TIMELINE.milestoneWeek,
          })}
        </CardDescription>
        <CardAction className="flex items-center gap-2 text-sm text-muted-foreground">
          {t('Members')}
          <div className="flex -space-x-1.5">
            {MEMBER_INITIALS.map((initials) => (
              <span
                key={initials}
                className="flex size-8 items-center justify-center rounded-full bg-secondary text-xs font-medium text-secondary-foreground ring-2 ring-card"
              >
                {initials}
              </span>
            ))}
          </div>
        </CardAction>
      </CardHeader>
      <CardContent className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {PROGRESS.map(({ label, value }) => (
          <Progress key={label} value={value}>
            <ProgressLabel>{t(label)}</ProgressLabel>
            <ProgressValue />
          </Progress>
        ))}
      </CardContent>
    </Card>
  )
}
