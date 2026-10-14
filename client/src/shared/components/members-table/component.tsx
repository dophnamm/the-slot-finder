import { Users } from 'lucide-react'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { MEMBER_TABLE_COLUMNS } from './constants'
import { MemberRow } from './components/member-row'
import type { MembersTableProps } from './types'

/** A program's members: expertise, hours, weeks and projects. Used by the Programs list preview. */
export function MembersTable({ members, isPending, programId }: MembersTableProps) {
  const { t } = useTranslation()

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>{t('Person')}</TableHead>
          <TableHead>{t('Expertise')}</TableHead>
          <TableHead className="text-right">{t('Hours')}</TableHead>
          <TableHead>{t('Weeks')}</TableHead>
          <TableHead>{t('Projects')}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {members.map((member) => (
          <MemberRow key={member.person.id} member={member} programId={programId} />
        ))}
        {isPending && <TableLoading columns={MEMBER_TABLE_COLUMNS} />}
        {!isPending && members.length === 0 && (
          <TableEmpty
            columns={MEMBER_TABLE_COLUMNS}
            icon={Users}
            title={t('Nobody has hours on this program yet.')}
            description={t('People appear here once a staffing draft for this program is accepted.')}
          />
        )}
      </TableBody>
    </Table>
  )
}
