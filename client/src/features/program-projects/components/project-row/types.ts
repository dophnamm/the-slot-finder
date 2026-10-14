import type { Project } from '@/shared/programs'

/** What a row needs to take part in reordering the program's projects. */
export interface ProjectRowReorder {
  /** This row is the one being dragged. */
  dragging: boolean
  onDragStart: (projectId: string) => void
  /** The dragged row passes over this one. */
  onDragOver: (projectId: string) => void
  onDragEnd: () => void
  /** Moves this row one place: -1 up, 1 down. */
  onMove: (projectId: string, delta: number) => void
}

export interface ProjectRowProps {
  project: Project
  /** Given when the user may change the order: a handle, and drag events on the row. */
  reorder?: ProjectRowReorder
}
