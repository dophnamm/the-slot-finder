import { Users } from 'lucide-react'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { HIDDEN_PERSON_COLUMNS, TEAM_TABLE_COLUMNS } from './constants'
import { TeamMemberRow } from './components/team-member-row'
import type { TeamTableProps } from './types'

/** A program's team: each member's skills, and their hours, weeks and projects in the program's baseline. */
export function TeamTable({
  members,
  isPending,
  programId,
  canManage,
  canViewPeople,
  removingId,
  onRemove,
}: TeamTableProps) {
  const { t } = useTranslation()
  const columns = TEAM_TABLE_COLUMNS + Number(canManage) - (canViewPeople ? 0 : HIDDEN_PERSON_COLUMNS)

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t('Person')}</TableHead>
          {canViewPeople && (
            <>
              <TableHead>{t('Expertise')}</TableHead>
              <TableHead className="text-right">{t('Hours in this program')}</TableHead>
              <TableHead>{t('Weeks')}</TableHead>
              <TableHead>{t('Projects')}</TableHead>
            </>
          )}
          {canManage && (
            <TableHead>
              <span className="sr-only">{t('Actions')}</span>
            </TableHead>
          )}
        </TableRow>
      </TableHeader>
      <TableBody>
        {members.map((member) => (
          <TeamMemberRow
            key={member.person.id}
            member={member}
            programId={programId}
            canManage={canManage}
            canViewPeople={canViewPeople}
            removing={removingId === member.person.id}
            onRemove={onRemove}
          />
        ))}
        {isPending && <TableLoading columns={columns} />}
        {!isPending && members.length === 0 && (
          <TableEmpty
            columns={columns}
            icon={Users}
            title={t('Nobody is in this team yet.')}
            description={
              canManage
                ? t('Add the people this program’s projects are staffed from.')
                : canViewPeople
                  ? t('A planner adds the people this program’s projects are staffed from.')
                  : t('The people this program’s projects are staffed from show here.')
            }
          />
        )}
      </TableBody>
    </Table>
  )
}
