import { lazy, Suspense } from "react";
import { Navigate, Route, Routes } from "react-router-dom";

import { PWAInstallPrompt } from "@/components/pwa/PWAInstallPrompt";
import { AppShell } from "@/components/layout/AppShell";
import { Skeleton } from "@/components/ui/skeleton";
import { AdminRoute, GuestRoute, ProtectedRoute } from "@/routes/RouteGuards";
import { LoginPage } from "@/pages/LoginPage";
import { RegisterPage } from "@/pages/RegisterPage";
import { useAuthStore } from "@/stores/auth-store";

const AdminDashboardPage = lazy(() =>
  import("@/pages/AdminDashboardPage").then((module) => ({
    default: module.AdminDashboardPage,
  })),
);
const ApiDocsPage = lazy(() =>
  import("@/pages/ApiDocsPage").then((module) => ({ default: module.ApiDocsPage })),
);
const DocsSwaggerPage = lazy(() =>
  import("@/pages/DocsSwaggerPage").then((module) => ({
    default: module.DocsSwaggerPage,
  })),
);
const FeedPage = lazy(() =>
  import("@/pages/FeedPage").then((module) => ({ default: module.FeedPage })),
);
const NotFoundPage = lazy(() =>
  import("@/pages/NotFoundPage").then((module) => ({ default: module.NotFoundPage })),
);
const PhotoDetailPage = lazy(() =>
  import("@/pages/PhotoDetailPage").then((module) => ({
    default: module.PhotoDetailPage,
  })),
);
const ProfilePage = lazy(() =>
  import("@/pages/ProfilePage").then((module) => ({ default: module.ProfilePage })),
);
const SocialLinksPage = lazy(() =>
  import("@/pages/SocialLinksPage").then((module) => ({
    default: module.SocialLinksPage,
  })),
);

function HomeRedirect() {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated);
  const isDocsHost = window.location.hostname.startsWith("docs.");

  if (isDocsHost) {
    return <Navigate to="/docs" replace />;
  }

  return <Navigate to={isAuthenticated ? "/feed" : "/login"} replace />;
}

export default function App() {
  return (
    <>
      <Suspense
        fallback={
          <main className="container py-6">
            <Skeleton className="h-64" />
          </main>
        }
      >
        <Routes>
          <Route path="/" element={<HomeRedirect />} />
          <Route path="/docs" element={<ApiDocsPage />} />
          <Route path="/docs/swagger" element={<DocsSwaggerPage />} />
          <Route path="/swagger" element={<DocsSwaggerPage />} />
          <Route path="/swagger/" element={<DocsSwaggerPage />} />
          <Route path="/swagger/index.html" element={<DocsSwaggerPage />} />

          <Route element={<GuestRoute />}>
            <Route path="/login" element={<LoginPage />} />
            <Route path="/register" element={<RegisterPage />} />
          </Route>

          <Route element={<ProtectedRoute />}>
            <Route element={<AppShell />}>
              <Route path="/feed" element={<FeedPage />} />
              <Route path="/photos/:photoId" element={<PhotoDetailPage />} />
              <Route path="/social-links" element={<SocialLinksPage />} />
              <Route path="/profile" element={<ProfilePage />} />
              <Route element={<AdminRoute />}>
                <Route path="/admin" element={<AdminDashboardPage />} />
              </Route>
            </Route>
          </Route>

          <Route path="*" element={<NotFoundPage />} />
        </Routes>
      </Suspense>
      <PWAInstallPrompt />
    </>
  );
}
