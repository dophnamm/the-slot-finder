import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { activeWarnings } from '@/shared/programs'
import { Badge } from '@/shared/ui/badge'
import type { RiskProjectProps } from './types'

export function RiskProject({ project }: RiskProjectProps) {
  const { t } = useTranslation()
  const { metrics } = project

  return (
    <li className="flex flex-col gap-1 rounded-md border p-3">
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
            {t(warning.label)}
          </Badge>
        ))}
        {metrics.understaffed && (
          <span className="text-xs text-muted-foreground">
            {t('{gap} h short of {budget} h', { gap: Math.round(metrics.gap_hours), budget: Math.round(metrics.budget_hours) })}
          </span>
        )}
      </div>
    </li>
  )
}
