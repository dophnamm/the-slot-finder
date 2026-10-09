import { useTranslation } from '@/shared/i18n'
import { Card, CardContent } from '@/shared/ui/card'

/** Projects tab of the program page. */
export function ProgramProjects() {
  const { t } = useTranslation()

  return (
    <Card>
      <CardContent className="text-muted-foreground">
        {t(
          'Per project: squad, status, start–end and next milestone, Time / Scope / Cost bars, staffing % and “Propose squad”.',
        )}
      </CardContent>
    </Card>
  )
}
