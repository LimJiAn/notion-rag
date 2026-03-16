import type { ButtonHTMLAttributes, PropsWithChildren } from "react";

import { cn } from "../../lib/utils";

type ButtonVariant = "default" | "secondary" | "outline" | "ghost";
type ButtonSize = "default" | "sm" | "lg";

type ButtonProps = PropsWithChildren<
  ButtonHTMLAttributes<HTMLButtonElement> & {
    variant?: ButtonVariant;
    size?: ButtonSize;
  }
>;

const variantClasses: Record<ButtonVariant, string> = {
  default:
    "bg-ink-900 text-white hover:bg-ink-700 disabled:bg-ink-300 disabled:text-white/80",
  secondary:
    "bg-ember-500 text-white hover:bg-700 disabled:bg-ember-100 disabled:text-ink-500",
  outline:
    "border border-ink-200 bg-white/70 text-ink-900 hover:bg-white disabled:text-ink-300",
  ghost: "bg-transparent text-ink-500 hover:bg-white/60 hover:text-ink-900",
};

const sizeClasses: Record<ButtonSize, string> = {
  default: "h-11 px-4 text-sm",
  sm: "h-9 px-3 text-sm",
  lg: "h-12 px-5 text-sm",
};

export function Button({
  className,
  variant = "default",
  size = "default",
  children,
  ...props
}: ButtonProps) {
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center gap-2 rounded-2xl font-display font-semibold transition disabled:cursor-not-allowed",
        variantClasses[variant],
        sizeClasses[size],
        className,
      )}
      {...props}
    >
      {children}
    </button>
  );
}
