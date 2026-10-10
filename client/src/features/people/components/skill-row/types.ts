import type { Expertise, SkillInput } from '../../types'

export interface SkillRowProps {
  index: number
  skill: SkillInput
  /** The expertises this row may pick. */
  expertises: Expertise[]
  onChange: (index: number, skill: SkillInput) => void
  onRemove: (index: number) => void
}
