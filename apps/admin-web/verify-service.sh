#!/bin/bash
echo "=== Admin Web Service Verification ==="
echo ""
echo "1. Vite Server Status:"
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/ && echo "(OK)"
echo ""
echo "2. Route Availability:"
for r in login dashboard devices users medication alerts analytics settings; do
    hc=$(curl -s -o /dev/null -w "%{http_code} http://localhost:3000$r" 2>/dev/null || echo "?")
    printf "  /%-10s => %s\n" "$r" "$hc"
done
echo ""
echo "3. Build Artifacts exists:"
if [ -f apps/admin-web/dist/index.html ]; then
    ls -lh apps/admin-web/dist/index.html
    js_count=$(find apps/admin-web/dist/assets -name "*.js" | wc -l)
    css_count=$(find apps/admin-web/dist/assets -name "*.css" | wc -l)
    echo "  JavaScript assets: $js_count"
    echo "  CSS assets: $css_count"
else
    echo "  dist/index.html not found"
fi