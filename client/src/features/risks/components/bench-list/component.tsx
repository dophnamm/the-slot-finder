import { generatePath } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { LoadError } from '@/shared/components/load-error'
import { TextLink } from '@/shared/components/text-link'
import { useTranslation } from '@/shared/i18n'
import { Skeleton } from '@/shared/ui/skeleton'
import { LOADING_REVEAL_CLASS } from '@/utils/constants'
import { cn, range } from '@/utils/functions'
import { RISK_LOADING_ITEMS } from '../../constants'
import { useBench } from '../../use-bench'

/** People with spare hours in the coming weeks, each linking to their page. */
export function BenchList() {
  const { t } = useTranslation()
  const { data, error, isPending, refetch } = useBench()
  const people = data?.people ?? []

  const handleRetry = () => void refetch()

  if (error && !data) return <LoadError framed error={error} onRetry={handleRetry} />

  return (
    <section className="flex flex-col gap-2">
      <h3 className="text-sm font-medium">{isPending ? t('Bench') : t('Bench ({count})', { count: people.length })}</h3>
      {people.length === 0 && !isPending && (
        <p className="text-sm text-muted-foreground">{t('Nobody is on the bench')}</p>
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
          <li key={person.person_id} className="flex flex-col gap-1 py-3 first:pt-0 last:pb-0">
            <TextLink to={generatePath(ROUTES.personDetail, { id: person.person_id })} className="font-medium">
              {person.name}
            </TextLink>
            <div className="flex flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground">
              <span>{person.primary_expertise}</span>
              <span className="tabular-nums">
                {t('{hours} hrs free in {weeks} weeks', {
                  hours: Math.round(person.free_hours),
                  weeks: data?.weeks ?? 0,
                })}
              </span>
            </div>
          </li>
        ))}
      </ul>
    </section>
  )
}
