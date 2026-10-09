import type { ReactNode } from 'react'

export interface PagePlaceholderProps {
  title: string
  description: string
  action?: ReactNode
  children?: ReactNode
}
