import { generatePath, useNavigate } from 'react-router'
import { ProgramsTable } from '@/features/programs'
import { ROUTES } from '@/resources/routes'
import type { Program } from '@/shared/programs'

export function ProgramsPage() {
  const navigate = useNavigate()

  const handleCreated = (program: Program) => navigate(generatePath(ROUTES.programDetail, { id: program.id }))

  return <ProgramsTable onCreated={handleCreated} />
}
