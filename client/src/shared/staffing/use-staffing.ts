import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { PROGRAMS_QUERY_KEY, PROJECTS_QUERY_KEY } from '@/shared/programs'
import { acceptDraft, editDraft, generateDraft, getDraft, rejectDraft } from './api'
import { draftQueryKey, PLAN_VERSION_CONFLICT } from './constants'
import type { Draft, DraftAcceptance, DraftEdit } from './types'

export function useDraft(id: string) {
  return useQuery({ queryKey: draftQueryKey(id), queryFn: () => getDraft(id) })
}

/** Every draft call returns the draft as it is now; keep that copy instead of refetching. */
function useStoreDraft() {
  const queryClient = useQueryClient()
  return (draft: Draft) => queryClient.setQueryData(draftQueryKey(draft.id), draft)
}

export function useGenerateDraft() {
  const store = useStoreDraft()
  return useMutation({ mutationFn: (projectIds: string[]) => generateDraft(projectIds), onSuccess: store })
}

export function useEditDraft(id: string) {
  const store = useStoreDraft()
  return useMutation({ mutationFn: (edit: DraftEdit) => editDraft(id, edit), onSuccess: store })
}

/** Accepting changes project statuses, so project and program lists refetch. A version
 * conflict is shown on the page, not as a toast. */
export function useAcceptDraft(id: string) {
  const queryClient = useQueryClient()
  const store = useStoreDraft()
  return useMutation({
    mutationFn: (acceptance: DraftAcceptance) => acceptDraft(id, acceptance),
    meta: { quietErrorCodes: [PLAN_VERSION_CONFLICT] },
    onSuccess: async (draft) => {
      store(draft)
      await queryClient.invalidateQueries({ queryKey: PROJECTS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: PROGRAMS_QUERY_KEY })
    },
    onError: (error) => {
      if (error.code === PLAN_VERSION_CONFLICT) void queryClient.invalidateQueries({ queryKey: draftQueryKey(id) })
    },
  })
}

export function useRejectDraft(id: string) {
  const store = useStoreDraft()
  return useMutation({ mutationFn: () => rejectDraft(id), onSuccess: store })
}
