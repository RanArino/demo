'use client';

import * as React from 'react';
import { cn } from '@/lib/utils';

interface ProgressProps extends React.HTMLAttributes<HTMLDivElement> {
  value?: number;
  max?: number;
  // Optional accessible name helpers
  label?: string;
  labelledBy?: string;
}

const Progress = React.forwardRef<HTMLDivElement, ProgressProps>(
  ({ className, value = 0, max = 100, label, labelledBy, ...props }, ref) => (
    <div
      ref={ref}
      role="progressbar"
      aria-valuemax={max}
      aria-valuemin={0}
      aria-valuenow={value}
      aria-label={label}
      aria-labelledby={labelledBy}
      className={cn(
        'relative h-4 w-full overflow-hidden rounded-full bg-secondary',
        className
      )}
      {...props}
    >
      <div
        className="h-full w-full flex-1 bg-primary transition-all"
        style={{
          transform: `translateX(-${100 - (value / max) * 100}%)`,
        }}
      />
    </div>
  )
);
Progress.displayName = 'Progress';

export { Progress };