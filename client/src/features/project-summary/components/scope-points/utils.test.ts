import { describe, expect, it } from 'vitest'
import { parsePoints, sliderPoints } from './utils'

describe('parsePoints', () => {
  it.each([
    ['12', 40, 12],
    ['12.5', 40, 12.5],
    [' 7 ', 40, 7],
    ['55', 40, 40],
    ['0', 40, 0],
  ])('reads %j as %d of %d', (text, scope, want) => {
    expect(parsePoints(text, scope)).toBe(want)
  })

  it.each(['', 'abc', '-1'])('rejects %j', (text) => {
    expect(parsePoints(text, 40)).toBeNull()
  })
})

describe('sliderPoints', () => {
  it('follows the typed value within the scope', () => {
    expect(sliderPoints('30', 10, 40)).toBe(30)
    expect(sliderPoints('90', 10, 40)).toBe(40)
  })

  it('keeps the saved value while the text is not a number', () => {
    expect(sliderPoints('', 10, 40)).toBe(10)
  })
})
