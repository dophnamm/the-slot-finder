import { useState } from 'react'
import { type Project, useReorderProjects } from '@/shared/programs'
import type { ProjectRowReorder } from './components/project-row'
import { moveBy, moveTo, orderedProjects } from './utils'

/**
 * Dragging rows (or the arrow keys on a row's handle) to set a program's project order. The new order
 * shows at once and is kept until the list comes back with the saved priorities, or until saving fails.
 */
export function useProjectOrder(programId: string, projects: Project[]) {
  const reorder = useReorderProjects(programId)
  const [order, setOrder] = useState<string[] | null>(null)
  const [draggingId, setDraggingId] = useState<string | null>(null)
  // The row the dragged one last moved past; passing over it again does not move it back.
  const [overId, setOverId] = useState<string | null>(null)
  const rows = orderedProjects(projects, order)
  const ids = rows.map((project) => project.id)

  const clearOrder = () => setOrder(null)
  const save = (next: string[]) => {
    const unchanged = next.every((id, index) => id === projects[index]?.id)
    if (unchanged) {
      clearOrder()
      return
    }
    setOrder(next)
    reorder.mutate(next, { onSettled: clearOrder })
  }

  const handlers: Omit<ProjectRowReorder, 'dragging'> = {
    onDragStart: (id) => {
      setDraggingId(id)
      setOrder(ids)
    },
    onDragOver: (id) => {
      if (!draggingId || id === overId) return
      setOverId(id === draggingId ? null : id)
      if (id !== draggingId) setOrder(moveTo(ids, draggingId, id))
    },
    onDragEnd: () => {
      if (draggingId) save(ids)
      setDraggingId(null)
      setOverId(null)
    },
    onMove: (id, delta) => save(moveBy(ids, id, delta)),
  }

  return { rows, draggingId, handlers, saving: reorder.isPending }
}
