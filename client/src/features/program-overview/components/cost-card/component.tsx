import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import { Stat } from '../stat'

export function CostCard() {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Cost (hours)')}</CardTitle>
        <CardDescription>{t('Actual hours arrive in phase 2.')}</CardDescription>
      </CardHeader>
      <CardContent className="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <Stat label={t('Budget')} value="2,760" />
        <Stat label={t('Planned')} value="2,430" />
        <Stat label={t('Planned to date')} value="1,120" />
        <Stat label={t('Gap')} value="330" />
      </CardContent>
    </Card>
  )
}
