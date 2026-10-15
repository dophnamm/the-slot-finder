import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { AssumptionText } from '@/shared/components/assumption-text'
import { LinkRow } from '@/shared/components/link-row'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { TableCell } from '@/shared/ui/table'
import { cn, formatDateTime } from '@/utils/functions'
import type { ScenarioRowProps } from './types'

/** One scenario: name and date, its assumptions in words, the baseline it was built on and its churn. */
export function ScenarioRow({ scenario }: ScenarioRowProps) {
  const { t } = useTranslation()
  const detailPath = generatePath(ROUTES.scenarioDetail, { id: scenario.id })
  const { summary } = scenario

  return (
    <LinkRow to={detailPath}>
      <TableCell>
        <div className="flex flex-col">
          <TextLink to={detailPath}>{scenario.name}</TextLink>
          <span className="text-xs text-muted-foreground">{formatDateTime(scenario.updated_at)}</span>
        </div>
      </TableCell>
      <TableCell className="whitespace-normal">
        <ul className="flex flex-col gap-0.5 text-sm">
          {scenario.assumptions.map((assumption, index) => (
            <li key={`${index}:${assumption.type}`}>
              <AssumptionText assumption={assumption} />
            </li>
          ))}
        </ul>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm">{t('Version {version}', { version: scenario.base_version })}</span>
          {scenario.stale && <Badge variant="amber">{t('Stale')}</Badge>}
        </div>
      </TableCell>
      <TableCell className="text-right tabular-nums">{summary.people_changed}</TableCell>
      <TableCell className="text-right tabular-nums">{t('{hours} hrs', { hours: summary.hours_moved })}</TableCell>
      <TableCell className={cn('text-right tabular-nums', summary.unmet_hours > 0 && 'text-destructive')}>
        {t('{hours} hrs', { hours: summary.unmet_hours })}
      </TableCell>
    </LinkRow>
  )
}
