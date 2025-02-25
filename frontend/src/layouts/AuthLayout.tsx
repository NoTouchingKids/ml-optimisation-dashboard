// src/layouts/AuthLayout.tsx
import React, { useEffect } from "react";
import Navigation from "../components/Navigation";
import auth from "../lib/auth";

interface AuthLayoutProps {
  children: React.ReactNode;
}

export default function AuthLayout({ children }: AuthLayoutProps) {
  useEffect(() => {
    if (!auth.isAuthenticated()) {
      window.location.href = "/login";
    }
  }, []);

  return (
    <div className="min-h-screen bg-background">
      <Navigation />
      <main>{children}</main>
    </div>
  );
}
