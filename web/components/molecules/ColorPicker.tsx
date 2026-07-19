import { SCOPE_COLORS } from "@/lib/colors";

export function ColorPicker({
  value,
  onChange,
}: {
  value: string;
  onChange: (color: string) => void;
}) {
  return (
    <div className="flex gap-1.5" role="radiogroup" aria-label="Scope colour">
      {SCOPE_COLORS.map((c) => (
        <button
          key={c}
          type="button"
          onClick={() => onChange(c)}
          aria-label={`Colour ${c}`}
          aria-pressed={value === c}
          className={`h-5 w-5 rounded-full ring-offset-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sage ${
            value === c ? "ring-2 ring-sage" : ""
          }`}
          style={{ backgroundColor: c }}
        />
      ))}
    </div>
  );
}
