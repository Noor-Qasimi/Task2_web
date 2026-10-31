#!/usr/bin/env bash
set -euo pipefail

mkdir -p certs
openssl req -x509 -newkey rsa:2048 -nodes \
  -keyout certs/tls.key \
  -out certs/tls.crt \
  -days 365 \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"

# Keep the private key readable only by its owner.
chmod 600 certs/tls.key
chmod 644 certs/tls.crt

printf '\nCreated certs/tls.crt and certs/tls.key (self-signed, for local testing only).\n'
