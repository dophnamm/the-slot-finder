import { createBrowserRouter } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { AuthCallbackPage } from '@/pages/auth-callback'
import { DraftsPage } from '@/pages/drafts'
import { EventsPage } from '@/pages/events'
import { IntakePage } from '@/pages/intake'
import { MySchedulePage } from '@/pages/my-schedule'
import { NotFoundPage } from '@/pages/not-found'
import { PeoplePage } from '@/pages/people'
import { PersonDetailPage } from '@/pages/person-detail'
import { ProgramDetailPage } from '@/pages/program-detail'
import { ProgramsPage } from '@/pages/programs'
import { ProjectDetailPage } from '@/pages/project-detail'
import { ScenarioDetailPage } from '@/pages/scenario-detail'
import { ScenariosPage } from '@/pages/scenarios'
import { StaffingDraftPage } from '@/pages/staffing-draft'
import { AppLayout } from './app-layout'
import type { BreadcrumbHandle, BreadcrumbKind } from './breadcrumbs'
import { HomeRoute } from './home-route'
import { RequireAuth } from './require-auth'
import { RequirePermission } from './require-permission'
import { RouteError } from './route-error'

/** The route's breadcrumb trail, read by Breadcrumbs in the app layout. */
const crumb = (breadcrumb: BreadcrumbKind): BreadcrumbHandle => ({ breadcrumb })

export const router = createBrowserRouter([
  {
    // Anything outside the app layout: the sign-in guard and the layout itself.
    errorElement: <RouteError fullPage />,
    children: [
      { path: ROUTES.authCallback, Component: AuthCallbackPage },
      {
        Component: RequireAuth,
        children: [
          {
            path: ROUTES.dashboard,
            Component: AppLayout,
            children: [
              {
                // A screen that throws replaces only itself; the sidebar and breadcrumbs stay.
                errorElement: <RouteError />,
                children: [
                  { index: true, Component: HomeRoute, handle: crumb('dashboard') },
                  {
                    element: <RequirePermission permission="programs:view" />,
                    children: [
                      { path: ROUTES.programs, Component: ProgramsPage, handle: crumb('programs') },
                      { path: ROUTES.programDetail, Component: ProgramDetailPage, handle: crumb('programDetail') },
                      { path: ROUTES.projectDetail, Component: ProjectDetailPage, handle: crumb('projectDetail') },
                    ],
                  },
                  {
                    element: <RequirePermission permission="intake:view" />,
                    children: [{ path: ROUTES.intake, Component: IntakePage, handle: crumb('intake') }],
                  },
                  {
                    element: <RequirePermission permission="staffing-drafts:view" />,
                    children: [
                      { path: ROUTES.drafts, Component: DraftsPage, handle: crumb('drafts') },
                      { path: ROUTES.staffingDraft, Component: StaffingDraftPage, handle: crumb('staffingDraft') },
                    ],
                  },
                  {
                    element: <RequirePermission permission="scenarios:view" />,
                    children: [
                      { path: ROUTES.scenarios, Component: ScenariosPage, handle: crumb('scenarios') },
                      { path: ROUTES.scenarioDetail, Component: ScenarioDetailPage, handle: crumb('scenarioDetail') },
                    ],
                  },
                  {
                    element: <RequirePermission permission="people:view" />,
                    children: [{ path: ROUTES.people, Component: PeoplePage, handle: crumb('people') }],
                  },
                  {
                    element: <RequirePermission permission="people:view" />,
                    children: [
                      { path: ROUTES.personDetail, Component: PersonDetailPage, handle: crumb('personDetail') },
                    ],
                  },
                  {
                    element: <RequirePermission permission="events:view" />,
                    children: [{ path: ROUTES.events, Component: EventsPage, handle: crumb('events') }],
                  },
                  {
                    element: <RequirePermission permission="my-schedule:view" />,
                    children: [{ path: ROUTES.mySchedule, Component: MySchedulePage, handle: crumb('mySchedule') }],
                  },
                  { path: ROUTES.notFound, Component: NotFoundPage },
                ],
              },
            ],
          },
        ],
      },
    ],
  },
])
