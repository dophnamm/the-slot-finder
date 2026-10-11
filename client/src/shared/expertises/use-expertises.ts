import { useQuery } from '@tanstack/react-query'
import { listExpertises } from './api'

const EXPERTISES_QUERY_KEY = ['expertises'] as const

/** The expertise catalogue; it rarely changes, so every screen shares one copy. */
export function useExpertises() {
  return useQuery({ queryKey: EXPERTISES_QUERY_KEY, queryFn: listExpertises })
}
