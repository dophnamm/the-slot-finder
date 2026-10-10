import { AbsenceList } from '@/features/absences'
import { useCurrentUser } from '@/shared/auth'
import { PagePlaceholder } from '@/shared/components/page-placeholder'
import { useTranslation } from '@/shared/i18n'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/shared/ui/card'

export function MySchedulePage() {
  const { t } = useTranslation()
  const { data: user, isPending } = useCurrentUser()

  if (isPending) return null

  return (
    <div className="flex flex-col gap-6">
      <PagePlaceholder
        title={t('My schedule')}
        description={t('Your assignments by week appear here once a staffing draft that includes you is accepted.')}
      />
      {user?.person_id ? (
        <Card>
          <CardHeader>
            <CardTitle>{t('My absences')}</CardTitle>
            <CardDescription>{t('Report leave or training. A planner approves it.')}</CardDescription>
          </CardHeader>
          <CardContent>
            <AbsenceList personId={user.person_id} canApprove={false} />
          </CardContent>
        </Card>
      ) : (
        <PagePlaceholder
          title={t('No person record')}
          description={t('Your account is not linked to a person yet. Ask a planner to add you with the same email.')}
        />
      )}
    </div>
  )
}
