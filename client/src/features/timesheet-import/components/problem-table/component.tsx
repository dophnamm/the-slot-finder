import { useTranslation } from '@/shared/i18n'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'
import { problemLabel } from '../../utils'
import type { ProblemTableProps } from './types'

/** The problems of a failed import: line, column and what is wrong. */
export function ProblemTable({ problems }: ProblemTableProps) {
  const { t } = useTranslation()

  return (
    <Table scroll="both">
      <TableHeader>
        <TableRow>
          <TableHead>{t('Line')}</TableHead>
          <TableHead>{t('Column')}</TableHead>
          <TableHead>{t('Problem')}</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {problems.map((problem) => {
          const label = problemLabel(problem.code)
          return (
            <TableRow key={`${problem.line}-${problem.column}-${problem.code}`}>
              <TableCell className="tabular-nums">{problem.line}</TableCell>
              <TableCell>{problem.column || t('—')}</TableCell>
              <TableCell>{label ? t(label) : problem.code}</TableCell>
            </TableRow>
          )
        })}
      </TableBody>
    </Table>
  )
}
