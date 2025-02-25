import { useEffect } from "react";
import auth from "../lib/auth";

export function AuthGuard({ children }: { children: React.ReactNode }) {
  useEffect(() => {
    if (!auth.isAuthenticated()) {
      window.location.href = "/login";
    }
  }, []);

  return <>{children}</>;
}
