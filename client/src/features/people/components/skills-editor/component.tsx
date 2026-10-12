import { Plus } from 'lucide-react'
import { useTranslation } from '@/shared/i18n'
import { Button } from '@/shared/ui/button'
import { Label } from '@/shared/ui/label'
import type { SkillInput } from '../../types'
import { SkillRow } from '../skill-row'
import type { SkillsEditorProps } from './types'
import { addSkill, optionsForRow, removeSkill, replaceSkill } from './utils'

/** Edits a person's skills: expertise, level 1–5 and which one is primary. */
export function SkillsEditor({ skills, expertises, onChange }: SkillsEditorProps) {
  const { t } = useTranslation()
  const canAdd = skills.length < expertises.length

  const handleSkillChange = (index: number, skill: SkillInput) => onChange(replaceSkill(skills, index, skill))
  const handleRemove = (index: number) => onChange(removeSkill(skills, index))
  const handleAdd = () => {
    const next = addSkill(skills, expertises)
    if (next) onChange(next)
  }

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <Label>{t('Expertise')}</Label>
        <Button type="button" variant="outline" size="sm" onClick={handleAdd} disabled={!canAdd}>
          <Plus />
          {t('Add expertise')}
        </Button>
      </div>
      {skills.length === 0 && <p className="text-sm text-muted-foreground">{t('Add at least one expertise.')}</p>}
      {skills.map((skill, index) => (
        <SkillRow
          key={skill.expertise_id}
          index={index}
          skill={skill}
          expertises={optionsForRow(skills, expertises, index)}
          onChange={handleSkillChange}
          onRemove={handleRemove}
        />
      ))}
    </div>
  )
}
