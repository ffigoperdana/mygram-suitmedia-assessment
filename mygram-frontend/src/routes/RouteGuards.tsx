import { Navigate, Outlet, useLocation } from "react-router-dom";

import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentUser } from "@/hooks/use-auth";
import { useAuthStore } from "@/stores/auth-store";

export function ProtectedRoute() {
  const location = useLocation();
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (!isAuthenticated) {
    return <Navigate to="/login" replace state={{ from: location }} />;
  }

  return <Outlet />;
}

export function AdminRoute() {
  const user = useAuthStore((state) => state.user);
  const currentUser = useCurrentUser();

  if (currentUser.isLoading && !user) {
    return <Skeleton className="h-64" />;
  }

  if ((currentUser.data ?? user)?.role !== "admin") {
    return <Navigate to="/feed" replace />;
  }

  return <Outlet />;
}

export function GuestRoute() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);

  if (isAuthenticated) {
    return <Navigate to="/feed" replace />;
  }

  return <Outlet />;
}
