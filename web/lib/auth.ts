"use client";

import { useEffect, useState } from "react";
import type { Session } from "@supabase/supabase-js";
import { useQuery } from "@tanstack/react-query";
import { supabase } from "./supabase";
import { api } from "./api";

// undefined while the initial session is still loading; null when signed out.
export function useSession() {
  const [session, setSession] = useState<Session | null | undefined>(undefined);

  useEffect(() => {
    supabase.auth.getSession().then(({ data }) => setSession(data.session));
    const { data } = supabase.auth.onAuthStateChange((_event, next) => setSession(next));
    return () => data.subscription.unsubscribe();
  }, []);

  return session;
}

// Asks the API whether this signed-in caller has a local account yet.
export function useMe(enabled: boolean) {
  return useQuery({ queryKey: ["me"], queryFn: () => api.me(), enabled });
}
