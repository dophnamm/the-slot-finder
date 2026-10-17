/** The `data-tour` value of a sidebar link, e.g. "My schedule" → "nav-my-schedule". */
export function navTourAnchor(label: string) {
  return `nav-${label.toLowerCase().replace(/\s+/g, '-')}`
}
