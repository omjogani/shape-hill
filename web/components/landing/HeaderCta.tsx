"use client";

import Link from "next/link";
import { useSession } from "@/lib/auth";

export function HeaderCta() {
  const session = useSession();

  if (session === undefined) return null;

  const [href, label] = session ? ["/app", "Open the app"] : ["/login", "Sign in"];

  return (
    <Link
      href={href}
      className="rounded-md bg-[var(--text)] px-3.5 py-1.5 font-medium text-[var(--bg)] transition-opacity hover:opacity-90"
    >
      {label}
    </Link>
  );
}
