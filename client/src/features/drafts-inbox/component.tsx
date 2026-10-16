import { FileStack } from 'lucide-react'
import { LoadError } from '@/shared/components/load-error'
import { Section } from '@/shared/components/section'
import { TableEmpty } from '@/shared/components/table-empty'
import { TableLoading } from '@/shared/components/table-loading'
import { Term } from '@/shared/components/term'
import { useTranslation } from '@/shared/i18n'
import { useOpenDrafts } from '@/shared/staffing'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { DraftRow } from './components/draft-row'
import { DRAFT_TABLE_COLUMNS } from './constants'

/** Every open staffing draft, manual and re-plan, newest first; each row opens the draft review. */
export function DraftsInbox() {
  const { t } = useTranslation()
  const { data, error, isPending, refetch } = useOpenDrafts()
  const drafts = data ?? []

  const handleRetry = () => void refetch()

  return (
    <Section
      className="min-h-0 flex-1"
      title={t('Open staffing drafts')}
      description={t(
        'Drafts waiting for review, newest first. Re-plan drafts come from changes such as leave or a slipped phase.',
      )}
    >
      <Table scroll="both">
        <TableHeader>
          <TableRow>
            <TableHead>{t('Created')}</TableHead>
            <TableHead>{t('Origin')}</TableHead>
            <TableHead>{t('Projects')}</TableHead>
            <TableHead>
              <Term term="baseline">{t('Baseline')}</Term>
            </TableHead>
            <TableHead className="text-right">{t('Unmet demands')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {drafts.map((draft) => (
            <DraftRow key={draft.id} draft={draft} />
          ))}
          {isPending && <TableLoading columns={DRAFT_TABLE_COLUMNS} />}
          {error && !data && <LoadError columns={DRAFT_TABLE_COLUMNS} error={error} onRetry={handleRetry} />}
          {!isPending && !error && drafts.length === 0 && (
            <TableEmpty columns={DRAFT_TABLE_COLUMNS} icon={FileStack} title={t('No open staffing drafts.')} />
          )}
        </TableBody>
      </Table>
    </Section>
  )
}
