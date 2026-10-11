import { Section } from '@/shared/components/section'
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
        <a className="text-sm text-primary hover:underline" href={program.charter_url} target="_blank" rel="noreferrer">
          {t('Open the charter')}
        </a>
      )}
    </Section>
  )
}
