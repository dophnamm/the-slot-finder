import { useParams } from 'react-router'
import { ProgramMembers } from '@/features/program-members'
import { ProgramOverview } from '@/features/program-overview'
import { ProgramProjects } from '@/features/program-projects'
import { useTranslation } from '@/shared/i18n'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'

export function ProgramDetailPage() {
  const { id = '' } = useParams()
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-4">
      <ProgramOverview programId={id} />
      <Tabs defaultValue="projects">
        <TabsList>
          <TabsTrigger value="projects">{t('Projects')}</TabsTrigger>
          <TabsTrigger value="members">{t('Members')}</TabsTrigger>
        </TabsList>
        <TabsContent value="projects">
          <ProgramProjects />
        </TabsContent>
        <TabsContent value="members">
          <ProgramMembers />
        </TabsContent>
      </Tabs>
    </div>
  )
}
