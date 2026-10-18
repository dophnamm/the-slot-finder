import * as React from 'react'
import { Select as SelectPrimitive } from '@base-ui/react/select'
import { cn } from 'cn'
import { CheckIcon, ChevronDownIcon, LoaderCircle } from 'lucide-react'

type SelectOption = { value: string; label: React.ReactNode }

type SelectProps = Omit<SelectPrimitive.Trigger.Props, 'value' | 'onChange' | 'children'> &
  Pick<SelectPrimitive.Root.Props<string>, 'id' | 'name' | 'required' | 'disabled'> & {
    options: SelectOption[]
    /** An empty string means nothing chosen: it shows the option whose value is '' (the placeholder), if any. */
    value: string
    onValueChange: (value: string) => void
    size?: 'sm' | 'default'
    /** Its options are still loading: a spinner takes the arrow's place and it cannot be opened yet. */
    isLoading?: boolean
  }

const ICON_CLASS = 'pointer-events-none size-4 shrink-0 text-muted-foreground select-none'

/** An empty string is "no value" for the callers; Base UI calls it null. */
const toBase = (value: string) => (value === '' ? null : value)

function Select({
  options,
  value,
  onValueChange,
  id,
  name,
  required,
  disabled,
  size = 'default',
  isLoading = false,
  className,
  ...props
}: SelectProps) {
  const items = options.map((option) => ({ value: toBase(option.value), label: option.label }))
  const handleValueChange = (next: string | null) => onValueChange(next ?? '')

  return (
    <SelectPrimitive.Root
      id={id}
      name={name}
      required={required}
      disabled={disabled || isLoading}
      items={items}
      value={toBase(value)}
      onValueChange={handleValueChange}
    >
      <SelectPrimitive.Trigger
        data-slot="select-trigger"
        data-size={size}
        aria-busy={isLoading || undefined}
        className={cn(
          'flex h-8 w-fit min-w-0 items-center justify-between gap-2 rounded-lg border border-input bg-transparent py-1 pr-2.5 pl-2.5 text-left text-sm whitespace-nowrap transition-colors outline-none select-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 aria-busy:opacity-100 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 data-[size=sm]:h-7 data-[size=sm]:rounded-[min(var(--radius-md),10px)] data-[size=sm]:py-0.5 dark:bg-input/30 dark:hover:bg-input/50 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40',
          className,
        )}
        {...props}
      >
        <SelectPrimitive.Value
          data-slot="select-value"
          className="line-clamp-1 flex-1 data-placeholder:text-muted-foreground"
        />
        <SelectPrimitive.Icon className="flex">
          {isLoading ? (
            <LoaderCircle className={cn(ICON_CLASS, 'animate-spin motion-reduce:animate-none')} />
          ) : (
            <ChevronDownIcon className={ICON_CLASS} />
          )}
        </SelectPrimitive.Icon>
      </SelectPrimitive.Trigger>
      <SelectPrimitive.Portal>
        <SelectPrimitive.Positioner className="isolate z-50 outline-none" alignItemWithTrigger={false} sideOffset={4}>
          <SelectPrimitive.Popup
            data-slot="select-content"
            className="z-50 max-h-(--available-height) min-w-(--anchor-width) origin-(--transform-origin) overflow-x-hidden overflow-y-auto rounded-lg bg-popover p-1 text-popover-foreground shadow-md ring-1 ring-foreground/10 duration-100 outline-none data-[side=bottom]:slide-in-from-top-2 data-[side=top]:slide-in-from-bottom-2 data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:overflow-hidden data-closed:fade-out-0 data-closed:zoom-out-95"
          >
            <SelectPrimitive.List>
              {items.map((item) => (
                <SelectPrimitive.Item
                  key={item.value ?? ''}
                  value={item.value}
                  data-slot="select-item"
                  className="relative flex cursor-default items-center gap-1.5 rounded-md py-1 pr-8 pl-1.5 text-sm outline-hidden select-none data-highlighted:bg-accent data-highlighted:text-accent-foreground data-disabled:pointer-events-none data-disabled:opacity-50"
                >
                  <SelectPrimitive.ItemText>{item.label}</SelectPrimitive.ItemText>
                  <SelectPrimitive.ItemIndicator className="pointer-events-none absolute right-2 flex items-center justify-center">
                    <CheckIcon className="size-4" />
                  </SelectPrimitive.ItemIndicator>
                </SelectPrimitive.Item>
              ))}
            </SelectPrimitive.List>
          </SelectPrimitive.Popup>
        </SelectPrimitive.Positioner>
      </SelectPrimitive.Portal>
    </SelectPrimitive.Root>
  )
}

export { Select }
export type { SelectOption }
