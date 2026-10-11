import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'
import { OverloadedPerson } from './components/overloaded-person'
import { RiskProject } from './components/risk-project'
import { useRisks } from './use-risks'

/** One list of what needs attention: projects that are understaffed, at risk or behind, and overloaded people. */
export function Risks() {
  const { t } = useTranslation()
  const { data, isPending } = useRisks()
  const projects = data?.projects ?? []
  const people = data?.people ?? []

  return (
    <Section
      card
      title={t('Risks')}
      description={t('Projects that are understaffed, at risk or behind on delivery, and people over 100 % in the coming weeks.')}
      className="h-full min-h-0"
      contentClassName="gap-5 overflow-y-auto"
    >
      {isPending && <p className="text-sm text-muted-foreground">{t('Loading…')}</p>}
      <section className="flex flex-col gap-2">
        <h3 className="text-sm font-medium">{t('Projects ({count})', { count: projects.length })}</h3>
        {projects.length === 0 && !isPending && <p className="text-sm text-muted-foreground">{t('No project needs attention.')}</p>}
        <ul className="flex flex-col gap-2">
          {projects.map((project) => (
            <RiskProject key={project.id} project={project} />
          ))}
        </ul>
      </section>
      <section className="flex flex-col gap-2">
        <h3 className="text-sm font-medium">{t('Overloaded people ({count})', { count: people.length })}</h3>
        {people.length === 0 && !isPending && <p className="text-sm text-muted-foreground">{t('Nobody is over 100 % in the coming weeks.')}</p>}
        <ul className="flex flex-col gap-2">
          {people.map((person) => (
            <OverloadedPerson key={person.person.id} person={person} />
          ))}
        </ul>
      </section>
    </Section>
  )
}
