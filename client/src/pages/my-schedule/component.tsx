import { AbsenceList } from '@/features/absences'
import { useCurrentUser } from '@/shared/auth'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { Section } from '@/shared/components/section'
import { useTranslation } from '@/shared/i18n'

export function MySchedulePage() {
  const { t } = useTranslation()
  const { data: user, isPending } = useCurrentUser()

  if (isPending) return null

  return (
    <div className="flex flex-col gap-8">
      <PagePlaceholder
        title={t('My schedule')}
        description={t('Your assignments by week appear here once a staffing draft that includes you is accepted.')}
      />
      {user?.person_id ? (
        <Section title={t('My absences')} description={t('Report leave or training. A planner approves it.')}>
          <AbsenceList personId={user.person_id} canApprove={false} />
        </Section>
      ) : (
        <PagePlaceholder
          title={t('No person record')}
          description={t('Your account is not linked to a person yet. Ask a planner to add you with the same email.')}
        />
      )}
    </div>
  )
}
