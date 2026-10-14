import { Star, UserMinus } from 'lucide-react'
import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { TextLink } from '@/shared/components/text-link'
import { useExpertiseBadge } from '@/shared/expertises'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatIsoWeekSpan } from '@/utils/functions'
import type { TeamMemberRowProps } from './types'

/** One member: name, skills, hours and weeks in this program, its projects, and Remove for planners. */
export function TeamMemberRow({ member, programId, canManage, busy, onRemove }: TeamMemberRowProps) {
  const { t } = useTranslation()
  const expertiseBadge = useExpertiseBadge()
  const { person } = member

  const handleRemove = () => onRemove(member)

  return (
    <TableRow>
      <TableCell>
        <div className="flex flex-col">
          <TextLink to={generatePath(ROUTES.personDetail, { id: person.id })}>{person.name}</TextLink>
          <span className="text-xs text-muted-foreground">{person.email}</span>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {person.skills.map((skill) => (
            <Badge
              key={skill.expertise_id}
              variant={expertiseBadge(skill.expertise_id)}
              title={skill.primary ? t('Primary expertise') : undefined}
            >
              {skill.primary && <Star data-icon="inline-start" aria-hidden />}
              {t('{expertise} · level {level}', { expertise: skill.expertise, level: skill.level })}
            </Badge>
          ))}
        </div>
      </TableCell>
      <TableCell className="text-right tabular-nums">{t('{hours} h', { hours: member.total_hours })}</TableCell>
      <TableCell className="text-sm">
        {member.from_week && member.to_week ? (
          formatIsoWeekSpan(member.from_week, member.to_week)
        ) : (
          <span className="text-muted-foreground">{t('Not staffed yet')}</span>
        )}
      </TableCell>
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
      {canManage && (
        <TableCell>
          <div className="flex justify-end">
            <Button
              variant="ghost"
              size="icon-sm"
              disabled={busy}
              aria-label={t('Remove {name} from the team', { name: person.name })}
              title={t('Remove {name} from the team', { name: person.name })}
              onClick={handleRemove}
            >
              <UserMinus />
            </Button>
          </div>
        </TableCell>
      )}
    </TableRow>
  )
}
