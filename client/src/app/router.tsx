import { createBrowserRouter } from 'react-router'
import { ROUTES } from '@/resources/routes'
import { DashboardPage } from '@/pages/dashboard'
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

export const router = createBrowserRouter([
  { path: ROUTES.login, Component: LoginPage },
  {
    path: ROUTES.dashboard,
    Component: AppLayout,
    children: [
      { index: true, Component: DashboardPage },
      { path: ROUTES.programs, Component: ProgramsPage },
      { path: ROUTES.programDetail, Component: ProgramDetailPage },
      { path: ROUTES.intake, Component: IntakePage },
      { path: ROUTES.projectDetail, Component: ProjectDetailPage },
      { path: ROUTES.staffingDraft, Component: StaffingDraftPage },
      { path: ROUTES.scenarios, Component: ScenariosPage },
      { path: ROUTES.people, Component: PeoplePage },
      { path: ROUTES.mySchedule, Component: MySchedulePage },
      { path: ROUTES.events, Component: EventsPage },
      { path: ROUTES.notFound, Component: NotFoundPage },
    ],
  },
])
