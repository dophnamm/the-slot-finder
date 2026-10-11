import { useParams } from 'react-router'
import { ProgramOverview } from '@/features/program-overview'
import { ProgramProjects } from '@/features/program-projects'
import { ReplanBanner } from '@/features/replan-banner'
import { useTranslation } from '@/shared/i18n'
import { Card, CardContent } from '@/shared/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'

export function ProgramDetailPage() {
  const { id = '' } = useParams()
  const { t } = useTranslation()

  return (
    <div className="flex flex-col gap-4">
      <ReplanBanner programId={id} />
      <ProgramOverview programId={id} />
      <Tabs defaultValue="projects">
        <TabsList>
          <TabsTrigger value="projects">{t('Projects')}</TabsTrigger>
          <TabsTrigger value="members">{t('Members')}</TabsTrigger>
        </TabsList>
        <TabsContent value="projects">
          <ProgramProjects programId={id} />
        </TabsContent>
        <TabsContent value="members">
          <Card>
            <CardContent className="text-sm text-muted-foreground">
              {t('Members appear here once a staffing draft for this program is accepted.')}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
