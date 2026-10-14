import * as React from 'react'
import { cn } from 'cn'
import { ChevronDownIcon, LoaderCircle } from 'lucide-react'

type NativeSelectProps = Omit<React.ComponentProps<'select'>, 'size'> & {
  size?: 'sm' | 'default'
  /** Its options are still loading: a spinner takes the arrow's place and it cannot be opened yet. */
  isLoading?: boolean
}

const SELECT_ICON_CLASS =
  'pointer-events-none absolute top-1/2 right-2.5 size-4 -translate-y-1/2 text-muted-foreground select-none'

function NativeSelect({ className, size = 'default', isLoading = false, disabled, ...props }: NativeSelectProps) {
  return (
    <div
      className={cn(
        'group/native-select relative w-fit has-[select:disabled]:not-has-[select[aria-busy=true]]:opacity-50',
        className,
      )}
      data-slot="native-select-wrapper"
      data-size={size}
    >
      <select
        data-slot="native-select"
        data-size={size}
        className="h-8 w-full min-w-0 appearance-none rounded-lg border border-input bg-transparent py-1 pr-8 pl-2.5 text-sm transition-colors outline-none select-none selection:bg-primary selection:text-primary-foreground placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 data-[size=sm]:h-7 data-[size=sm]:rounded-[min(var(--radius-md),10px)] data-[size=sm]:py-0.5 dark:bg-input/30 dark:hover:bg-input/50 dark:aria-invalid:border-destructive/50 dark:aria-invalid:ring-destructive/40"
        disabled={disabled || isLoading}
        aria-busy={isLoading || undefined}
        {...props}
      />
      {isLoading ? (
        <LoaderCircle
          className={cn(SELECT_ICON_CLASS, 'animate-spin motion-reduce:animate-none')}
          aria-hidden="true"
          data-slot="native-select-icon"
        />
      ) : (
        <ChevronDownIcon className={SELECT_ICON_CLASS} aria-hidden="true" data-slot="native-select-icon" />
      )}
    </div>
  )
}

function NativeSelectOption({ className, ...props }: React.ComponentProps<'option'>) {
  return (
    <option data-slot="native-select-option" className={cn('bg-[Canvas] text-[CanvasText]', className)} {...props} />
  )
}

function NativeSelectOptGroup({ className, ...props }: React.ComponentProps<'optgroup'>) {
  return (
    <optgroup
      data-slot="native-select-optgroup"
      className={cn('bg-[Canvas] text-[CanvasText]', className)}
      {...props}
    />
  )
}

export { NativeSelect, NativeSelectOptGroup, NativeSelectOption }
