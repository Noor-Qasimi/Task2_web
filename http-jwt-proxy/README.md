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
   bash scripts/generate-cert.sh
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

## Assignment answers

### 1. How can a service be run only on an internal network?

A service can be restricted to an internal network by placing it on a private container network and not publishing its container port to the host machine. In this project, Docker Compose creates an `internal` network with the setting `internal: true`. The legacy `go-httpbin` service is attached only to this network. Its port `8081` is declared with `expose`, not `ports`, so the port is available to other containers on the network but is not mapped to a host port.

The reverse proxy is connected to both the external-facing `edge` network and the private `internal` network. It can receive client requests through its published HTTP and HTTPS ports, then communicate with the legacy service using the internal service address `legacy:8081`. A client outside the Docker network cannot connect directly to the legacy service through a published host port. This reduces the service's network exposure and ensures that requests reach it through the proxy.

### 2. Explain the philosophy of TLS termination in addition to JWT authentication.

TLS termination means that the reverse proxy handles the client's TLS connection. The client connects to the proxy over HTTPS, and the proxy uses its TLS certificate and private key to establish an encrypted connection. TLS protects the request while it travels between the client and the proxy and helps the client verify the server's identity.

After receiving the HTTPS request, the proxy terminates the TLS connection and processes the HTTP request. JWT authentication is a separate security step: the proxy checks that the request contains a valid bearer token, verifies its HS256 signature, and validates its issuer, audience, expiration, and issued-at claim. Only after successful validation does the proxy forward the request to the legacy service. The Authorization header is removed before forwarding, so the original bearer token is not passed to the upstream application.

In this implementation, communication from the proxy to the legacy service uses HTTP over the private Docker network. Therefore, TLS is terminated at the proxy rather than being end-to-end between the client and the legacy service. This design centralizes certificate handling and authentication at the entry point. In a deployment where the internal network is not fully trusted or where policy requires encryption on every network segment, TLS should also be used between the proxy and the upstream service.

TLS and JWT solve different problems and complement each other: TLS protects data in transit and authenticates the server endpoint, while JWT lets the proxy authenticate and authorize requests at the application level. Using JWT does not replace the need for HTTPS.

### 3. Implement a JWT authentication reverse proxy for the legacy service.

This project implements the requested proxy in Go and uses `ghcr.io/mccutchen/go-httpbin` as the legacy HTTP service. Docker Compose runs the legacy service on the private `internal` network and exposes only the reverse proxy's HTTP and HTTPS ports to the host.

The implementation is divided into the following parts:

- **Reverse proxy and TLS listener — `cmd/proxy/main.go`:** starts the HTTPS server using the configured certificate and private key, and starts an HTTP server that redirects requests to HTTPS. It also provides a `/healthz` endpoint.
- **JWT authentication — `internal/auth/jwt.go`:** reads a bearer token from the Authorization header and accepts only HS256 tokens. It validates the signature, issuer, audience, expiration, and issued-at claim. Requests with missing or invalid tokens receive an HTTP `401 Unauthorized` response.
- **Upstream forwarding — `internal/proxy/reverse.go`:** forwards authenticated requests to the configured upstream service. It removes the Authorization header before forwarding and sets forwarding headers. If the upstream service is unavailable, it returns an HTTP `502 Bad Gateway` response.
- **Token utility — `cmd/token/main.go`:** provides a local command-line utility for creating short-lived JWTs for development and verification.
- **Container configuration — `compose.yaml`:** defines the proxy, the legacy service, and the optional token utility. The proxy joins both the `edge` and `internal` networks, while the legacy service joins only the private `internal` network.

To run the project, create a local `.env` file from `.env.example`, set a strong `JWT_SECRET` of at least 32 bytes, generate a development TLS certificate with `bash scripts/generate-cert.sh`, and start the containers with `docker compose up -d --build`. A request to the legacy route without a valid bearer token should be rejected. A request with a valid token should pass through the proxy to `go-httpbin`. The integration verification script is `bash scripts/verify.sh`, and the Go unit tests can be run with `go test ./...`.

The implementation keeps the legacy service off the host-published ports, enforces HTTPS for client traffic, and places JWT validation at the proxy boundary. The self-signed certificate is intended only for local development; a production deployment should use a certificate issued and managed appropriately and store the JWT signing secret securely.

## AI assistance

AI assistance was used to help review and organize the implementation and its documentation. The code and configuration should be reviewed and tested before deployment.
