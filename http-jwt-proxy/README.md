# HTTP JWT Reverse Proxy

This project places a Go reverse proxy in front of a legacy HTTP service. The proxy terminates TLS, validates bearer JWTs, and forwards only authenticated requests. The legacy service is reachable only on a private Docker network and does not publish a host port.

## Architecture

```text
Client -- HTTPS + Bearer JWT --> Go reverse proxy
                                  | TLS termination
                                  | JWT validation
                                  v
                           Private Docker network
                                  |
                                  v
                           Legacy go-httpbin
```

## Components

- `cmd/proxy`: HTTPS listener, HTTP-to-HTTPS redirect, health endpoint, and middleware wiring.
- `cmd/token`: local utility for creating short-lived test tokens.
- `internal/auth`: HS256 JWT validation (signature, issuer, audience, expiry, and issued-at).
- `internal/proxy`: upstream forwarding; removes the Authorization header before forwarding.
- `compose.yaml`: proxy, legacy service, and optional token utility.
- `scripts/verify.sh`: integration checks for redirects, health, token rejection/acceptance, and backend isolation.

## Requirements

- Docker with Docker Compose
- Go 1.23 or later (only if running tests/building outside Docker)
- OpenSSL for creating a local development certificate

## Local setup

Run these commands from this project directory.

1. Create your local environment file:

   **PowerShell**
   ```powershell
   Copy-Item .env.example .env
   ```

   **Bash**
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and set `JWT_SECRET` to a private random value of at least 32 bytes. Do not commit the real `.env` file.

3. Create a local self-signed certificate and key:

   ```bash
   mkdir -p certs
   openssl req -x509 -newkey rsa:2048 -nodes -keyout certs/tls.key -out certs/tls.crt -days 365 -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
   ```

   These files are for local testing only. The `certs/` directory is excluded from Git.

4. Build and start:

   ```bash
   docker compose up -d --build
   docker compose ps
   ```

5. Check the health endpoint and redirect:

   ```bash
   curl -k -i https://localhost/healthz
   curl -i http://localhost/healthz
   ```

   The HTTPS health request should return `200 OK`; the HTTP request should return `308 Permanent Redirect`.

## JWT tests

Create a short-lived test token with the optional token utility:

```bash
docker compose run --rm token /bin/token -subject student -ttl 5m
```

A request without a token should be rejected:

```bash
curl -k -i https://localhost/get
```

A request with a valid token should be forwarded to the legacy service:

```bash
TOKEN="$(docker compose run --rm token /bin/token -subject student -ttl 5m)"
curl -k -i -H "Authorization: Bearer $TOKEN" "https://localhost/get?source=verification"
```

Run the integration checks:

```bash
bash scripts/verify.sh
```

Stop the services:

```bash
docker compose down
```

## Security notes

- Keep `.env`, generated certificates, and private keys out of source control.
- The self-signed certificate is for local development, not production.
- Use a strong, unique JWT signing secret and keep it in a secret manager in deployed environments.
- The example uses HS256; anyone who obtains the signing secret can create tokens.
- The legacy service has no published host port and is attached only to the internal Docker network.
- The proxy removes the bearer token before forwarding the request upstream.
- The `/healthz` endpoint is intentionally unauthenticated; routes to the legacy service require a valid token.

## Tests

Unit tests are provided alongside the authentication, proxy, and command packages. Run them with:

```bash
go test ./...
```

## AI assistance

AI assistance was used to help review and organize the implementation and its documentation. The code and configuration should be reviewed and tested before deployment.
