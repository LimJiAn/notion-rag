import type { HTMLAttributes } from "react";

import { cn } from "../../lib/utils";

export function Badge({ className, ...props }: HTMLAttributes<HTMLSpanElement>) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full bg-moss-100 px-2.5 py-1 font-display text-xs font-semibold text-moss-700",
        className,
      )}
      {...props}
    />
  );
}
