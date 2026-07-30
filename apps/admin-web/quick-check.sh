#!/bin/bash
echo "Checking all admin-web routes..."
for route in /login /dashboard /devices /users /medication /alerts /analytics /settings; do
  http_code=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:3000$route" 2>/dev/null || echo "?")
  printf "  %-15s => %s\n" "$route" "$http_code"
done

# Check login page content has expected elements
echo ""
echo "=== Login Page Content Check ==="
login_page=$(curl -s http://localhost:3000/login)
if echo "$login_page" | grep -q "login-container"; then
  echo "  ✓ login-container found"
else
  echo "  ✗ login-container NOT found"
fi
if echo "$login_page" | grep -q '"用户名"'; then
  echo "  ✓ Chinese username label found"
else
  echo "  ✗ Chinese username label NOT found"
fi

# Check dashboard with auth simulated (just check it returns 200)
echo ""
echo "=== Dashboard Check ==="
dashboard=$(curl -s http://localhost:3000/dashboard)
if echo "$dashboard" | grep -q '颐贞'; then
  echo "  ✓ Dashboard loaded successfully (brand name found)"
else
  echo "  ✗ Dashboard might have issues"
fi