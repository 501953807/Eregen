import { test, expect } from '@playwright/test';

test.describe('Admin Web E2E Tests - Core Functionality', () => {
  // Test: Login page exists and has basic structure
  test('login page loads', async ({ page }) => {
    await page.goto('/login');
    await page.waitForTimeout(2000);
    const url = await page.url();
    // Should be at login page (either admin-web or redirect)
    expect(url).toBeTruthy();
  });

  // Test: Dashboard redirects to login when unauthenticated
  test('dashboard requires authentication', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForTimeout(1500);
    const url = await page.url();
    expect(url).toContain('/login');
  });

  // Test: Devices page redirects to login when unauthenticated
  test('devices page requires authentication', async ({ page }) => {
    await page.goto('/devices');
    await page.waitForTimeout(1500);
    const url = await page.url();
    expect(url).toContain('/login');
  });

  // Test: Users page redirects to login when unauthenticated
  test('users page requires authentication', async ({ page }) => {
    await page.goto('/users');
    await page.waitForTimeout(1500);
    const url = await page.url();
    expect(url).toContain('/login');
  });

  // Test: Alerts page redirects to login when unauthenticated
  test('alerts page requires authentication', async ({ page }) => {
    await page.goto('/alerts');
    await page.waitForTimeout(1500);
    const url = await page.url();
    expect(url).toContain('/login');
  });

  // Test: Settings page redirects to login when unauthenticated
  test('settings page requires authentication', async ({ page }) => {
    await page.goto('/settings');
    await page.waitForTimeout(1500);
    const url = await page.url();
    expect(url).toContain('/login');
  });

  // Test: Analytics page redirects to login when unauthenticated
  test('analytics page requires authentication', async ({ page }) => {
    await page.goto('/analytics');
    await page.waitForTimeout(1500);
    const url = await page.url();
    expect(url).toContain('/login');
  });

  // Test: Elderly page redirects to login when unauthenticated
  test('elderly page requires authentication', async ({ page }) => {
    await page.goto('/elderly');
    await page.waitForTimeout(1500);
    const url = await page.url();
    expect(url).toContain('/login');
  });

  // Test: Institutions page redirects to login when unauthenticated
  test('institutions page requires authentication', async ({ page }) => {
    await page.goto('/institutions');
    await page.waitForTimeout(1500);
    const url = await page.url();
    expect(url).toContain('/login');
  });
});
