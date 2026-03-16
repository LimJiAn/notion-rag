import type { InputHTMLAttributes } from "react";

import { cn } from "../../lib/utils";

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "flex h-11 w-full rounded-2xl border border-ink-200 bg-white/80 px-4 text-sm text-ink-900 shadow-sm transition placeholder:text-ink-300 focus:border-ink-300 focus:bg-white",
        className,
      )}
      {...props}
    />
  );
}
