/// <reference types="vitest" />

import path from "node:path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        rewrite: (incomingPath) => incomingPath.replace(/^\/api/, ""),
      },
    },
  },
  preview: {
    port: 3000,
  },
  test: {
    environment: "jsdom",
    exclude: ["tests/e2e/**", "tests/e2e-real/**", "node_modules/**", "dist/**"],
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
    css: true,
  },
});
