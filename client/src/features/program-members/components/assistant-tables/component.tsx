import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/shared/ui/table'
import { EXTERNAL_ASSISTANTS, INTERNAL_ASSISTANTS } from '../../mock-data'

export function AssistantTables() {
  const { t } = useTranslation()

  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>{t('Internal')}</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Person')}</TableHead>
                <TableHead>{t('Email')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {INTERNAL_ASSISTANTS.map((row) => (
                <TableRow key={row.email}>
                  <TableCell>{row.person}</TableCell>
                  <TableCell>{row.email}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>{t('External')}</CardTitle>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('Name')}</TableHead>
                <TableHead>{t('Email')}</TableHead>
                <TableHead>{t('Description')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {EXTERNAL_ASSISTANTS.map((row) => (
                <TableRow key={row.email}>
                  <TableCell>{row.name}</TableCell>
                  <TableCell>{row.email}</TableCell>
                  <TableCell>{row.description}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
