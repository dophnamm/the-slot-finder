import type { Expertise, SkillInput } from '../../types'
import { DEFAULT_SKILL_LEVEL } from './constants'

/** Replaces one skill; marking it primary makes every other skill secondary. */
export function replaceSkill(skills: SkillInput[], index: number, skill: SkillInput): SkillInput[] {
  return skills.map((current, i) => {
    if (i === index) return skill
    return skill.primary ? { ...current, primary: false } : current
  })
}

/** Removes one skill; when it was primary, the first remaining skill becomes primary. */
export function removeSkill(skills: SkillInput[], index: number): SkillInput[] {
  const next = skills.filter((_, i) => i !== index)
  if (skills[index]?.primary && next.length > 0) next[0] = { ...next[0], primary: true }
  return next
}

/** Adds the first expertise not used yet; the first skill is primary. Null when all are used. */
export function addSkill(skills: SkillInput[], expertises: Expertise[]): SkillInput[] | null {
  const unused = expertises.find((expertise) => !skills.some((skill) => skill.expertise_id === expertise.id))
  if (!unused) return null
  return [...skills, { expertise_id: unused.id, level: DEFAULT_SKILL_LEVEL, primary: skills.length === 0 }]
}

/** Expertises a row may pick: its own, plus those no other row uses. */
export function optionsForRow(skills: SkillInput[], expertises: Expertise[], index: number): Expertise[] {
  const usedElsewhere = new Set(skills.filter((_, i) => i !== index).map((skill) => skill.expertise_id))
  return expertises.filter((expertise) => !usedElsewhere.has(expertise.id))
}
