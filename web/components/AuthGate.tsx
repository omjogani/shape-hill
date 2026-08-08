"use client";

import { usePathname, useRouter } from "next/navigation";
import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useMe, useSession } from "@/lib/auth";

const CALLBACK = "/auth/callback";

function Full({ children }: { children: React.ReactNode }) {
  return (
    <main className="flex flex-1 items-center justify-center text-sm text-sage">{children}</main>
  );
}

export function AuthGate({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const qc = useQueryClient();
  const session = useSession();
  const me = useMe(!!session);

  const onboarded = me.data?.onboarded;

  // On any user change (including sign-out) drop cached data so one account never
  // shows another's hills or onboarding state. Harmless on first load: nothing is
  // cached yet.
  const userId = session?.user?.id ?? null;
  useEffect(() => {
    qc.clear();
  }, [userId, qc]);

  useEffect(() => {
    if (pathname === CALLBACK || session === undefined) return;

    if (!session) {
      if (pathname !== "/login") router.replace("/login");
      return;
    }
    if (me.isLoading) return;

    if (onboarded === false && pathname !== "/onboarding") {
      router.replace("/onboarding");
    } else if (onboarded && (pathname === "/login" || pathname === "/onboarding")) {
      router.replace("/");
    }
  }, [pathname, session, me.isLoading, onboarded, router]);

  if (pathname === CALLBACK) return <>{children}</>;
  if (session === undefined || (session && me.isLoading)) return <Full>Loading…</Full>;

  if (!session) return pathname === "/login" ? <>{children}</> : <Full>Redirecting…</Full>;
  if (onboarded === false)
    return pathname === "/onboarding" ? <>{children}</> : <Full>Redirecting…</Full>;
  // Onboarded: keep them out of the auth-only pages until the redirect lands.
  if (pathname === "/login" || pathname === "/onboarding") return <Full>Redirecting…</Full>;
  return <>{children}</>;
}
