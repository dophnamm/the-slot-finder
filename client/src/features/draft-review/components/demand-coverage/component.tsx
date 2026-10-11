import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { CoverageRow } from './components/coverage-row'
import type { DemandCoverageProps } from './types'

/** Each covered demand: the hours it needs and what the draft gives it. */
export function DemandCoverage({ demands }: DemandCoverageProps) {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Coverage')}</CardTitle>
        <CardDescription>{t('Budget hours each demand needs and the hours this draft assigns.')}</CardDescription>
      </CardHeader>
      <CardContent>
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
      </CardContent>
    </Card>
  )
}
