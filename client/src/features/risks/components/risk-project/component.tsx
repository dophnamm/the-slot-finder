import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { Term } from '@/shared/components/term'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { activeWarnings } from '@/shared/programs'
import { Badge } from '@/shared/ui/badge'
import type { RiskProjectProps } from './types'

export function RiskProject({ project }: RiskProjectProps) {
  const { t } = useTranslation()
  const { metrics } = project

  return (
    <li className="flex flex-col gap-1 py-3 first:pt-0 last:pb-0">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <TextLink to={generatePath(ROUTES.projectDetail, { programId: project.program_id, id: project.id })}>
          {project.name}
        </TextLink>
        <span className="text-xs text-muted-foreground">{t('Deadline {week}', { week: project.deadline_week })}</span>
      </div>
      <span className="text-xs text-muted-foreground">{project.program}</span>
      <div className="flex flex-wrap items-center gap-1">
        {activeWarnings(metrics).map((warning) => (
          <Badge key={warning.flag} variant={warning.badge}>
            <Term term={warning.term}>{t(warning.label)}</Term>
          </Badge>
        ))}
        {metrics.understaffed && (
          <span className="text-xs text-muted-foreground">
            {t('{gap} hrs short of {budget} hrs', {
              gap: Math.round(metrics.gap_hours),
              budget: Math.round(metrics.budget_hours),
            })}
          </span>
        )}
      </div>
    </li>
  )
}
