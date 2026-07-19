import { Route, Routes } from "react-router-dom";
import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";

import { AdminRoute, GuestRoute, ProtectedRoute } from "@/routes/RouteGuards";
import { useAuthStore } from "@/stores/auth-store";
import { renderWithProviders } from "@/test/test-utils";

const adminUser = {
  id: 1,
  username: "admin",
  email: "admin@mygram.local",
  age: 24,
  role: "admin" as const,
  status: "active" as const,
};

const regularUser = {
  ...adminUser,
  id: 2,
  username: "figo",
  email: "figo@mygram.local",
  role: "user" as const,
};

function renderGuardTree(route: string) {
  return renderWithProviders(
    <Routes>
      <Route element={<ProtectedRoute />}>
        <Route path="/feed" element={<p>Feed page</p>} />
        <Route element={<AdminRoute />}>
          <Route path="/admin" element={<p>Admin dashboard</p>} />
        </Route>
      </Route>
      <Route element={<GuestRoute />}>
        <Route path="/login" element={<p>Login page</p>} />
      </Route>
    </Routes>,
    { route },
  );
}

describe("RouteGuards", () => {
  beforeEach(() => {
    useAuthStore.getState().logout();
  });

  it("redirects anonymous users to login", async () => {
    renderGuardTree("/feed");

    expect(await screen.findByText("Login page")).toBeInTheDocument();
  });

  it("keeps normal users out of admin routes", async () => {
    useAuthStore.getState().setSession("regular-token", regularUser);

    renderGuardTree("/admin");

    expect(await screen.findByText("Feed page")).toBeInTheDocument();
    expect(screen.queryByText("Admin dashboard")).not.toBeInTheDocument();
  });

  it("allows admin users into admin routes after refresh", async () => {
    useAuthStore.getState().setSession("admin-token", adminUser);

    renderGuardTree("/admin");

    expect(await screen.findByText("Admin dashboard")).toBeInTheDocument();
  });

  it("redirects authenticated users away from guest pages", async () => {
    useAuthStore.getState().setSession("regular-token", regularUser);

    renderGuardTree("/login");

    expect(await screen.findByText("Feed page")).toBeInTheDocument();
  });
});
