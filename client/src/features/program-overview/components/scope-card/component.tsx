import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardHeader, CardTitle } from '@/shared/ui/card'
import { Stat } from '../stat'

export function ScopeCard() {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Scope (points)')}</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-3 gap-4">
        <Stat label={t('Total')} value="120" />
        <Stat label={t('Done')} value="54" />
        <Stat label={t('Done %')} value="45%" />
      </CardContent>
    </Card>
  )
}
