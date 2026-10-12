import { Users } from 'lucide-react'
import { useState } from 'react'
import { Pager } from '@/shared/components/pager'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { PersonFormDialog } from './components/person-form-dialog'
import { PersonRow } from './components/person-row'
import { PEOPLE_PAGE_SIZE, PEOPLE_TABLE_COLUMNS } from './constants'
import type { PeopleTableProps, Person } from './types'
import { usePeople } from './use-people'

/** Everyone who can be staffed, with their skills; planners add and edit people here. */
export function PeopleTable({ onShowAbsences }: PeopleTableProps) {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Person | null>(null)
  const { data, isPending } = usePeople(page)
  const people = data?.data ?? []

  const handleAdd = () => {
    setEditing(null)
    setFormOpen(true)
  }
  const handleEdit = (person: Person) => {
    setEditing(person)
    setFormOpen(true)
  }

  return (
    <Section
      title={t('People')}
      description={t('Expertise, levels and weekly hours of everyone who can be staffed.')}
      action={<Button onClick={handleAdd}>{t('Add person')}</Button>}
    >
      <Table footer={<Pager page={page} total={data?.meta.total} pageSize={PEOPLE_PAGE_SIZE} onPageChange={setPage} />}>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Person')}</TableHead>
            <TableHead>{t('Role')}</TableHead>
            <TableHead>{t('Hours')}</TableHead>
            <TableHead>{t('Expertise')}</TableHead>
            <TableHead>{t('Status')}</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {people.map((person) => (
            <PersonRow key={person.id} person={person} onEdit={handleEdit} onShowAbsences={onShowAbsences} />
          ))}
          {isPending && <TableLoading columns={PEOPLE_TABLE_COLUMNS} />}
          {!isPending && people.length === 0 && (
            <TableEmpty
              columns={PEOPLE_TABLE_COLUMNS}
              icon={Users}
              title={t('No people yet.')}
              description={t('Add the people who can be staffed, with their expertise and levels.')}
            />
          )}
        </TableBody>
      </Table>
      <PersonFormDialog key={editing?.id ?? 'new'} open={formOpen} person={editing} onOpenChange={setFormOpen} />
    </Section>
  )
}
