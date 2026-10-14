/** `[0, 1, …, count - 1]`, e.g. to draw a number of placeholder rows. */
export function range(count: number): number[] {
  return Array.from({ length: Math.max(0, count) }, (_, index) => index)
}
