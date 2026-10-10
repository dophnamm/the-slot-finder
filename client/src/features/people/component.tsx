import { useState } from 'react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Card, CardAction, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { PersonFormDialog } from './components/person-form-dialog'
import { PersonRow } from './components/person-row'
import { PEOPLE_PAGE_SIZE } from './constants'
import type { PeopleTableProps, Person } from './types'
import { usePeople } from './use-people'
import { pageCount } from './utils'

/** Everyone who can be staffed, with their skills; planners add and edit people here. */
export function PeopleTable({ onShowAbsences }: PeopleTableProps) {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Person | null>(null)
  const { data, isPending } = usePeople(page)
  const people = data?.data ?? []
  const pages = pageCount(data?.meta.total, PEOPLE_PAGE_SIZE)

  const handleAdd = () => {
    setEditing(null)
    setFormOpen(true)
  }
  const handleEdit = (person: Person) => {
    setEditing(person)
    setFormOpen(true)
  }
  const handlePrevious = () => setPage((current) => Math.max(1, current - 1))
  const handleNext = () => setPage((current) => Math.min(pages, current + 1))

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('People')}</CardTitle>
        <CardDescription>{t('Expertise, levels and weekly hours of everyone who can be staffed.')}</CardDescription>
        <CardAction>
          <Button onClick={handleAdd}>{t('Add person')}</Button>
        </CardAction>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Person')}</TableHead>
              <TableHead>{t('Role')}</TableHead>
              <TableHead>{t('Hours')}</TableHead>
              <TableHead>{t('Skills')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
              <TableHead />
            </TableRow>
          </TableHeader>
          <TableBody>
            {people.map((person) => (
              <PersonRow key={person.id} person={person} onEdit={handleEdit} onShowAbsences={onShowAbsences} />
            ))}
          </TableBody>
        </Table>
        {isPending && <p className="py-6 text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
        {!isPending && people.length === 0 && (
          <p className="py-6 text-center text-sm text-muted-foreground">{t('No people yet.')}</p>
        )}
      </CardContent>
      <CardFooter className="justify-end gap-2">
        <span className="text-sm text-muted-foreground">{t('Page {page} of {pages}', { page, pages })}</span>
        <Button variant="outline" size="sm" onClick={handlePrevious} disabled={page <= 1}>
          {t('Previous')}
        </Button>
        <Button variant="outline" size="sm" onClick={handleNext} disabled={page >= pages}>
          {t('Next')}
        </Button>
      </CardFooter>
      <PersonFormDialog key={editing?.id ?? 'new'} open={formOpen} person={editing} onOpenChange={setFormOpen} />
    </Card>
  )
}
