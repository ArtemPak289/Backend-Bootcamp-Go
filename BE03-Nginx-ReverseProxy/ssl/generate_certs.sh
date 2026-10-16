#!/bin/bash
set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$DIR"

echo "==> Generating OpenSSL configuration..."
cat << 'OPENSSL_CONF' > openssl.cnf
[req]
default_bits        = 2048
prompt              = no
default_md          = sha256
distinguished_name  = req_distinguished_name
req_extensions      = req_ext
x509_extensions     = v3_ca

[req_distinguished_name]
C  = RU
ST = Moscow
L  = Moscow
O  = School 21
OU = Backend Engineering
CN = appliance-store.local

[req_ext]
subjectAltName = @alt_names

[v3_ca]
subjectAltName = @alt_names
basicConstraints = critical, CA:true
keyUsage = critical, digitalSignature, keyEncipherment, keyCertSign

[alt_names]
DNS.1 = appliance-store.local
DNS.2 = store.local
DNS.3 = localhost
IP.1  = 127.0.0.1
OPENSSL_CONF

echo "==> Generating private key and self-signed certificate..."
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout server.key \
    -out server.crt \
    -config openssl.cnf

chmod 600 server.key
chmod 644 server.crt
rm -f openssl.cnf

echo "==> Certificate generated successfully at $DIR/server.crt and $DIR/server.key"
openssl x509 -in server.crt -text -noout | grep -E "(Issuer|Subject:|DNS:|IP Address:)"
