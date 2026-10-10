import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@/index.css'
import { startAuthSession } from '@/shared/auth'
import { AppProviders } from './app-providers'

// Restore the Keycloak session (if any) while the first screen renders.
void startAuthSession()

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <AppProviders />
  </StrictMode>,
)
