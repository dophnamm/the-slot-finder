import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { EVENTS_QUERY_KEY } from '@/shared/events'
import { translate } from '@/shared/i18n'
import { PROGRAMS_QUERY_KEY, PROJECTS_QUERY_KEY } from '@/shared/programs'
import { SCENARIOS_QUERY_KEY } from '@/shared/scenarios'
import { INSIGHTS_QUERY_KEY } from '@/utils/constants'
import { acceptDraft, editDraft, generateDraft, getDraft, listAlternatives, listOpenDrafts, rejectDraft } from './api'
import {
  alternativesQueryKey,
  DRAFT_SUPERSEDED,
  draftQueryKey,
  OPEN_DRAFTS_QUERY_KEY,
  PLAN_VERSION_CONFLICT,
} from './constants'
import type { Draft, DraftAcceptance, DraftEdit } from './types'
import { applySetHours } from './utils'

/** Open drafts for the inbox and the sidebar count; enabled only for roles that may see drafts. */
export function useOpenDrafts(enabled = true) {
  return useQuery({ queryKey: OPEN_DRAFTS_QUERY_KEY, queryFn: listOpenDrafts, enabled })
}

export function useDraft(id: string) {
  return useQuery({ queryKey: draftQueryKey(id), queryFn: () => getDraft(id) })
}

/** Who could take over an assignment; read when the swap dialog opens, so it is never served stale. */
export function useAlternatives(id: string, assignmentId: string) {
  return useQuery({
    queryKey: alternativesQueryKey(id, assignmentId),
    queryFn: () => listAlternatives(id, assignmentId),
    gcTime: 0,
  })
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
  const queryClient = useQueryClient()
  const store = useStoreDraft()
  return useMutation({
    mutationFn: (projectIds: string[]) => generateDraft(projectIds),
    onSuccess: async (draft) => {
      store(draft)
      await queryClient.invalidateQueries({ queryKey: OPEN_DRAFTS_QUERY_KEY })
    },
  })
}

export function useEditDraft(id: string) {
  const store = useStoreDraft()
  const reload = useReloadSuperseded(id)
  const queryClient = useQueryClient()
  const key = draftQueryKey(id)
  return useMutation({
    mutationFn: (edit: DraftEdit) => editDraft(id, edit),
    meta: { quietErrorCodes: [DRAFT_SUPERSEDED] },
    // Hour changes show at once, so the grid's colours follow the keyboard; the answer replaces them.
    onMutate: async (edit) => {
      const previous = queryClient.getQueryData<Draft>(key)
      if (previous && edit.set && !edit.add && !edit.remove) {
        await queryClient.cancelQueries({ queryKey: key })
        queryClient.setQueryData(key, applySetHours(previous, edit.set))
      }
      return { previous }
    },
    onSuccess: store,
    onError: (error, _edit, context) => {
      if (context?.previous) queryClient.setQueryData(key, context.previous)
      reload(error)
    },
  })
}

/** Accepting changes the baseline and project statuses, so insights, the change history,
 * project and program lists and scenarios (now stale) refetch. A version conflict is shown on the page, not as a toast. */
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
      await queryClient.invalidateQueries({ queryKey: SCENARIOS_QUERY_KEY })
      await queryClient.invalidateQueries({ queryKey: OPEN_DRAFTS_QUERY_KEY })
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
      await queryClient.invalidateQueries({ queryKey: OPEN_DRAFTS_QUERY_KEY })
    },
    onError: reload,
  })
}
