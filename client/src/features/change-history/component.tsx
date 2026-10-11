import { useState } from 'react'
import { Pager } from '@/shared/components/pager'
import { useEvents } from '@/shared/events'
import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { ChangeRow } from './components/change-row'
import { HISTORY_PAGE_SIZE } from './constants'

/** Every change that may break the plan, with what the re-plan made of it. */
export function ChangeHistory() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const { data, isPending } = useEvents(undefined, page, HISTORY_PAGE_SIZE)
  const events = data?.data ?? []

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Change history')}</CardTitle>
        <CardDescription>
          {t('Leave, holidays and slipped phases or deadlines. Each change is checked against the baseline; a conflict produces a staffing draft.')}
        </CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('When')}</TableHead>
              <TableHead>{t('Change')}</TableHead>
              <TableHead>{t('Status')}</TableHead>
              <TableHead>{t('Draft')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {events.map((event) => (
              <ChangeRow key={event.id} event={event} />
            ))}
          </TableBody>
        </Table>
        {isPending && <p className="py-6 text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
        {!isPending && events.length === 0 && <p className="py-6 text-center text-sm text-muted-foreground">{t('No changes yet.')}</p>}
      </CardContent>
      <CardFooter className="justify-end">
        <Pager page={page} total={data?.meta.total} pageSize={HISTORY_PAGE_SIZE} onPageChange={setPage} />
      </CardFooter>
    </Card>
  )
}
