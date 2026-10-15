import { QueryClientProvider } from '@tanstack/react-query'
import { RouterProvider } from 'react-router/dom'
import { queryClient } from '@/shared/api/query-client'
import { ConfirmDialog } from '@/shared/components/confirm-dialog'
import { Toaster } from '@/shared/ui/sonner'
import { router } from '../router'

export function AppProviders() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
      <ConfirmDialog />
      <Toaster richColors />
    </QueryClientProvider>
  )
}
