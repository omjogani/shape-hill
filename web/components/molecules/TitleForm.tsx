"use client";

import { useForm } from "@tanstack/react-form";
import { z } from "zod";
import { useUpdateTitle } from "@/lib/hooks";
import { FieldError } from "../atoms/FieldError";

const schema = z.object({
  title: z.string().min(1, "Title can't be empty").max(120, "Keep it under 120 characters"),
});

function saveStatus(m: { isPending: boolean; isError: boolean; isSuccess: boolean }) {
  if (m.isPending) return "Saving…";
  if (m.isError) return "Couldn't save title";
  if (m.isSuccess) return "Saved";
  return "";
}

export function TitleForm({ slug, title }: { slug: string; title: string }) {
  const update = useUpdateTitle(slug);

  const form = useForm({
    defaultValues: { title },
    validators: { onChange: schema },
    onSubmit: async ({ value }) => {
      if (value.title !== title) await update.mutateAsync(value.title);
    },
  });

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit();
      }}
    >
      <form.Field name="title">
        {(field) => (
          <div>
            <input
              value={field.state.value}
              onChange={(e) => field.handleChange(e.target.value)}
              onBlur={() => {
                field.handleBlur();
                form.handleSubmit();
              }}
              aria-label="Hill title"
              className="w-full bg-transparent font-display text-3xl font-semibold tracking-tight text-ink focus:outline-none"
            />
            <FieldError errors={field.state.meta.errors} />
          </div>
        )}
      </form.Field>
      <p className="mt-1 h-4 font-mono text-xs text-sage">{saveStatus(update)}</p>
    </form>
  );
}
