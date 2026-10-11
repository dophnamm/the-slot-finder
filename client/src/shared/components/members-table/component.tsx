import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { MemberRow } from './components/member-row'
import type { MembersTableProps } from './types'

/** A program's members: expertise, hours, weeks and projects. Used by the Members tab and the Programs list preview. */
export function MembersTable({ members, isPending, programId }: MembersTableProps) {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-4">
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
        </TableBody>
      </Table>
      {isPending && <p className="py-6 text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
      {!isPending && members.length === 0 && (
        <p className="py-6 text-center text-sm text-muted-foreground">{t('Nobody has hours on this program yet.')}</p>
      )}
    </div>
  )
}
