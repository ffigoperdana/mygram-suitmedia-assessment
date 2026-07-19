import { create } from "zustand";
import { persist } from "zustand/middleware";

import type { AuthUser } from "@/api/types";

type AccountNotice = {
  type: "banned" | "session-expired";
  message: string;
} | null;

type JwtPayload = Partial<AuthUser> & {
  exp?: number;
};

type AuthState = {
  token: string | null;
  user: AuthUser | null;
  notice: AccountNotice;
  isAuthenticated: boolean;
  setSession: (token: string, user?: AuthUser | null) => void;
  setUser: (user: AuthUser) => void;
  setNotice: (notice: AccountNotice) => void;
  logout: (notice?: AccountNotice) => void;
};

function decodeJwtPayload(token: string): JwtPayload | null {
  try {
    const payload = token.split(".")[1];
    const decoded = window.atob(payload.replace(/-/g, "+").replace(/_/g, "/"));
    return JSON.parse(decoded) as JwtPayload;
  } catch {
    return null;
  }
}

function userFromToken(token: string): AuthUser | null {
  const payload = decodeJwtPayload(token);
  if (!payload?.id || !payload.username || !payload.email) {
    return null;
  }

  return {
    id: payload.id,
    username: payload.username,
    email: payload.email,
    age: payload.age,
    role: payload.role === "admin" ? "admin" : "user",
    status: payload.status === "banned" ? "banned" : "active",
  };
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      notice: null,
      isAuthenticated: false,
      setSession: (token, user) => {
        set({
          token,
          user: user ?? userFromToken(token),
          notice: null,
          isAuthenticated: true,
        });
      },
      setUser: (user) => {
        set({ user, isAuthenticated: true, notice: null });
      },
      setNotice: (notice) => {
        set({ notice });
      },
      logout: (notice = null) => {
        set({
          token: null,
          user: null,
          notice,
          isAuthenticated: false,
        });
      },
    }),
    {
      name: "mygram-auth",
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        notice: state.notice,
        isAuthenticated: state.isAuthenticated,
      }),
    },
  ),
);
