"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { supabase } from "@/lib/supabase";

// supabase-js exchanges the ?code in the URL for a session on its own
// (detectSessionInUrl). We just wait for the session to land, then leave.
export default function AuthCallbackPage() {
  const router = useRouter();

  useEffect(() => {
    const { data } = supabase.auth.onAuthStateChange((_event, session) => {
      if (session) router.replace("/app");
    });
    supabase.auth.getSession().then(({ data: { session } }) => {
      if (session) router.replace("/app");
    });
    return () => data.subscription.unsubscribe();
  }, [router]);

  return <main className="flex flex-1 items-center justify-center text-sm text-sage">Signing you in…</main>;
}
