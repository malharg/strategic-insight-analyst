// frontend/app/page.tsx
"use client";

import AuthForm from "@/components/ui/AuthForm";
import { useAuth } from "@/context/AuthContext";
import { auth } from "@/lib/firebase";

export default function Home() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <main className="flex min-h-screen flex-col items-center justify-center p-24">
        <p>Loading...</p>
      </main>
    );
  }

  return (
    <main className="flex min-h-screen flex-col items-center justify-center p-24 bg-gray-50">
      {user ? (
        <div>
          <h1 className="text-2xl font-bold">Welcome, {user.email}!</h1>
          <p>You are logged in.</p>
          <button
            onClick={() => auth.signOut()}
            className="mt-4 px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Sign Out
          </button>
        </div>
      ) : (
        <AuthForm />
      )}
    </main>
  );
}
