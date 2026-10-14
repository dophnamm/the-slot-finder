import type { Expertise, SkillInput } from '../../types'

export interface SkillsEditorProps {
  skills: SkillInput[]
  expertises: Expertise[]
  /** The expertise catalogue is still loading. */
  expertisesLoading: boolean
  onChange: (skills: SkillInput[]) => void
}
