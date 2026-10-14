import type { Project } from '@/shared/programs'

/** The ids with `id` moved to where `targetId` is, as when a dragged row passes over another. */
export function moveTo(ids: string[], id: string, targetId: string): string[] {
  const target = ids.indexOf(targetId)
  if (id === targetId || target < 0 || !ids.includes(id)) return ids
  const rest = ids.filter((other) => other !== id)
  return [...rest.slice(0, target), id, ...rest.slice(target)]
}

/** The ids with `id` moved `delta` places up (negative) or down, stopping at either end. */
export function moveBy(ids: string[], id: string, delta: number): string[] {
  const from = ids.indexOf(id)
  const to = Math.min(ids.length - 1, Math.max(0, from + delta))
  if (from < 0 || from === to) return ids
  return moveTo(ids, id, ids[to])
}

/** The projects in a pending order while it is saved, else as the server listed them. */
export function orderedProjects(projects: Project[], order: string[] | null): Project[] {
  if (!order) return projects
  const place = new Map(order.map((id, index) => [id, index]))
  return [...projects].sort((a, b) => (place.get(a.id) ?? order.length) - (place.get(b.id) ?? order.length))
}
