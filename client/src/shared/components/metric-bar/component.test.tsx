import { cleanup, render, screen } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { MetricBar } from './component'

afterEach(cleanup)

describe('MetricBar', () => {
  it('shows a rounded percentage', () => {
    render(<MetricBar label="Staffing" value={79.6} />)
    expect(screen.getByText('80%')).toBeTruthy()
    expect(screen.getByText('80%').className).not.toContain('text-destructive')
  })

  it('shows an unknown value as a dash', () => {
    render(<MetricBar label="Staffing" value={null} />)
    expect(screen.getByText('—')).toBeTruthy()
  })

  it('marks a warning in red', () => {
    render(<MetricBar label="Scope" value={30} warning />)
    expect(screen.getByText('30%').className).toContain('text-destructive')
  })

  it('shows Time past 100 % as it is', () => {
    render(<MetricBar label="Time" value={150} warning />)
    expect(screen.getByText('150%')).toBeTruthy()
  })
})
