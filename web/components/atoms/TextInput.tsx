import { forwardRef, type InputHTMLAttributes } from "react";

export const TextInput = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  function TextInput({ className = "", ...props }, ref) {
    return (
      <input
        ref={ref}
        className={`w-full rounded-md border border-hill bg-transparent px-3 py-1.5 text-sm text-ink placeholder:text-sage/60 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-sage ${className}`}
        {...props}
      />
    );
  },
);
