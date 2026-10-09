import { useTranslation } from '@/shared/i18n'
import { Card, CardContent } from '@/shared/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/shared/ui/tabs'
import { AssistantTables } from './components/assistant-tables'

/** Members tab of the program page: members and assistants. */
export function ProgramMembers() {
  const { t } = useTranslation()

  return (
    <Tabs defaultValue="member">
      <TabsList variant="line">
        <TabsTrigger value="member">{t('Member')}</TabsTrigger>
        <TabsTrigger value="assistant">{t('Assistant')}</TabsTrigger>
      </TabsList>
      <TabsContent value="member">
        <Card>
          <CardContent className="text-muted-foreground">
            {t(
              'Per person: squad role, expertise, hours here and overall utilization over the next 4 weeks.',
            )}
          </CardContent>
        </Card>
      </TabsContent>
      <TabsContent value="assistant">
        <AssistantTables />
      </TabsContent>
    </Tabs>
  )
}
