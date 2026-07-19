import { Route, Routes } from "react-router-dom";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { mygramApi } from "@/api/mygram";
import { ProfilePage } from "@/pages/ProfilePage";
import { useAuthStore } from "@/stores/auth-store";
import { renderWithProviders } from "@/test/test-utils";

describe("ProfilePage", () => {
  it("updates the current user profile through /me", async () => {
    const currentUser = {
      id: 2,
      username: "figo",
      email: "figo@example.com",
      age: 25,
      role: "user" as const,
      status: "active" as const,
    };
    useAuthStore.getState().setSession("user-token", currentUser);
    vi.spyOn(mygramApi, "me").mockResolvedValue(currentUser);
    const updateMe = vi.spyOn(mygramApi, "updateMe").mockResolvedValue({
      ...currentUser,
      username: "figo-updated",
      email: "figo-updated@example.com",
      age: 26,
    });

    renderWithProviders(
      <Routes>
        <Route path="/profile" element={<ProfilePage />} />
      </Routes>,
      { route: "/profile" },
    );

    const user = userEvent.setup();
    const username = await screen.findByLabelText("Username");
    await user.clear(username);
    await user.type(username, "figo-updated");
    await user.clear(screen.getByLabelText("Email"));
    await user.type(screen.getByLabelText("Email"), "figo-updated@example.com");
    await user.clear(screen.getByLabelText("Age"));
    await user.type(screen.getByLabelText("Age"), "26");
    await user.click(screen.getByRole("button", { name: "Save profile" }));

    await waitFor(() => {
      expect(updateMe).toHaveBeenCalledWith({
        username: "figo-updated",
        email: "figo-updated@example.com",
        age: 26,
      });
    });

    expect(useAuthStore.getState().user?.email).toBe("figo-updated@example.com");
  });
});
