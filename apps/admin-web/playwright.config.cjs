module.exports = {
  timeout: 30000,

  use: {
    baseURL: 'http://localhost:3000',
    screenshot: 'only-on-failure',
    trace: 'retain-on-failure'
  },

  projects: [
    {
      name: 'Chrome',
      use: { browserName: 'chromium', launchOptions: { headless: true } }
    }
  ],

  // Only run .cjs and .ts test files
  testMatch: ['**/*.spec.{cjs,ts,js}']
}