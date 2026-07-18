#!/usr/bin/env bash
# Generate self-signed CA and client/server certificates for prototype TLS.
# Usage: ./scripts/gen-certs.sh [output_dir]
# Default output_dir: certs/
set -euo pipefail

OUT_DIR="${1:-certs}"
mkdir -p "$OUT_DIR"

CA_KEY="$OUT_DIR/ca.key"
CA_CRT="$OUT_DIR/ca.crt"
SERVER_KEY="$OUT_DIR/server.key"
SERVER_CRT="$OUT_DIR/server.crt"
CLIENT_KEY="$OUT_DIR/client.key"
CLIENT_CRT="$OUT_DIR/client.crt"

echo "=== Generating self-signed CA ==="
openssl genrsa -out "$CA_KEY" 2048 2>/dev/null
openssl req -new -x509 -days 3650 \
  -key "$CA_KEY" \
  -out "$CA_CRT" \
  -subj "/C=CN/O=Eregen/CN=Eregen Root CA" \
  2>/dev/null
echo "  CA cert written to $CA_CRT"

echo "=== Generating server (EMQX) certificate ==="
openssl genrsa -out "$SERVER_KEY" 2048 2>/dev/null
cat > "$OUT_DIR/server.cnf" <<EOF
[req]
distinguished_name = req_distinguished_name
req_extensions = v3_req
prompt = no

[req_distinguished_name]
C = CN
O = Eregen
CN = mqtt.eregen.dev

[v3_req]
subjectAltName = @alt_names

[alt_names]
DNS.1 = mqtt.eregen.dev
DNS.2 = localhost
IP.1 = 127.0.0.1
EOF

openssl req -new -key "$SERVER_KEY" \
  -out "$OUT_DIR/server.csr" \
  -config "$OUT_DIR/server.cnf" \
  2>/dev/null
openssl x509 -req -days 3650 \
  -in "$OUT_DIR/server.csr" \
  -CA "$CA_CRT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$SERVER_CRT" \
  -extfile "$OUT_DIR/server.cnf" -extensions v3_req \
  2>/dev/null
echo "  Server cert written to $SERVER_CRT"

echo "=== Generating client (firmware) certificate ==="
openssl genrsa -out "$CLIENT_KEY" 2048 2>/dev/null
openssl req -new -key "$CLIENT_KEY" \
  -out "$OUT_DIR/client.csr" \
  -subj "/C=CN/O=Eregen/CN=bracelet-device" \
  2>/dev/null
openssl x509 -req -days 3650 \
  -in "$OUT_DIR/client.csr" \
  -CA "$CA_CRT" -CAkey "$CA_KEY" -CAcreateserial \
  -out "$CLIENT_CRT" \
  2>/dev/null
echo "  Client cert written to $CLIENT_CRT"

echo "=== Cleanup ==="
rm -f "$OUT_DIR"/*.csr "$OUT_DIR"/*.srl "$OUT_DIR"/*.cnf

echo ""
echo "Generated certificates:"
ls -la "$OUT_DIR"/*.crt "$OUT_DIR"/*.key
echo ""
echo "Next steps:"
echo "  1. Copy ca.crt + server.crt + server.key to cloud/gateway/config/emqx-tls/"
echo "  2. Copy ca.crt to firmware/bracelet/entry/ (as CA for SSL)"
echo "  3. Update docker-compose.yml to mount certs into EMQX container"
