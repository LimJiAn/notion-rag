import type { TextareaHTMLAttributes } from "react";

import { cn } from "../../lib/utils";

export function Textarea({ className, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(
        "flex min-h-[120px] w-full rounded-md border border-ink-200 bg-white px-3 py-3 text-sm text-ink-900 transition placeholder:text-ink-300 focus:border-ember-500",
        className,
      )}
      {...props}
    />
  );
}
