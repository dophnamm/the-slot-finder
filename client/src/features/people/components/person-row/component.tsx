import { Star } from 'lucide-react'
import { ROLE_LABELS } from '@/resources/rbac'
import { useExpertiseBadge } from '@/shared/expertises'
import { useTranslation } from '@/shared/i18n'
import { Badge } from '@/shared/ui/badge'
import { Button } from '@/shared/ui/button'
import { TableCell, TableRow } from '@/shared/ui/table'
import { formatDate } from '@/utils/functions'
import type { PersonRowProps } from './types'

export function PersonRow({ person, onEdit, onShowAbsences }: PersonRowProps) {
  const { t } = useTranslation()
  const expertiseBadge = useExpertiseBadge()

  const handleEdit = () => onEdit(person)
  const handleShowAbsences = () => onShowAbsences?.(person)

  return (
    <TableRow>
      <TableCell>
        <div className="flex flex-col">
          <span className="font-medium">{person.name}</span>
          <span className="text-xs text-muted-foreground">{person.email}</span>
        </div>
      </TableCell>
      <TableCell>{t(ROLE_LABELS[person.role])}</TableCell>
      <TableCell>
        <div className="flex flex-col">
          <span>{t('{hours} h per week', { hours: person.weekly_hours })}</span>
          <span className="text-xs text-muted-foreground">
            {t('{hours} h non-project', { hours: person.non_project_hours })}
          </span>
        </div>
      </TableCell>
      <TableCell>
        <div className="flex flex-wrap gap-1">
          {person.skills.map((skill) => (
            <Badge key={skill.expertise_id} variant={expertiseBadge(skill.expertise_id)} title={skill.primary ? t('Primary expertise') : undefined}>
              {skill.primary && <Star data-icon="inline-start" aria-hidden />}
              {t('{expertise} · level {level}', { expertise: skill.expertise, level: skill.level })}
            </Badge>
          ))}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex flex-col items-start gap-1">
          <Badge variant={person.active ? 'green' : 'gray'}>{person.active ? t('Active') : t('Inactive')}</Badge>
          {person.start_date && (
            <span className="text-xs text-muted-foreground">{t('From {date}', { date: formatDate(person.start_date) })}</span>
          )}
          {person.end_date && (
            <span className="text-xs text-muted-foreground">{t('Until {date}', { date: formatDate(person.end_date) })}</span>
          )}
        </div>
      </TableCell>
      <TableCell>
        <div className="flex justify-end gap-2">
          {onShowAbsences && (
            <Button variant="outline" size="sm" onClick={handleShowAbsences}>
              {t('Absences')}
            </Button>
          )}
          <Button variant="outline" size="sm" onClick={handleEdit}>
            {t('Edit')}
          </Button>
        </div>
      </TableCell>
    </TableRow>
  )
}
