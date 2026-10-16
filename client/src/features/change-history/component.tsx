import { History } from 'lucide-react'
import { useState } from 'react'
import { LoadError } from '@/shared/components/load-error'
import { Pager } from '@/shared/components/pager'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { useEvents } from '@/shared/events'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { ChangeRow } from './components/change-row'
import { HISTORY_PAGE_SIZE, HISTORY_TABLE_COLUMNS } from './constants'

/** Every change that may break the plan, with what the re-plan made of it. */
export function ChangeHistory() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const { data, error, isPending, isPlaceholderData, refetch } = useEvents(undefined, page, HISTORY_PAGE_SIZE)
  const events = data?.data ?? []

  const handleRetry = () => void refetch()

  return (
    <Section
      refreshing={isPlaceholderData}
      className="min-h-0 flex-1"
      title={t('Change history')}
      description={t(
        'Leave, holidays and slipped phases or deadlines. Each change is checked against the baseline; a conflict produces a staffing draft.',
      )}
    >
      <Table
        scroll="both"
        footer={<Pager page={page} total={data?.meta.total} pageSize={HISTORY_PAGE_SIZE} onPageChange={setPage} />}
      >
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
          {isPending && <TableLoading columns={HISTORY_TABLE_COLUMNS} />}
          {error && !data && (
            <LoadError columns={HISTORY_TABLE_COLUMNS} message={error.message} onRetry={handleRetry} />
          )}
          {!isPending && !error && events.length === 0 && (
            <TableEmpty
              columns={HISTORY_TABLE_COLUMNS}
              icon={History}
              title={t('No changes yet.')}
              description={t('Leave, holidays and moved phases or deadlines appear here.')}
            />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
