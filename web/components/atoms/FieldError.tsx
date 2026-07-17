// TanStack Form + Zod surface errors as objects with a `message`; older/string
// validators surface plain strings. Handle both.
export function FieldError({ errors }: { errors: unknown[] }) {
  const message = errors
    .map((e) => (typeof e === "string" ? e : (e as { message?: string })?.message))
    .filter(Boolean)
    .join(", ");

  if (!message) return null;
  return <p className="mt-1 text-xs text-alarm">{message}</p>;
}
