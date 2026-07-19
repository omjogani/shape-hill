"use client";

import { useForm } from "@tanstack/react-form";
import { z } from "zod";
import { useUpdateScope } from "@/lib/hooks";
import type { Scope } from "@/lib/api";
import { Button } from "../atoms/Button";
import { TextInput } from "../atoms/TextInput";
import { FieldError } from "../atoms/FieldError";
import { ColorPicker } from "./ColorPicker";

const schema = z.object({
  title: z.string().min(1, "Name the scope").max(80, "Keep it short"),
  color: z.string(),
});

export function EditScopeForm({
  slug,
  scope,
  onDone,
}: {
  slug: string;
  scope: Scope;
  onDone: () => void;
}) {
  const update = useUpdateScope(slug);

  const form = useForm({
    defaultValues: { title: scope.Title, color: scope.Color },
    validators: { onChange: schema },
    onSubmit: async ({ value }) => {
      await update.mutateAsync({ scopeId: scope.ID, title: value.title, color: value.color });
      onDone();
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
                autoFocus
                aria-label="Scope title"
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
              {isSubmitting ? "Saving…" : "Save"}
            </Button>
          )}
        </form.Subscribe>
        <Button type="button" variant="ghost" onClick={onDone}>
          Cancel
        </Button>
      </div>

      <form.Field name="color">
        {(field) => <ColorPicker value={field.state.value} onChange={field.handleChange} />}
      </form.Field>
    </form>
  );
}
