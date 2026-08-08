"use client";

import { supabase } from "@/lib/supabase";
import { Button } from "@/components/atoms/Button";

export default function LoginPage() {
  const signIn = () =>
    supabase.auth.signInWithOAuth({
      provider: "google",
      options: { redirectTo: `${window.location.origin}/auth/callback` },
    });

  return (
    <main className="flex flex-1 flex-col items-center justify-center gap-6 px-8">
      <div className="text-center">
        <p className="font-mono text-xs uppercase tracking-widest text-sage">shapehill</p>
        <h1 className="font-display text-3xl">Sign in</h1>
      </div>
      <Button onClick={signIn}>Continue with Google</Button>
    </main>
  );
}
