import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { EVENTS_QUERY_KEY } from '@/shared/events'
import { translate } from '@/shared/i18n'
import { PROGRAMS_QUERY_KEY, PROJECTS_QUERY_KEY } from '@/shared/programs'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { acceptDraft, editDraft, generateDraft, getDraft, rejectDraft } from './api'
import { DRAFT_SUPERSEDED, draftQueryKey, PLAN_VERSION_CONFLICT } from './constants'
import type { Draft, DraftAcceptance, DraftEdit } from './types'

export function useDraft(id: string) {
  return useQuery({ queryKey: draftQueryKey(id), queryFn: () => getDraft(id) })
}

/** Every draft call returns the draft as it is now; keep that copy instead of refetching. */
function useStoreDraft() {
  const queryClient = useQueryClient()
  return (draft: Draft) => queryClient.setQueryData(draftQueryKey(draft.id), draft)
}

/** When a change replaced the draft, says so and reloads it, so the page links to the newer version. */
function useReloadSuperseded(id: string) {
  const queryClient = useQueryClient()
  return (error: { code: string }) => {
    if (error.code !== DRAFT_SUPERSEDED) return
    toast.info(translate('A change produced a newer version of this draft.'))
    void queryClient.invalidateQueries({ queryKey: draftQueryKey(id) })
  }
}

export function useGenerateDraft() {
  const store = useStoreDraft()
  return useMutation({ mutationFn: (projectIds: string[]) => generateDraft(projectIds), onSuccess: store })
}

export function useEditDraft(id: string) {
  const store = useStoreDraft()
  const reload = useReloadSuperseded(id)
  return useMutation({
    mutationFn: (edit: DraftEdit) => editDraft(id, edit),
    meta: { quietErrorCodes: [DRAFT_SUPERSEDED] },
    onSuccess: store,
    onError: reload,
  })
}

/** Accepting changes the baseline and project statuses, so insights, the change history and
 * project and program lists refetch. A version conflict is shown on the page, not as a toast. */
export function useAcceptDraft(id: string) {
  const queryClient = useQueryClient()
  const store = useStoreDraft()
  const reload = useReloadSuperseded(id)
  return useMutation({
    mutationFn: (acceptance: DraftAcceptance) => acceptDraft(id, acceptance),
    meta: { quietErrorCodes: [PLAN_VERSION_CONFLICT, DRAFT_SUPERSEDED] },
    onSuccess: async (draft) => {
      store(draft)
      await queryClient.invalidateQueries({ queryKey: PROJECTS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: PROGRAMS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: INSIGHTS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: EVENTS_QUERY_KEY })
    },
    onError: (error) => {
      if (error.code === PLAN_VERSION_CONFLICT) void queryClient.invalidateQueries({ queryKey: draftQueryKey(id) })
      reload(error)
    },
  })
}

export function useRejectDraft(id: string) {
  const queryClient = useQueryClient()
  const store = useStoreDraft()
  const reload = useReloadSuperseded(id)
  return useMutation({
    mutationFn: () => rejectDraft(id),
    meta: { quietErrorCodes: [DRAFT_SUPERSEDED] },
    onSuccess: async (draft) => {
      store(draft)
      await queryClient.invalidateQueries({ queryKey: EVENTS_QUERY_KEY })
    },
    onError: reload,
  })
}
