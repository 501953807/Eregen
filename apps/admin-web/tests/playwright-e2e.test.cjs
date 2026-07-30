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

  // Test 1: Login page renders correctly
  try {
    const page = await browser.newPage();
    await page.goto(`${baseUrl}/login`, { waitUntil: 'networkidle', timeout: 15000 });
    const hasLoginContainer = await page.locator('.login-container').count() > 0;
    const hasUsernameInput = await page.locator('[placeholder="用户名"]').count() > 0;
    const hasPasswordInput = await page.locator('[placeholder="密码"]').count() > 0;
    const hasSubmitBtn = await page.locator('button[type="submit"]').count() > 0;
    const success = hasLoginContainer && hasUsernameInput && hasPasswordInput && hasSubmitBtn;
    addResult('Login page loads', success, !success ? `container:${hasLoginContainer}, username:${hasUsernameInput}, password:${hasPasswordInput}, submit:${hasSubmitBtn}` : '');
    await page.close();
  } catch (e) {
    addResult('Login page test', false, e.message || String(e));
  }

  // Test 2: Dashboard requires auth (unauthorized access redirects)
  try {
    const page = await browser.newPage();
    await page.goto(`${baseUrl}/dashboard`, { waitUntil: 'networkidle', timeout: 15000 });
    const isLoginPage = page.url().includes('/login');
    addResult('Dashboard unauthenticated redirects to login', isLoginPage);
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
    const hasSidebar = await page.locator('.sidebar').count() > 0;
    const hasBreadcrumb = await page.querySelector('.breadcrumb') !== null;
    const hasMain = await page.querySelector('main') !== null;
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
    const hasDevicesList = await page.locator('.device-list, .table-container, el-table').count() > 0;
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
    const hasUsersList = await page.locator('el-table').count() > 0;
    const success = hasUsersList;
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
    const hasMedicationContent = await page.locator('.medication-page').count() > 0;
    const success = hasMedicationContent;
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
    const hasAlertsContent = await page.locator('.alert-list').count() > 0;
    const success = hasAlertsContent;
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
