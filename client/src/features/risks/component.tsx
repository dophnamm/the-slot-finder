import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'

import { LoadError } from '@/shared/components/load-error'
import { Section } from '@/shared/components/section'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { Skeleton } from '@/shared/ui/skeleton'
import { LOADING_REVEAL_CLASS } from '@/utils/constants'
import { cn, range } from '@/utils/functions'
import { OverloadedPerson } from './components/overloaded-person'
import { RiskProject } from './components/risk-project'
import { RISK_LOADING_ITEMS } from './constants'
import { useRisks } from './use-risks'

/** One list of what needs attention: projects that are understaffed, at risk or behind, and overloaded people. */
export function Risks() {
  const { t } = useTranslation()
  const { data, error, isPending, refetch } = useRisks()
  const projects = data?.projects ?? []
  const people = data?.people ?? []
  const sustained = data?.sustained_overload ?? []

  const failed = Boolean(error) && !data

  const handleRetry = () => void refetch()

  return (
    <Section
      card
      title={t('Risks')}
      description={t(
        'Projects that are understaffed, at risk or behind on delivery, and people over 100 % in the coming weeks.',
      )}
      className="h-full min-h-0"
      contentClassName="gap-5 overflow-y-auto"
    >
      {failed && error && <LoadError framed error={error} onRetry={handleRetry} />}
      {!failed && (
        <>
          {isPending && (
            <span className="sr-only" role="status">
              {t('Loading…')}
            </span>
          )}
          <section className="flex flex-col gap-2">
            <h3 className="text-sm font-medium">
              {isPending ? t('Projects') : t('Projects ({count})', { count: projects.length })}
            </h3>
            {projects.length === 0 && !isPending && (
              <p className="text-sm text-muted-foreground">{t('No project needs attention.')}</p>
            )}
            {isPending && (
              <ul aria-busy className={cn('flex flex-col divide-y', LOADING_REVEAL_CLASS)}>
                {range(RISK_LOADING_ITEMS).map((item) => (
                  <li key={item} className="flex flex-col gap-2 py-3 first:pt-0 last:pb-0">
                    <Skeleton className="h-3.5 w-1/2" />
                    <Skeleton className="h-3 w-1/3" />
                  </li>
                ))}
              </ul>
            )}
            <ul className="flex flex-col divide-y">
              {projects.map((project) => (
                <RiskProject key={project.id} project={project} />
              ))}
            </ul>
          </section>
          {sustained.length > 0 && (
            <section className="flex flex-col gap-2">
              <h3 className="text-sm font-medium">{t('Sustained overload ({count})', { count: sustained.length })}</h3>
              <ul className="flex flex-col divide-y">
                {sustained.map((item) => (
                  <li key={item.person.id} className="py-3 text-sm text-destructive first:pt-0 last:pb-0">
                    <TextLink
                      to={generatePath(ROUTES.personDetail, { id: item.person.id })}
                      className="text-destructive"
                    >
                      {t('{name}: over 100 % for {weeks} weeks from {week}', {
                        name: item.person.name,
                        weeks: item.weeks,
                        week: item.from_week,
                      })}
                    </TextLink>
                  </li>
                ))}
              </ul>
            </section>
          )}
          <section className="flex flex-col gap-2">
            <h3 className="text-sm font-medium">
              {isPending ? t('Overloaded people') : t('Overloaded people ({count})', { count: people.length })}
            </h3>
            {people.length === 0 && !isPending && (
              <p className="text-sm text-muted-foreground">{t('Nobody is over 100 % in the coming weeks.')}</p>
            )}
            {isPending && (
              <ul aria-busy className={cn('flex flex-col divide-y', LOADING_REVEAL_CLASS)}>
                {range(RISK_LOADING_ITEMS).map((item) => (
                  <li key={item} className="flex flex-col gap-2 py-3 first:pt-0 last:pb-0">
                    <Skeleton className="h-3.5 w-1/2" />
                    <Skeleton className="h-3 w-1/3" />
                  </li>
                ))}
              </ul>
            )}
            <ul className="flex flex-col divide-y">
              {people.map((person) => (
                <OverloadedPerson key={person.person.id} person={person} />
              ))}
            </ul>
          </section>
        </>
      )}
    </Section>
  )
}
