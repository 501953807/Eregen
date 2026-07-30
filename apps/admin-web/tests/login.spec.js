import { test, expect } from '@playwright/test';

test.describe('Admin Web E2E Tests', () => {
  const setupAuth = async (page) => {
    await page.addInitScript(() => {
      localStorage.setItem('admin_token', 'fake-jwt-token');
      localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
    });
  };

  // Test: Login page renders correctly
  test('login page should render', async ({ page }) => {
    await page.goto('/login');
    // Wait for page content to appear (Vue renders after initial load)
    await page.waitForTimeout(3000);
    // Check that login form is present
    await expect(page.locator('form')).toBeVisible();
    await expect(page.locator('input[placeholder="请输入用户名"]')).toBeVisible();
    await expect(page.locator('button[type="primary"]')).toContainText('登录');
  });

  // Test: Unauthenticated user cannot access protected routes
  test('unauthenticated user blocked from dashboard', async ({ page }) => {
    await page.goto('/dashboard');
    // Should redirect to login (navigation guard in effect)
    // Wait a moment to see if we end up at login
    await page.waitForTimeout(2000);
    const url = await page.url();
    expect(url).toContain('/login');
  });

  // Test: Dashboard loads with authentication data
  test('dashboard shows data when authenticated', async ({ page }) => {
    await setupAuth(page);
    await page.goto('/dashboard');

    // Wait for main app content to mount - check for KPI text
    await page.waitForFunction(() => document.body.innerText.includes('在线设备'), {
      timeout: 30000
    });

    // Verify key dashboard elements are present
    await expect(page.getByText('在线设备')).toBeVisible();
    await expect(page.getByText('活跃家属')).toBeVisible();
    await expect(page.getByText('待处理告警')).toBeVisible();
  });

  // Test: Devices page loads
  test('devices page displays header', async ({ page }) => {
    await setupAuth(page);
    await page.goto('/devices');

    // Wait for page content to render
    await page.waitForFunction(() => document.body.innerText.includes('设备管理'), {
      timeout: 25000
    });

    // Verify header text is visible
    await expect(page.getByText('设备管理')).toBeVisible();
  });

  // Test: Users page loads
  test('users page displays header', async ({ page }) => {
    await setupAuth(page);
    await page.goto('/users');

    // Wait for page content to render
    await page.waitForFunction(() => document.body.innerText.includes('用户管理'), {
      timeout: 25000
    });

    // Verify header text is visible
    await expect(page.getByText('用户管理')).toBeVisible();
  });
});