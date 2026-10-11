import { ROW_CONTROL_SELECTOR } from './constants'

/** Whether a click landed on a link, button or field inside the row. */
export function isInsideControl(target: EventTarget | null): boolean {
  return target instanceof Element && target.closest(ROW_CONTROL_SELECTOR) !== null
}

/** Whether the user was selecting text, which should not open the page. */
export function isSelectingText(): boolean {
  return (window.getSelection()?.toString() ?? '') !== ''
}
