import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { GOAL, OBJECTIVES } from '../../mock-data'

export function BriefCard() {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Program brief')}</CardTitle>
        <CardDescription>{t('Goal: {goal}', { goal: GOAL })}</CardDescription>
      </CardHeader>
      <CardContent>
        <ol className="list-decimal space-y-1 pl-5">
          {OBJECTIVES.map((objective) => (
            <li key={objective}>{objective}</li>
          ))}
        </ol>
      </CardContent>
    </Card>
  )
}
