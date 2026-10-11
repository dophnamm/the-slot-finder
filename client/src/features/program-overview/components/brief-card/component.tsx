import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'
import type { BriefCardProps } from './types'

export function BriefCard({ program }: BriefCardProps) {
  const { t } = useTranslation()

  return (
    <Card>
      <CardHeader>
        <CardTitle>{t('Program brief')}</CardTitle>
        <CardDescription>{program.goal ? t('Goal: {goal}', { goal: program.goal }) : t('No goal yet.')}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col gap-3">
        {program.objectives.length > 0 && (
          <ol className="list-decimal space-y-1 pl-5">
            {program.objectives.map((objective) => (
              <li key={objective}>{objective}</li>
            ))}
          </ol>
        )}
        {program.charter_url && (
          <a className="text-sm text-primary hover:underline" href={program.charter_url} target="_blank" rel="noreferrer">
            {t('Open the charter')}
          </a>
        )}
      </CardContent>
    </Card>
  )
}
