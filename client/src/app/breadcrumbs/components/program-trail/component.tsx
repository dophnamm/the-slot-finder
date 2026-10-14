import { ROUTES } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'
import { useProgram } from '@/shared/programs'
import { Trail } from '../trail'
import type { ProgramTrailProps } from './types'

/** Programs › the program's name. */
export function ProgramTrail({ programId }: ProgramTrailProps) {
  const { t } = useTranslation()
  const { data: program } = useProgram(programId)

  return (
    <Trail
      crumbs={[
        { label: t('Programs'), to: ROUTES.programs },
        program ? { label: program.name } : { label: t('Loading…'), pending: true },
      ]}
    />
  )
}
