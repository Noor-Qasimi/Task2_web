# HTTP JWT Reverse Proxy

This project places a Go reverse proxy at the entry point to a legacy HTTP service. The proxy receives external requests on host port **80**, validates the request's bearer JWT, and forwards authenticated requests to the legacy service on the private Docker network at **port 8081**.

The legacy service is not published on a host port. It can be reached only by containers connected to the private `internal` network, including the reverse proxy.

## Architecture

```text
External client
     |
     | HTTP request to host port 80
     v
Go reverse proxy (container port 8080)
     |
     | JWT validation
     | forward only authenticated requests
     v
Private Docker network (internal)
     |
     | HTTP to legacy:8081
     v
Legacy go-httpbin service
```

## Components

- `cmd/proxy`: HTTP listener, health endpoint, and JWT middleware wiring.
- `cmd/token`: local utility for creating short-lived test tokens.
- `internal/auth`: HS256 JWT validation (signature, issuer, audience, expiry, and issued-at).
- `internal/proxy`: upstream forwarding; removes the Authorization header before forwarding.
- `compose.yaml`: proxy, legacy service, and optional token utility.
- `scripts/verify.sh`: integration checks for health, token rejection/acceptance, and backend isolation.

## Requirements

- Docker with Docker Compose
- Go 1.23 or later (only if running tests/building outside Docker)

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

3. Build and start the services:

   ```bash
   docker compose up -d --build
   docker compose ps
   ```

   The proxy is published on host port 80. If port 80 is already in use, set another available host port in `.env` using `PROXY_HTTP_PORT`.

4. Check the health endpoint:

   ```bash
   curl -i http://localhost/healthz
   ```

   The expected response is `200 OK`.

## JWT tests

Create a short-lived test token with the optional token utility:

```bash
docker compose run --rm token /bin/token -subject student -ttl 5m
```

A request without a token should be rejected:

```bash
curl -i http://localhost/get
```

A request with a valid token should be forwarded to the legacy service:

```bash
TOKEN="$(docker compose run --rm token /bin/token -subject student -ttl 5m)"
curl -i -H "Authorization: Bearer $TOKEN" "http://localhost/get?source=verification"
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

- The legacy service has no published host port and is attached only to the Docker network named `internal`, which is configured with `internal: true`.
- The proxy is the only service that publishes a client-facing port. It connects to both the `edge` and `internal` networks.
- Requests to legacy-service routes must contain a valid JWT. Missing or invalid tokens are rejected by the proxy.
- The proxy removes the bearer token from the request before forwarding it to the upstream service.
- The `/healthz` endpoint is intentionally unauthenticated and does not proxy to the legacy service.
- This assignment configuration uses HTTP on port 80 to match the specified request flow. HTTP does not encrypt traffic. For a real Internet-facing production deployment, HTTPS should be enabled at the proxy or at a trusted TLS-terminating load balancer in front of it.

## Tests

Unit tests are provided alongside the authentication, proxy, and command packages. Run them with:

```bash
go test ./...
```

## Assignment answers

### 1. How can a service be run only on an internal network?

A service can be restricted to an internal network by placing it on a private container network and not publishing its container port to the host machine. In this project, Docker Compose creates a network named `internal` with the setting `internal: true`. The legacy `go-httpbin` service is attached only to this network. Its port `8081` is declared with `expose`, not `ports`, so it is available to other containers on the network but is not mapped to a host port.

The reverse proxy is connected to both the external-facing `edge` network and the private `internal` network. It receives client requests through the published host port 80 and communicates with the legacy service using the internal service address `legacy:8081`. A client outside the Docker network cannot connect directly to the legacy service through a published host port. Therefore, external requests must pass through the proxy, where authentication is checked before they reach the legacy application.

### 2. Explain the philosophy of TLS termination in addition to JWT authentication.

TLS termination means that a reverse proxy handles the client's encrypted TLS connection. The client connects to the proxy using HTTPS, and the proxy uses its TLS certificate and private key to establish an encrypted connection. TLS protects data while it travels between the client and the proxy and helps the client verify the server's identity.

After receiving an HTTPS request, the proxy terminates the TLS connection and processes the request. JWT authentication is a separate security step: the proxy checks that the request contains a valid bearer token, verifies its signature, and validates its issuer, audience, and time-related claims. Only after successful validation does the proxy forward the request to the legacy service. These mechanisms have different purposes: TLS protects data in transit and authenticates the server endpoint, while JWT authentication verifies the identity and claims represented by the token. JWT does not replace HTTPS.

For this assignment's specified network flow, the proxy receives external HTTP requests on port 80 and forwards authenticated requests to the internal service on port 8081. Therefore, the current configuration does **not** establish a TLS connection and does not perform TLS termination. TLS termination is explained here as a complementary security concept; enabling it in this deployment would require an HTTPS listener and a certificate, normally on port 443, or a trusted TLS-terminating load balancer in front of the proxy. The assignment's requested port-80 behavior is kept in the implementation.

### 3. Implement a JWT authentication reverse proxy for the legacy service.

This project implements the requested reverse proxy in Go and uses `ghcr.io/mccutchen/go-httpbin` as the legacy HTTP service. Docker Compose publishes the proxy on host port 80 and connects it to the private Docker network. The legacy service listens on port 8081 inside that network and has no host-published port.

The implementation is divided into the following parts:

- **HTTP reverse proxy — `cmd/proxy/main.go`:** starts the HTTP server on the configured address (container port 8080, mapped to host port 80), provides the `/healthz` endpoint, and applies JWT middleware to requests that are forwarded to the legacy service.
- **JWT authentication — `internal/auth/jwt.go`:** reads a bearer token from the Authorization header and accepts only HS256 tokens. It validates the signature, issuer, audience, expiration, and issued-at claim. Requests with missing or invalid tokens receive an HTTP `401 Unauthorized` response.
- **Upstream forwarding — `internal/proxy/reverse.go`:** forwards authenticated requests to the configured upstream service, `http://legacy:8081`. It removes the Authorization header before forwarding. If the upstream service is unavailable, it returns an HTTP `502 Bad Gateway` response.
- **Token utility — `cmd/token/main.go`:** provides a local command-line utility for creating short-lived JWTs for development and verification.
- **Container configuration — `compose.yaml`:** defines the proxy, the legacy service, and the optional token utility. The proxy joins both the `edge` and `internal` networks, while the legacy service joins only the private `internal` network.

To run the project, create a local `.env` file from `.env.example`, set a strong `JWT_SECRET` of at least 32 bytes, and start the containers with `docker compose up -d --build`. The proxy receives requests on port 80. A request to a legacy route without a valid bearer token should be rejected with HTTP 401. A request with a valid token should pass through the proxy to `go-httpbin`. The integration verification script is `bash scripts/verify.sh`, and the Go unit tests can be run with `go test ./...`.

This implementation follows the requested network arrangement: the external request enters through the proxy on port 80, JWT authentication is checked at the proxy, and only authenticated requests are forwarded to the legacy service on internal port 8081. The legacy service is not directly exposed to external clients.

## AI assistance disclosure

AI assistance was used to help review the implementation, identify differences between the configuration and the assignment requirements, and improve the README documentation. The model used was **GPT-6**. The model's parameter count is **not publicly disclosed**. The implementation must be reviewed, tested, and understood by the student, who is responsible for explaining the code during the hand-in.
