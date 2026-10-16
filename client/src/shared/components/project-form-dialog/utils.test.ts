import { describe, expect, it } from 'vitest'
import type { Project } from '@/shared/programs'
import { addIsoWeeks, currentIsoWeek, isoWeekDate } from '@/utils/functions'
import type { ProjectFormValues } from './types'
import { canSave } from './utils'

const thisWeek = currentIsoWeek()
const values: ProjectFormValues = {
  name: 'Checkout',
  start_earliest: isoWeekDate(thisWeek, 1),
  deadline: isoWeekDate(addIsoWeeks(thisWeek, 10), 1),
  scope_points: '0',
  done_points: '0',
  lead_id: '',
}
const outside = isoWeekDate(addIsoWeeks(thisWeek, 157), 1)

describe('canSave window', () => {
  it('accepts dates inside the window', () => {
    expect(canSave(values)).toBe(true)
    expect(canSave({ ...values, deadline: isoWeekDate(addIsoWeeks(thisWeek, 156), 7) })).toBe(true)
  })

  it('rejects a deadline just outside the window', () => {
    expect(canSave({ ...values, deadline: outside })).toBe(false)
  })

  it('rejects a start just outside the window', () => {
    expect(canSave({ ...values, start_earliest: isoWeekDate(addIsoWeeks(thisWeek, -157), 1) })).toBe(false)
  })

  it('lets an unchanged out-of-window deadline through when editing', () => {
    const project = { start_earliest: values.start_earliest, deadline: outside } as Project
    expect(canSave({ ...values, deadline: outside }, project)).toBe(true)
  })
})
