/** The typed done points, kept within 0..scope; null when the text is not a number of points. */
export function parsePoints(text: string, scope: number): number | null {
  const trimmed = text.trim()
  if (trimmed === '') return null
  const points = Number(trimmed)
  if (!Number.isFinite(points) || points < 0) return null
  return Math.min(points, scope)
}

/** Where the slider sits: the typed points, or the saved ones while the text is not a number. */
export function sliderPoints(text: string, saved: number, scope: number): number {
  return parsePoints(text, scope) ?? Math.min(saved, scope)
}

/** The slider reports one number, or a list with one thumb. */
export function firstValue(value: number | readonly number[]): number {
  return typeof value === 'number' ? value : value[0]
}
