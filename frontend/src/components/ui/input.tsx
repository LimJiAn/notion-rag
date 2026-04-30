import type { InputHTMLAttributes } from "react";

import { cn } from "../../lib/utils";

export function Input({ className, ...props }: InputHTMLAttributes<HTMLInputElement>) {
  return (
    <input
      className={cn(
        "flex h-10 w-full rounded-md border border-ink-200 bg-white px-3 text-sm text-ink-900 transition placeholder:text-ink-300 focus:border-ember-500",
        className,
      )}
      {...props}
    />
  );
}
