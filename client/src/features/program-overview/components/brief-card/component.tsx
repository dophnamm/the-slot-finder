import { Section } from '@/shared/components/section'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import type { BriefCardProps } from './types'

export function BriefCard({ program }: BriefCardProps) {
  const { t } = useTranslation()

  return (
    <Section
      card
      title={t('Program brief')}
      description={program.goal ? t('Goal: {goal}', { goal: program.goal }) : t('No goal yet.')}
      contentClassName="gap-3"
    >
      {program.objectives.length > 0 && (
        <ol className="list-decimal space-y-1 pl-5">
          {program.objectives.map((objective) => (
            <li key={objective}>{objective}</li>
          ))}
        </ol>
      )}
      {program.charter_url && (
        <TextLink className="text-sm" to={program.charter_url} external>
          {t('Open the charter')}
        </TextLink>
      )}
    </Section>
  )
}
