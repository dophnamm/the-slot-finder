import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/shared/ui/table'
import { STAFFING } from '../../mock-data'

export function StaffingByExpertise() {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Staffing by expertise')}</CardTitle>
        <CardDescription>{t('Budget, planned and forecast hours per expertise.')}</CardDescription>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('Expertise')}</TableHead>
              <TableHead className="text-right">{t('Budget')}</TableHead>
              <TableHead className="text-right">{t('Planned')}</TableHead>
              <TableHead className="text-right">{t('Forecast')}</TableHead>
              <TableHead className="text-right">{t('Gap')}</TableHead>
              <TableHead className="text-right">{t('Staffing %')}</TableHead>
              <TableHead className="text-right">{t('Contribution %')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody className="tabular-nums">
            {STAFFING.map((row) => (
              <TableRow key={row.expertise}>
                <TableCell className="font-medium">{row.expertise}</TableCell>
                <TableCell className="text-right">{row.budget}</TableCell>
                <TableCell className="text-right">{row.planned}</TableCell>
                <TableCell className="text-right">{row.forecast}</TableCell>
                <TableCell className="text-right">{row.gap}</TableCell>
                <TableCell className="text-right">{row.staffing}%</TableCell>
                <TableCell className="text-right">{row.contribution}%</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
