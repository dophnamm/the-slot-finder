import { useNavigate } from 'react-router'
import { toast } from 'sonner'
import { ROUTES } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'

/** Shows a success toast that also says a re-plan will run, with a link to the drafts inbox. */
export function useReplanToast() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  return (message: string) =>
    toast.success(message, {
      description: t('A re-plan will run shortly.'),
      action: { label: t('Open staffing drafts'), onClick: () => navigate(ROUTES.drafts) },
    })
}
