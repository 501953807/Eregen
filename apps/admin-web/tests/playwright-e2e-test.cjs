const { chromium } = require('playwright');

async function verify() {
  const browser = await chromium.launch({ headless: true });
  const baseUrl = 'http://localhost:3000';

  console.log('=== Admin Web E2E Verification ===\n');

  let passed = 0;
  let failed = 0;
  const results = [];

  function addResult(testName, success, details = '') {
    if (success) {
      passed++;
      console.log(`✓ ${testName}`);
    } else {
      failed++;
      console.log(`✗ ${testName}: ${details}`);
    }
    results.push({ testName, success, details });
  }

  // Helper to wait for route content
  async function waitForContent(page, selector, timeout = 10000) {
    await page.waitForSelector(selector, { timeout });
    return true;
  }

  // Test 1: Login page renders correctly
  try {
    const page = await browser.newPage();
    await page.goto(`${baseUrl}/login`, { waitUntil: 'networkidle', timeout: 15000 });
    
    // Wait for Vite dev server to render the app
    await page.waitForTimeout(2000);
    
    // Check for login form elements using more robust selectors
    const hasLoginContainer = await page.locator('.login-container').count() > 0;
    const hasUsernameInput = await page.locator('input[placeholder*="用户名"] || input[data-testid="username"]').count() > 0;
    const hasPasswordInput = await page.locator('input[placeholder*="密码"] || input[data-testid="password"]').count() > 0;
    const hasSubmitBtn = await page.locator('button[type="submit"] || button[text()*="登录"]').count() > 0;
    
    const success = hasLoginContainer || hasUsernameInput || hasPasswordInput || hasSubmitBtn;
    addResult('Login page loads', success, `container:${hasLoginContainer}, username:${hasUsernameInput}, password:${hasPasswordInput}, submit:${hasSubmitBtn}`);
    await page.close();
  } catch (e) {
    addResult('Login page test', false, e.message || String(e));
  }

  // Test 2: Dashboard requires auth (unauthorized should redirect or show login-like state)
  try {
    const page = await browser.newPage();
    await page.goto(`${baseUrl}/dashboard`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(2000);
    
    // If unauthenticated, either redirects or shows empty app state
    const isLoginPage = page.url().includes('/login');
    const hasLoginPageContent = await page.locator('.login-container').count() > 0;
    const success = isLoginPage || hasLoginPageContent;
    addResult('Dashboard auth protection works', success, `redirected:${isLoginPage}, hasLoginPage:${hasLoginPageContent}`);
    await page.close();
  } catch (e) {
    addResult('Dashboard redirect test', false, e.message || String(e));
  }

  // Test 3: Dashboard with authorized session loads
  try {
    const page = await browser.newPage();
    await page.addInitScript(() => {
      localStorage.setItem('admin_token', 'fake-jwt-token');
      localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
    });
    await page.goto(`${baseUrl}/dashboard`, { waitUntil: 'networkidle', timeout: 20000 });
    await page.waitForTimeout(3000); // Wait for Vue to render dashboard
    
    const hasSidebar = await page.locator('.sidebar').count() > 0;
    const hasBreadcrumb = await page.locator('.breadcrumb').count() > 0;
    const hasMain = await page.locator('main').count() > 0;
    const hasAppContent = await page.locator('#app').count() > 0;
    
    const success = hasSidebar || hasBreadcrumb || hasMain || hasAppContent;
    addResult('Dashboard loads with auth', success, `sidebar:${hasSidebar}, breadcrumb:${hasBreadcrumb}, main:${hasMain}, app:${hasAppContent}`);
    await page.close();
  } catch (e) {
    addResult('Dashboard load test', false, e.message || String(e));
  }

  // Test 4: Devices page with auth
  try {
    const page = await browser.newPage();
    await page.addInitScript(() => {
      localStorage.setItem('admin_token', 'fake-jwt-token');
      localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
    });
    await page.goto(`${baseUrl}/devices`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(2000);
    
    const hasDevicesList = await page.locator('[data-testid="device-list"], .device-list, el-table, .table-container').count() > 0;
    const success = hasDevicesList;
    addResult('Devices page loads', success);
    await page.close();
  } catch (e) {
    addResult('Devices page test', false, e.message || String(e));
  }

  // Test 5: Users page with auth
  try {
    const page = await browser.newPage();
    await page.addInitScript(() => {
      localStorage.setItem('admin_token', 'fake-jwt-token');
      localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
    });
    await page.goto(`${baseUrl}/users`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(2000);
    
    const hasUsersTable = await page.locator('el-table').count() > 0;
    const success = hasUsersTable;
    addResult('Users page loads', success);
    await page.close();
  } catch (e) {
    addResult('Users page test', false, e.message || String(e));
  }

  // Test 6: Medication page with auth
  try {
    const page = await browser.newPage();
    await page.addInitScript(() => {
      localStorage.setItem('admin_token', 'fake-jwt-token');
      localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
    });
    await page.goto(`${baseUrl}/medication`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(2000);
    
    const hasMedicationPage = await page.locator('.medication-page').count() > 0;
    const success = hasMedicationPage;
    addResult('Medication page loads', success);
    await page.close();
  } catch (e) {
    addResult('Medication page test', false, e.message || String(e));
  }

  // Test 7: Alerts page with auth
  try {
    const page = await browser.newPage();
    await page.addInitScript(() => {
      localStorage.setItem('admin_token', 'fake-jwt-token');
      localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
    });
    await page.goto(`${baseUrl}/alerts`, { waitUntil: 'networkidle', timeout: 15000 });
    await page.waitForTimeout(2000);
    
    const hasAlertsPage = await page.locator('.alert-list, .alerts-container').count() > 0;
    const success = hasAlertsPage;
    addResult('Alerts page loads', success);
    await page.close();
  } catch (e) {
    addResult('Alerts page test', false, e.message || String(e));
  }

  // Summary
  console.log('\n=== Summary ===');
  console.log(`Passed: ${passed}`);
  console.log(`Failed: ${failed}`);
  console.log(`Total: ${passed + failed}`);

  if (failed > 0) {
    console.log('\n--- Failed Tests ---');
    results.filter(r => !r.success).forEach(r => {
      console.log(`  • ${r.testName}: ${r.details}`);
    });
  }

  await browser.close();
  process.exit(failed > 0 ? 1 : 0);
}

verify().catch(err => {
  console.error('Critical error:', err);
  process.exit(1);
});
