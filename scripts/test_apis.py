#!/usr/bin/env python3
"""Test all admin APIs after bug fixes"""
import requests, json, time

BASE = 'http://localhost:8089/api/v1'
r = requests.post(f'{BASE}/auth/login', json={'method':'email','credential':'admin@eregen.com','secret':'Admin@123'}, timeout=10)
token = r.json()['data']['token']
h = {'Authorization': f'Bearer {token}', 'Content-Type': 'application/json'}

tests = [
    ('GET', f'{BASE}/admin/community-wb/elders?page=1&page_size=5', 'community_elders'),
    ('GET', f'{BASE}/admin/health-records?personId=person-001&chain=self&limit=3', 'health_records'),
    ('GET', f'{BASE}/admin/medications?chain=self&page=1&page_size=5', 'medications'),
    ('GET', f'{BASE}/admin/alerts?page=1&page_size=5', 'alerts'),
    ('GET', f'{BASE}/admin/medical/patients?page=1&page_size=5', 'medical_patients'),
    ('GET', f'{BASE}/admin/persons?page=1&page_size=5', 'persons'),
]

for method, url, name in tests:
    r = requests.get(url, headers=h, timeout=10)
    if r.status_code == 200:
        d = r.json()
        data = d.get('data')
        count = len(data) if isinstance(data, list) else '?'
        print(f'OK   {name}: {count} records')
    else:
        print(f'ERR  {name}: {r.status_code} - {r.text[:100]}')
    time.sleep(0.3)

print('\nDone')
