import type { Expertise, SkillInput } from '../../types'

export interface SkillsEditorProps {
  skills: SkillInput[]
  expertises: Expertise[]
  onChange: (skills: SkillInput[]) => void
}
