import { House } from 'lucide-react'
import { Fragment } from 'react'
import { Link } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { useTranslation } from '@/shared/i18n'
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbPage, BreadcrumbSeparator } from '@/shared/ui/breadcrumb'
import type { TrailProps } from './types'

/** Renders a trail after a home icon: links up to the current page, which is plain text. */
export function Trail({ crumbs }: TrailProps) {
  const { t } = useTranslation()

  return (
    <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink aria-label={t('Home')} render={<Link to={ROUTES.dashboard} />}>
            <House className="size-4" />
          </BreadcrumbLink>
        </BreadcrumbItem>
        {crumbs.map((crumb, index) => (
          <Fragment key={`${index}:${crumb.label}`}>
            <BreadcrumbSeparator />
            <BreadcrumbItem>
              {crumb.to && index < crumbs.length - 1 ? (
                <BreadcrumbLink render={<Link to={crumb.to} />}>{crumb.label}</BreadcrumbLink>
              ) : (
                <BreadcrumbPage>{crumb.label}</BreadcrumbPage>
              )}
            </BreadcrumbItem>
          </Fragment>
        ))}
      </BreadcrumbList>
    </Breadcrumb>
  )
}
