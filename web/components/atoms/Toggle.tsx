"use client";

export function Toggle({
  on,
  label,
  ariaLabel,
  disabled,
  onToggle,
}: {
  on: boolean;
  label: string;
  ariaLabel: string;
  disabled?: boolean;
  onToggle: () => void;
}) {
  return (
    <button
      type="button"
      role="switch"
      aria-checked={on}
      aria-label={ariaLabel}
      onClick={onToggle}
      disabled={disabled}
      className="flex items-center gap-2 rounded-md px-1 py-1 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-sage disabled:opacity-50"
    >
      <span
        className={`relative h-5 w-9 rounded-full transition-colors ${on ? "bg-sage" : "bg-hill"}`}
      >
        <span
          className={`absolute left-0.5 top-0.5 h-4 w-4 rounded-full transition-transform ${
            on ? "translate-x-4 bg-paper" : "bg-sage"
          }`}
        />
      </span>
      <span className="font-mono text-[11px] uppercase tracking-widest text-sage">{label}</span>
    </button>
  );
}
