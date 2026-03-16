import type { TextareaHTMLAttributes } from "react";

import { cn } from "../../lib/utils";

export function Textarea({ className, ...props }: TextareaHTMLAttributes<HTMLTextAreaElement>) {
  return (
    <textarea
      className={cn(
        "flex min-h-[160px] w-full rounded-[24px] border border-ink-200 bg-white/80 px-4 py-4 text-sm text-ink-900 shadow-sm transition placeholder:text-ink-300 focus:border-ink-300 focus:bg-white",
        className,
      )}
      {...props}
    />
  );
}
