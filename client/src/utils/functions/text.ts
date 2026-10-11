/**
 * Text folded for searching: lower case, no accents, Vietnamese đ as d, so "nguyen" finds "Nguyễn"
 * and "do" finds "Đỗ".
 */
export function toSearchText(value: string): string {
  return value
    .normalize('NFD')
    .replace(/\p{Diacritic}/gu, '')
    .replace(/đ/gi, 'd')
    .toLowerCase()
    .trim()
}
