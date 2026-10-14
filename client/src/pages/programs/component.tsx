import { generatePath, useNavigate } from 'react-router'
import { ProgramsTable } from '@/features/programs'
import { ROUTES } from '@/resources/routes'
import type { Program } from '@/shared/programs'
import { PageFill } from '@/shared/components/page-fill'

export function ProgramsPage() {
  const navigate = useNavigate()

  const handleCreated = (program: Program) => navigate(generatePath(ROUTES.programDetail, { id: program.id }))

  return (
    <PageFill>
      <ProgramsTable onCreated={handleCreated} />
    </PageFill>
  )
}
