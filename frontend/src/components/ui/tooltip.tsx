"use client"

import * as React from "react"

/*
  No-op tooltip components. We keep the same export surface but render nothing,
  effectively removing tooltips without impacting call sites.
*/

type PropsWithChildren = React.PropsWithChildren<unknown>

const TooltipProvider = ({ children }: PropsWithChildren) => <>{children}</>

const Tooltip = ({ children }: PropsWithChildren) => <>{children}</>

const TooltipTrigger = React.forwardRef<HTMLElement, React.HTMLAttributes<HTMLElement> & { asChild?: boolean }>(
  ({ children }, _ref) => <>{children}</>
)
TooltipTrigger.displayName = "TooltipTrigger"

type TooltipContentProps = React.HTMLAttributes<HTMLDivElement> & {
  side?: 'top' | 'bottom' | 'left' | 'right' | string
}

const TooltipContent = React.forwardRef<HTMLDivElement, TooltipContentProps>(
  (_props, _ref) => null
)
TooltipContent.displayName = "TooltipContent"

export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider }