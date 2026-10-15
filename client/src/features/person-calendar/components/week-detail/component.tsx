import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { BAND_LABELS } from '@/shared/utilization'
import { formatIsoWeekRange } from '@/utils/functions'
import type { WeekDetailProps } from './types'
import { projectPath } from './utils'

/** The selected week: load against capacity, then the hours per project, each linking to its project. */
export function WeekDetail({ week, canOpenProjects }: WeekDetailProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-3 rounded-lg border p-4 text-sm">
      <div className="flex flex-col">
        <span className="font-medium">
          {t('{week} ({dates})', { week: week.week, dates: formatIsoWeekRange(week.week) })}
        </span>
        <span className="text-muted-foreground">
          {t('Planned {load} of {capacity} hrs · {band}', {
            load: week.load,
            capacity: week.capacity,
            band: t(BAND_LABELS[week.band]),
          })}
        </span>
      </div>
      {week.tentative && (
        <span className="text-xs text-muted-foreground">{t('Planned leave this week, not approved yet.')}</span>
      )}
      {week.projects.length === 0 && <span className="text-muted-foreground">{t('No assignments this week.')}</span>}
      <ul className="flex flex-col gap-2">
        {week.projects.map((project) => (
          <li key={`${project.project_id}-${project.expertise}`} className="flex items-start justify-between gap-2">
            <div className="flex min-w-0 flex-col">
              {canOpenProjects ? (
                <TextLink to={projectPath(project)}>{project.project_name}</TextLink>
              ) : (
                <span className="font-medium">{project.project_name}</span>
              )}
              <span className="text-xs text-muted-foreground">
                {t('{expertise} · {program}', { expertise: project.expertise, program: project.program_name })}
              </span>
            </div>
            <span className="shrink-0">{t('{hours} hrs', { hours: project.hours })}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}
