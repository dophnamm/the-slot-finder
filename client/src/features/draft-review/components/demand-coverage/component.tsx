import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { CoverageRow } from './components/coverage-row'
import type { DemandCoverageProps } from './types'

/** Each covered demand: the hours it needs and what the draft gives it. */
export function DemandCoverage({ demands }: DemandCoverageProps) {
  const { t } = useTranslation()

  return (
    <Section
      title={t('Coverage')}
      description={t('Budget hours each demand needs and the hours this draft assigns.')}
    >
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Project')}</TableHead>
            <TableHead>{t('Expertise')}</TableHead>
            <TableHead>{t('Phase')}</TableHead>
            <TableHead>{t('Weeks')}</TableHead>
            <TableHead className="text-right">{t('Needed')}</TableHead>
            <TableHead className="text-right">{t('Assigned')}</TableHead>
            <TableHead>{t('Status')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {demands.map((demand) => (
            <CoverageRow key={demand.id} demand={demand} />
          ))}
        </TableBody>
      </Table>
    </Section>
  )
}
