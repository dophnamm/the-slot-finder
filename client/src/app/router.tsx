import { createBrowserRouter } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { AuthCallbackPage } from '@/pages/auth-callback'
import { EventsPage } from '@/pages/events'
import { IntakePage } from '@/pages/intake'
import { LoginPage } from '@/pages/login'
import { MySchedulePage } from '@/pages/my-schedule'
import { NotFoundPage } from '@/pages/not-found'
import { PeoplePage } from '@/pages/people'
import { ProgramDetailPage } from '@/pages/program-detail'
import { ProgramsPage } from '@/pages/programs'
import { ProjectDetailPage } from '@/pages/project-detail'
import { ScenariosPage } from '@/pages/scenarios'
import { StaffingDraftPage } from '@/pages/staffing-draft'
import { AppLayout } from './app-layout'
import { HomeRoute } from './home-route'
import { RequireAuth } from './require-auth'
import { RequirePermission } from './require-permission'

export const router = createBrowserRouter([
  { path: ROUTES.login, Component: LoginPage },
  { path: ROUTES.authCallback, Component: AuthCallbackPage },
  {
    Component: RequireAuth,
    children: [
      {
        path: ROUTES.dashboard,
        Component: AppLayout,
        children: [
          { index: true, Component: HomeRoute },
          {
            element: <RequirePermission permission="programs:view" />,
            children: [
              { path: ROUTES.programs, Component: ProgramsPage },
              { path: ROUTES.programDetail, Component: ProgramDetailPage },
              { path: ROUTES.projectDetail, Component: ProjectDetailPage },
            ],
          },
          {
            element: <RequirePermission permission="intake:view" />,
            children: [{ path: ROUTES.intake, Component: IntakePage }],
          },
          {
            element: <RequirePermission permission="staffing-drafts:review" />,
            children: [{ path: ROUTES.staffingDraft, Component: StaffingDraftPage }],
          },
          {
            element: <RequirePermission permission="scenarios:view" />,
            children: [{ path: ROUTES.scenarios, Component: ScenariosPage }],
          },
          {
            element: <RequirePermission permission="people:manage" />,
            children: [{ path: ROUTES.people, Component: PeoplePage }],
          },
          {
            element: <RequirePermission permission="events:view" />,
            children: [{ path: ROUTES.events, Component: EventsPage }],
          },
          {
            element: <RequirePermission permission="my-schedule:view" />,
            children: [{ path: ROUTES.mySchedule, Component: MySchedulePage }],
          },
          { path: ROUTES.notFound, Component: NotFoundPage },
        ],
      },
    ],
  },
])
