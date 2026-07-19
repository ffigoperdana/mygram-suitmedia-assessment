import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: true,
  retries: 0,
  outputDir: "test-results/mocked",
  reporter: [
    ["list"],
    ["html", { outputFolder: "playwright-report/mocked", open: "never" }],
  ],
  use: {
    baseURL: "http://127.0.0.1:3000",
    trace: "retain-on-failure",
  },
  webServer: {
    command: "npm run dev -- --host 127.0.0.1 --port 3000",
    url: "http://127.0.0.1:3000",
    reuseExistingServer: true,
    timeout: 120_000,
  },
  projects: [
    {
      name: "chrome-desktop",
      use: {
        ...devices["Desktop Chrome"],
      },
    },
    {
      name: "chrome-mobile",
      use: {
        ...devices["Pixel 7"],
      },
    },
    {
      name: "chrome-tablet",
      use: {
        viewport: { width: 834, height: 1194 },
        deviceScaleFactor: 2,
        hasTouch: true,
        isMobile: false,
      },
    },
  ],
});
