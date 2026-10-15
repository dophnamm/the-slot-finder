import { Fragment } from 'react'
import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { TextLink } from '@/shared/components/text-link'
import { useExpertiseBadge } from '@/shared/expertises'
import { useTranslation } from '@/shared/i18n'
import { ASSIGNMENT_KIND_BADGE, ASSIGNMENT_KIND_LABELS, PRIMARY_KIND } from '@/shared/staffing'
import { Badge } from '@/shared/ui/badge'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatIsoWeekSpan } from '@/utils/functions'
import type { MemberRowProps } from './types'

/** One member: hours per expertise (cross-expertise and stretch marked), total, weeks and projects. */
export function MemberRow({ member, programId }: MemberRowProps) {
  const { t } = useTranslation()
  const expertiseBadge = useExpertiseBadge()

  return (
    <TableRow>
      <TableCell className="font-medium">{member.person.name}</TableCell>
      <TableCell>
        <div className="flex flex-wrap items-center gap-1">
          {member.expertises.map((row) => (
            <Fragment key={`${row.expertise_id}:${row.kind}`}>
              <Badge variant={expertiseBadge(row.expertise_id)}>
                {t('{expertise} · {hours} hrs', { expertise: row.expertise, hours: row.hours })}
              </Badge>
              {row.kind !== PRIMARY_KIND && (
                <Badge variant={ASSIGNMENT_KIND_BADGE[row.kind]}>{t(ASSIGNMENT_KIND_LABELS[row.kind])}</Badge>
              )}
            </Fragment>
          ))}
        </div>
      </TableCell>
      <TableCell className="text-right tabular-nums">{t('{hours} hrs', { hours: member.total_hours })}</TableCell>
      <TableCell className="text-sm">{formatIsoWeekSpan(member.from_week, member.to_week)}</TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-x-3 gap-y-1">
          {member.projects.map((project) => (
            <TextLink
              key={project.id}
              className="text-sm"
              to={generatePath(ROUTES.projectDetail, { programId, id: project.id })}
            >
              {project.name}
            </TextLink>
          ))}
        </div>
      </TableCell>
    </TableRow>
  )
}
