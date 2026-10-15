import { ROUTES } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'
import { useProgram } from '@/shared/programs'
import { Trail } from '../trail'
import type { ProgramTrailProps } from './types'

/** Programs › the program's name, or Program when it cannot be loaded. */
export function ProgramTrail({ programId }: ProgramTrailProps) {
  const { t } = useTranslation()
  const { data: program, error } = useProgram(programId)
  const current = program
    ? { label: program.name }
    : error
      ? { label: t('Program') }
      : { label: t('Loading…'), pending: true }

  return <Trail crumbs={[{ label: t('Programs'), to: ROUTES.programs }, current]} />
}
