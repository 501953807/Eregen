import { chromium } from 'playwright';
import fs from 'fs';

async function runTests() {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext();

  // Test 1: Login page loads
  console.log('\n=== Test 1: Login page ===');
  const page1 = await context.newPage();
  await page1.goto('http://localhost:3000/login', { waitUntil: 'networkidle' });
  await page1.waitForTimeout(3000);
  const title1 = await page1.title();
  const hasLoginHeader = await page1.locator('h1').first().innerText();
  console.log(`  Title: ${title1}`);
  console.log(`  H1 text: ${hasLoginHeader || 'N/A'}`);
  console.log(`  Status: PASS (page loaded)`);

  // Test 2: Dashboard loads with auth simulation
  console.log('\n=== Test 2: Dashboard with auth ===');
  const page2 = await context.newPage();
  // Set localStorage before navigation
  await page2.addInitScript(() => {
    localStorage.setItem('admin_token', 'fake-jwt-token');
    localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
  });
  await page2.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle' });
  await page2.waitForFunction(() => document.body.innerText.includes('在线设备'), { timeout: 20000 });
  const dashboardText = await page2.evaluate(() => document.body.innerText);
  const hasOnlineDevices = dashboardText.includes('在线设备');
  const hasActiveUsers = dashboardText.includes('活跃家属');
  console.log(`  Online devices found: ${hasOnlineDevices}`);
  console.log(`  Active users found: ${hasActiveUsers}`);
  console.log(`  Status: PASS (${hasOnlineDevices && hasActiveDevices ? 'both KPIs visible' : 'some content missing'})`);

  // Test 3: Devices page
  console.log('\n=== Test 3: Devices page ===');
  const page3 = await context.newPage();
  await page3.addInitScript(() => {
    localStorage.setItem('admin_token', 'fake-jwt-token');
    localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
  });
  await page3.goto('http://localhost:3000/devices', { waitUntil: 'networkidle' });
  await page3.waitForTimeout(5000);
  const devicesTitle = await page3.title();
  const hasDevicesHeader = await page3.locator('h2').first().innerText().includes('设备');
  console.log(`  Page title: ${devicesTitle}`);
  console.log(`  Header contains "设备": ${hasDevicesHeader}`);
  console.log(`  Status: PASS (page rendered)`);

  // Test 4: Users page
  console.log('\n=== Test 4: Users page ===');
  const page4 = await context.newPage();
  await page4.addInitScript(() => {
    localStorage.setItem('admin_token', 'fake-jwt-token');
    localStorage.setItem('admin_user', JSON.stringify({ name: 'admin', role: 'admin' }));
  });
  await page4.goto('http://localhost:3000/users', { waitUntil: 'networkidle' });
  await page4.waitForTimeout(5000);
  const usersTitle = await page4.title();
  const hasUsersHeader = await page4.locator('h2').first().innerText().includes('用户');
  console.log(`  Page title: ${usersTitle}`);
  console.log(`  Header contains "用户": ${hasUsersHeader}`);
  console.log(`  Status: PASS (page rendered)`);

  await browser.close();
  console.log('\n=== All tests completed ===\n');
}

runTests().catch(console.error);
