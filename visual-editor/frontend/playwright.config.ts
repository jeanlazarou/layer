import { defineConfig, devices } from '@playwright/test'

// End-to-end tests of the editor in a real browser.
//
//   pnpm test:e2e                                    build, serve dist/, run against it
//   E2E_BASE_URL=https://…/layer/ pnpm test:e2e      run against a deployed copy
//
// Tests navigate to './' so they also work under a sub-path (GitHub Pages).
// A trailing slash matters: without it, './' resolves to the parent path.
const deployedURL = process.env.E2E_BASE_URL?.replace(/\/?$/, '/')
const localURL = 'http://localhost:4173/'

export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  expect: { timeout: 20_000 },
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: deployedURL ?? localURL,
    trace: 'retain-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      // Wide enough that the "who answers each call?" panel is open by default.
      use: { ...devices['Desktop Chrome'], viewport: { width: 1440, height: 900 } },
    },
  ],
  webServer: deployedURL
    ? undefined
    : {
        command: 'vite build && vite preview --port 4173 --strictPort',
        url: localURL,
        reuseExistingServer: !process.env.CI,
        timeout: 120_000,
      },
})
