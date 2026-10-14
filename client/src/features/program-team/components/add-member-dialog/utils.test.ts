import { describe, expect, it } from 'vitest'
import type { Person } from '@/shared/people'
import { candidateOptions } from './utils'

const person = (id: string, name: string, active = true, primary?: string) =>
  ({
    id,
    name,
    active,
    skills: primary ? [{ expertise_id: primary, expertise: primary, level: 3, primary: true }] : [],
  }) as unknown as Person

describe('candidateOptions', () => {
  it('lists active people outside the team by name, with their primary expertise', () => {
    const people = [
      person('u3', 'Chi Le', true, 'QA'),
      person('u1', 'Ada Ng'),
      person('u2', 'Bao Tran'),
      person('u4', 'Dung', false),
    ]
    expect(candidateOptions(people, ['u2'])).toEqual([
      { id: 'u1', name: 'Ada Ng', expertise: null },
      { id: 'u3', name: 'Chi Le', expertise: 'QA' },
    ])
  })
})
