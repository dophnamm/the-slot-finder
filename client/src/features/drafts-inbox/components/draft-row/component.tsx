import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { LinkRow } from '@/shared/components/link-row'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { TableCell } from '@/shared/ui/table'
import { cn, formatDateTime } from '@/utils/functions'
import type { DraftRowProps } from './types'
import { programNames } from './utils'

/** One open draft: when and why it was made, what it covers, its baseline and what it could not staff. */
export function DraftRow({ draft }: DraftRowProps) {
  const { t } = useTranslation()
  const reviewPath = generatePath(ROUTES.staffingDraft, { id: draft.id })

  return (
    <LinkRow to={reviewPath}>
      <TableCell>
        <TextLink to={reviewPath}>{formatDateTime(draft.created_at)}</TextLink>
      </TableCell>
      <TableCell>
        <Badge variant={draft.auto ? 'amber' : 'outline'}>{draft.auto ? t('Re-plan') : t('Manual')}</Badge>
      </TableCell>
      <TableCell className="whitespace-normal">
        <div className="flex flex-col">
          <span>{draft.projects.map((project) => project.name).join(', ')}</span>
          <span className="text-xs text-muted-foreground">{programNames(draft.projects)}</span>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm">{t('Version {version}', { version: draft.base_version })}</span>
          {draft.stale && <Badge variant="amber">{t('Stale')}</Badge>}
        </div>
      </TableCell>
      <TableCell className={cn('text-right tabular-nums', draft.unmet_count > 0 && 'text-destructive')}>
        {draft.unmet_count}
      </TableCell>
    </LinkRow>
  )
}
