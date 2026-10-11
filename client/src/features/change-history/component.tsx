import { useState } from 'react'
import { Pager } from '@/shared/components/pager'
import { Section } from '@/shared/components/section'
import { useEvents } from '@/shared/events'
import { useTranslation } from '@/shared/i18n'
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
    <Section
      title={t('Change history')}
      description={t('Leave, holidays and slipped phases or deadlines. Each change is checked against the baseline; a conflict produces a staffing draft.')}
      footer={<Pager page={page} total={data?.meta.total} pageSize={HISTORY_PAGE_SIZE} onPageChange={setPage} />}
    >
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
    </Section>
  )
}
