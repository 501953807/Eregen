import { test, expect } from '@playwright/test';

const BASE = 'http://localhost:3100';
const ADMIN_API = 'http://localhost:8085';

async function login(page) {
  await page.goto(`${BASE}/login`);
  await page.waitForSelector('#email', { timeout: 5000 });
  await page.fill('#email', 'admin@eregen.com');
  await page.fill('#password', 'Admin@123');
  await page.click('button[type="submit"]');
  // Wait for redirect to dashboard
  await page.waitForURL(url => url.pathname.includes('/dashboard') || url.pathname === '/', { timeout: 10000 });
}

test.describe('Admin Web - Full Page Verification', () => {
  test.beforeEach(async ({ page }) => {
    await login(page);
  });

  const pages = [
    { path: '/dashboard', name: 'Dashboard' },
    { path: '/devices', name: 'Devices' },
    { path: '/users', name: 'Users' },
    { path: '/alerts', name: 'Alerts' },
    { path: '/analytics', name: 'Analytics' },
    { path: '/elderly', name: 'Elderly' },
    { path: '/institutions', name: 'Institutions' },
    { path: '/settings', name: 'Settings' },
    { path: '/ota', name: 'OTA' },
    { path: '/persons', name: 'Persons' },
    { path: '/self', name: 'Self Chain' },
    { path: '/hospital', name: 'Hospital Chain' },
    { path: '/community', name: 'Community Chain' },
    { path: '/regulatory', name: 'Regulatory Dashboard' },
    { path: '/medication', name: 'Medication' },
    { path: '/medical', name: 'Medical Wristband' },
  ];

  for (const { path, name } of pages) {
    test(`${name} page loads`, async ({ page }) => {
      await page.goto(`${BASE}${path}`);
      await page.waitForTimeout(2000);
      const url = await page.url();
      expect(url).toContain(path);
      // Page should not show login redirect
      expect(url).not.toContain('/login');
    });
  }

  test('API endpoints return 200 (no 500 errors)', async ({ request }) => {
    const loginResp = await request.post(`${ADMIN_API}/api/v1/auth/login`, {
      data: { method: 'email', credential: 'admin@eregen.com', secret: 'Admin@123' },
    });
    const token = (await loginResp.json()).data.token;
    const headers = { Authorization: `Bearer ${token}` };

    const endpoints = [
      ['GET', `${ADMIN_API}/api/v1/admin/self/elderly`],
      ['GET', `${ADMIN_API}/api/v1/admin/alerts`],
      ['GET', `${ADMIN_API}/api/v1/admin/hospital/patients`],
      ['GET', `${ADMIN_API}/api/v1/admin/hospital/stats/overview`],
      ['GET', `${ADMIN_API}/api/v1/admin/devices`],
      ['GET', `${ADMIN_API}/api/v1/admin/users`],
      ['GET', `${ADMIN_API}/api/v1/admin/persons?page=1&page_size=10`],
      ['GET', `${ADMIN_API}/api/v1/admin/medical/wristbands`],
      ['GET', `${ADMIN_API}/api/v1/admin/self/elderly`],
      ['GET', `${ADMIN_API}/api/v1/admin/audit/logs?limit=10`],
      ['GET', `${ADMIN_API}/api/v1/admin/settings/system`],
      ['GET', `${ADMIN_API}/api/v1/admin/regulatory/dashboard/patient-overview`],
      ['GET', `${ADMIN_API}/api/v1/admin/regulatory/dashboard/patient-list`],
      ['GET', `${ADMIN_API}/api/v1/admin/self/medication-rules/exists_placeholder`],
    ];

    for (const [method, url] of endpoints) {
      const resp = await request.fetch(url, { headers, method });
      const status = resp.status();
      expect(status, `GET ${url}`).toBeLessThan(500);
    }
  });
});
