import { SKELETON_TEXT_WIDTHS } from '@/utils/constants'

/** A bar length for one cell, varied by row and column so no two neighbours match. */
export function skeletonWidth(row: number, column: number): string {
  return SKELETON_TEXT_WIDTHS[(row * 2 + column) % SKELETON_TEXT_WIDTHS.length]
}
