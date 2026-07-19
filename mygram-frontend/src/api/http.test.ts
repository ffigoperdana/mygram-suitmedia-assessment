import { AxiosError, type InternalAxiosRequestConfig } from "axios";
import { beforeEach, describe, expect, it } from "vitest";

import { http } from "@/api/http";
import { useAuthStore } from "@/stores/auth-store";

function rejectedResponse(status: number, data: Record<string, string>) {
  return async (config: InternalAxiosRequestConfig) =>
    Promise.reject(
      new AxiosError("Request failed", undefined, config, undefined, {
        config,
        data,
        headers: {},
        status,
        statusText: status === 403 ? "Forbidden" : "Unauthorized",
      }),
    );
}

describe("http client auth handling", () => {
  beforeEach(() => {
    useAuthStore.getState().setSession("active-token", {
      id: 1,
      username: "figo",
      email: "figo@mygram.local",
      role: "user",
      status: "active",
    });
  });

  it("clears auth state and stores a banned-account notice on banned responses", async () => {
    await expect(
      http.get("/api/v1/me", {
        adapter: rejectedResponse(403, { message: "Account is banned" }),
      }),
    ).rejects.toThrow("Request failed");

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.token).toBeNull();
    expect(state.notice).toEqual({
      type: "banned",
      message: "Account is banned",
    });
  });

  it("clears auth state and stores an expired-session notice on 401 responses", async () => {
    await expect(
      http.get("/api/v1/me", {
        adapter: rejectedResponse(401, { message: "Token expired" }),
      }),
    ).rejects.toThrow("Request failed");

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(false);
    expect(state.notice?.type).toBe("session-expired");
  });
});
