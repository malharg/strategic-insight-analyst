
"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import AuthForm from "@/components/ui/AuthForm";
import { useAuth } from "@/context/AuthContext";

export default function Home() {
  const { user, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && user) {
      router.push("/dashboard");
    }
  }, [user, loading, router]);

  if (loading) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center">
        <p>Loading session...</p>
      </main>
    );
  }

  // If user is already logged in, show a message while redirecting
  if (user) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center">
        <p>Redirecting to your dashboard...</p>
      </main>
    );
  }

  // If no user, show the authentication form
  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-4 bg-gray-50">
      <AuthForm />
    </main>
  );
}