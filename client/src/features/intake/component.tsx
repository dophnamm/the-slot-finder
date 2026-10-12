import { useState } from 'react'
import { Pager } from '@/shared/components/pager'
import { Section } from '@/shared/components/section'
import { PROGRAMS_PAGE_SIZE, useProjects } from '@/shared/programs'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { IntakeRow } from './components/intake-row'
import { INTAKE_STATUS } from './constants'

/** Incoming projects across programs, highest priority and earliest deadline first. */
export function IntakeQueue() {
  const { t } = useTranslation()
  const [page, setPage] = useState(1)
  const { data, isPending } = useProjects(INTAKE_STATUS, page)
  const projects = data?.data ?? []

  return (
    <Section
      title={t('Intake queue')}
      description={t('Incoming projects across programs, highest priority and earliest deadline first.')}
    >
      <Table
        footer={<Pager page={page} total={data?.meta.total} pageSize={PROGRAMS_PAGE_SIZE} onPageChange={setPage} />}
      >
        <TableHeader>
          <TableRow>
            <TableHead>{t('Project')}</TableHead>
            <TableHead>{t('Program')}</TableHead>
            <TableHead>{t('Start – deadline')}</TableHead>
            <TableHead>{t('Current phase')}</TableHead>
            <TableHead>{t('Lead')}</TableHead>
            <TableHead />
          </TableRow>
        </TableHeader>
        <TableBody>
          {projects.map((project) => (
            <IntakeRow key={project.id} project={project} />
          ))}
        </TableBody>
      </Table>
      {isPending && <p className="py-6 text-center text-sm text-muted-foreground">{t('Loading…')}</p>}
      {!isPending && projects.length === 0 && (
        <p className="py-6 text-center text-sm text-muted-foreground">{t('No incoming projects.')}</p>
      )}
    </Section>
  )
}
