import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import { Save } from 'lucide-react'
import type { FormEvent } from 'react'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { Button } from './button'

afterEach(cleanup)

describe('Button', () => {
  it('shows a spinner while busy, stays focusable and ignores clicks', () => {
    const onClick = vi.fn()
    render(
      <Button isLoading onClick={onClick}>
        <Save />
        Save
      </Button>,
    )
    const button = screen.getByRole('button', { name: 'Save' })
    expect(button.getAttribute('aria-busy')).toBe('true')
    expect(button.getAttribute('aria-disabled')).toBe('true')
    expect(button.hasAttribute('disabled')).toBe(false)
    expect(button.querySelector('[data-slot=spinner]')).toBeTruthy()
    fireEvent.click(button)
    expect(onClick).not.toHaveBeenCalled()
  })

  it('keeps a busy submit button from submitting its form again', () => {
    const onSubmit = vi.fn((event: FormEvent<HTMLFormElement>) => event.preventDefault())
    render(
      <form onSubmit={onSubmit}>
        <Button type="submit" isLoading>
          Save
        </Button>
      </form>,
    )
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(onSubmit).not.toHaveBeenCalled()
  })

  it('stays busy rather than greyed out when the caller also disables it', () => {
    render(
      <Button isLoading disabled>
        Save
      </Button>,
    )
    expect(screen.getByRole('button', { name: 'Save' }).hasAttribute('disabled')).toBe(false)
  })

  it('works as before without isLoading', () => {
    const onClick = vi.fn()
    render(<Button onClick={onClick}>Save</Button>)
    fireEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(onClick).toHaveBeenCalledOnce()
    expect(screen.getByRole('button', { name: 'Save' }).getAttribute('aria-busy')).toBeNull()
  })
})
