import { useEffect } from 'react'
import { generatePath, useNavigate } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useCurrentUser } from '@/shared/auth'
import type { Program } from '@/shared/programs'
import { onlyProgramId } from './utils'

/** A manager whose programs list is a single program lands on that program; planners always see the list. */
export function useOnlyProgramRedirect(programs: Program[], total: number | undefined) {
  const navigate = useNavigate()
  const { data: user } = useCurrentUser()
  const programId = onlyProgramId(user?.role, programs, total)

  useEffect(() => {
    if (programId) navigate(generatePath(ROUTES.programDetail, { id: programId }), { replace: true })
  }, [navigate, programId])
}
