import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./test/e2e",
  timeout: 30_000,
  expect: { timeout: 8_000 },
  fullyParallel: false,
  // The acceptance suite intentionally shares one local account, database, and
  // governance quota. Parallel workers would race on those real product states.
  workers: 1,
  reporter: "list",
  use: {
    baseURL: process.env.PASTE_E2E_BASE_URL || "http://localhost:3010",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile", use: { ...devices["Pixel 7"] } },
  ],
});
