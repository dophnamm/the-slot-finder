import type { PlainKey } from '@/shared/i18n'

export interface ProgramPreviewProps {
  programId: string
}

/** One figure of the preview's top line. */
export interface PreviewFigure {
  label: PlainKey
  hours: number
  warning: boolean
}
