import { Menu } from 'lucide-react'
import { useState } from 'react'
import { Outlet } from 'react-router'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Sheet, SheetContent, SheetTitle } from '@/shared/ui/sheet'
import { Breadcrumbs } from '../breadcrumbs'
import { Sidebar } from './components/sidebar'
import { PAGE_AREA_CLASS } from './constants'

/** White sidebar and header around a gray page area; below lg the sidebar slides in from the left. */
export function AppLayout() {
  const { t } = useTranslation()
  const [isNavOpen, setIsNavOpen] = useState(false)

  const handleOpenNav = () => setIsNavOpen(true)
  const handleCloseNav = () => setIsNavOpen(false)

  return (
    <div className="flex h-svh overflow-hidden bg-canvas">
      <aside className="hidden w-60 shrink-0 lg:flex">
        <Sidebar />
      </aside>
      <Sheet open={isNavOpen} onOpenChange={setIsNavOpen}>
        <SheetContent side="left" showCloseButton={false} className="w-64 gap-0 p-0">
          <SheetTitle className="sr-only">{t('Navigation')}</SheetTitle>
          <Sidebar onNavigate={handleCloseNav} />
        </SheetContent>
      </Sheet>
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-16 shrink-0 items-center gap-2 border-b bg-background px-4 lg:px-6">
          <Button
            variant="ghost"
            size="icon-lg"
            className="lg:hidden"
            aria-label={t('Open navigation')}
            onClick={handleOpenNav}
          >
            <Menu />
          </Button>
          <Breadcrumbs />
        </header>
        <main className="flex min-h-0 flex-1 flex-col overflow-y-auto p-4 lg:p-6">
          <div className={PAGE_AREA_CLASS}>
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  )
}
