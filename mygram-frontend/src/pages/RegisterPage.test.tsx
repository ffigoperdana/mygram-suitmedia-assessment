import { Route, Routes } from "react-router-dom";
import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { mygramApi } from "@/api/mygram";
import { RegisterPage } from "@/pages/RegisterPage";
import { renderWithProviders } from "@/test/test-utils";

describe("RegisterPage", () => {
  it("submits the registration fields", async () => {
    const register = vi.spyOn(mygramApi, "register").mockResolvedValue({
      id: 12,
      username: "figo",
      email: "figo@example.com",
      age: 25,
      role: "user",
      status: "active",
    });

    renderWithProviders(
      <Routes>
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/login" element={<p>Login page</p>} />
      </Routes>,
      { route: "/register" },
    );

    const user = userEvent.setup();
    await user.type(screen.getByLabelText("Username"), "figo");
    await user.type(screen.getByLabelText("Email"), "figo@example.com");
    await user.clear(screen.getByLabelText("Age"));
    await user.type(screen.getByLabelText("Age"), "25");
    await user.type(screen.getByLabelText("Password"), "secret123");
    await user.click(screen.getByRole("button", { name: "Create account" }));

    await waitFor(() => {
      expect(register).toHaveBeenCalledWith(
        {
          username: "figo",
          email: "figo@example.com",
          password: "secret123",
          age: 25,
        },
        expect.anything(),
      );
    });
  });
});
