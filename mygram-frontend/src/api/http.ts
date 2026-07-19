import axios, { AxiosError } from "axios";

import type { ApiErrorBody } from "@/api/types";
import { useAuthStore } from "@/stores/auth-store";

const useSameOriginApi = import.meta.env.VITE_USE_SAME_ORIGIN_API === "true";
const configuredApiBaseUrl = useSameOriginApi
  ? ""
  : (import.meta.env.VITE_API_BASE_URL ?? "").replace(/\/$/, "");

export const apiBaseUrl = configuredApiBaseUrl;
export const apiDisplayBaseUrl =
  configuredApiBaseUrl || (typeof window !== "undefined" ? window.location.origin : "");

export const http = axios.create({
  baseURL: apiBaseUrl,
  headers: {
    "Content-Type": "application/json",
  },
});

http.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token;

  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }

  return config;
});

http.interceptors.response.use(
  (response) => response,
  (error: AxiosError<ApiErrorBody>) => {
    const message = getApiErrorMessage(error);

    if (error.response?.status === 401) {
      useAuthStore.getState().logout({
        type: "session-expired",
        message: "Your session expired. Sign in again to continue.",
      });
    }

    if (error.response?.status === 403 && message.toLowerCase().includes("banned")) {
      useAuthStore.getState().logout({
        type: "banned",
        message,
      });
    }

    return Promise.reject(error);
  },
);

export function getApiErrorMessage(error: unknown) {
  if (axios.isAxiosError<ApiErrorBody>(error)) {
    return (
      error.response?.data?.message ??
      error.response?.data?.error_message ??
      error.response?.data?.error ??
      error.response?.data?.err ??
      error.message
    );
  }

  return "Something went wrong";
}
