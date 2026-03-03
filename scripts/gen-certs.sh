#!/usr/bin/env bash
# ============================================
# Generate self-signed mTLS certificates for InfraWatch
# ============================================
set -euo pipefail

CERT_DIR="${1:-certs}"
DAYS=365
SUBJ_CA="/C=US/ST=CA/O=InfraWatch/CN=InfraWatch CA"
SUBJ_SERVER="/C=US/ST=CA/O=InfraWatch/CN=infrawatch-server"
SUBJ_AGENT="/C=US/ST=CA/O=InfraWatch/CN=infrawatch-agent"

mkdir -p "$CERT_DIR"
cd "$CERT_DIR"

echo "==> Generating CA key and certificate..."
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days $DAYS -key ca.key -out ca.crt -subj "$SUBJ_CA"

echo "==> Generating server key and certificate..."
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -subj "$SUBJ_SERVER"

cat > server_ext.cnf <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = infrawatch-server
IP.1 = 127.0.0.1
IP.2 = 0.0.0.0
EOF

openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out server.crt -days $DAYS -extfile server_ext.cnf

echo "==> Generating agent (client) key and certificate..."
openssl genrsa -out agent.key 2048
openssl req -new -key agent.key -out agent.csr -subj "$SUBJ_AGENT"

cat > agent_ext.cnf <<EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF

openssl x509 -req -in agent.csr -CA ca.crt -CAkey ca.key -CAcreateserial \
  -out agent.crt -days $DAYS -extfile agent_ext.cnf

# Cleanup CSR and extension files
rm -f *.csr *.cnf *.srl

echo ""
echo "==> Certificates generated in $CERT_DIR/:"
ls -la
echo ""
echo "Copy to /etc/infrawatch/certs/ on each machine."
