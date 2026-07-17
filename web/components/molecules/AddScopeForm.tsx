"use client";

import { useForm } from "@tanstack/react-form";
import { z } from "zod";
import { useAddScope } from "@/lib/hooks";
import { Button } from "../atoms/Button";
import { TextInput } from "../atoms/TextInput";
import { FieldError } from "../atoms/FieldError";

const COLORS = ["#2F4C64", "#55704B", "#937129", "#6B4A63", "#3F5E63", "#8A5A2B"];

const schema = z.object({
  title: z.string().min(1, "Name the scope").max(80, "Keep it short"),
  color: z.string(),
});

export function AddScopeForm({ slug, nextSortOrder }: { slug: string; nextSortOrder: number }) {
  const add = useAddScope(slug);

  const form = useForm({
    defaultValues: { title: "", color: COLORS[0] },
    validators: { onChange: schema },
    onSubmit: async ({ value, formApi }) => {
      await add.mutateAsync({ title: value.title, color: value.color, sort_order: nextSortOrder });
      formApi.reset();
    },
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
      className="flex flex-col gap-2"
    >
      <div className="flex items-start gap-2">
        <form.Field name="title">
          {(field) => (
            <div className="flex-1">
              <TextInput
                placeholder="Add a scope…"
                value={field.state.value}
                onChange={(e) => field.handleChange(e.target.value)}
                onBlur={field.handleBlur}
              />
              <FieldError errors={field.state.meta.errors} />
            </div>
          )}
        </form.Field>

        <form.Subscribe selector={(s) => [s.canSubmit, s.isSubmitting] as const}>
          {([canSubmit, isSubmitting]) => (
            <Button type="submit" disabled={!canSubmit || isSubmitting}>
              {isSubmitting ? "Adding…" : "Add"}
            </Button>
          )}
        </form.Subscribe>
      </div>

      <form.Field name="color">
        {(field) => (
          <div className="flex gap-1.5" role="radiogroup" aria-label="Scope colour">
            {COLORS.map((c) => (
              <button
                key={c}
                type="button"
                onClick={() => field.handleChange(c)}
                aria-label={`Colour ${c}`}
                aria-pressed={field.state.value === c}
                className={`h-5 w-5 rounded-full ring-offset-2 ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-sage ${
                  field.state.value === c ? "ring-2 ring-sage" : ""
                }`}
                style={{ backgroundColor: c }}
              />
            ))}
          </div>
        )}
      </form.Field>

      {add.isError && <p className="text-xs text-alarm">Couldn&apos;t add scope. Try again.</p>}
    </form>
  );
}
