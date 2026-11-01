# JWT-Authenticated Reverse Proxy for a Legacy HTTP Service

## Project overview

This project places a Go reverse proxy in front of a legacy HTTP service. The proxy is the only service published to the host. It validates a JSON Web Token (JWT) before forwarding a request to the legacy service, which remains reachable only through Docker's private internal network.

The implementation uses `ghcr.io/mccutchen/go-httpbin` as the legacy service, as permitted by the assignment.

## 1. How can a service be run only on an internal network?

A service can be kept off the public Internet by placing it on a private network and not publishing its container port to the host.

In this project, Docker Compose provides that isolation:

- The `legacy` service is connected only to the Docker network named `internal`.
- The `internal` network is configured with `internal: true`, which makes it an isolated network.
- The legacy service listens on container port `8081` and declares that port with `expose`, but it has no `ports` mapping. Therefore, port `8081` is not published on the host.
- The `proxy` service is connected to both `edge` and `internal`. It is the only application service that publishes a host port.
- External clients connect to the proxy on host port `80`. After JWT validation, the proxy forwards accepted requests to `http://legacy:8081` over the internal Docker network.

The request path is:

```text
External client
     |
     | HTTP request to host port 80
     v
Go reverse proxy (container port 8080)
     |
     | Validate Authorization: Bearer <JWT>
     | Reject invalid requests with 401
     v
Docker internal network
     |
     | http://legacy:8081
     v
Legacy go-httpbin service
(container port 8081; no host port published)
```

`expose` and `ports` have different purposes. `expose` documents the container port used for communication between connected containers; it does not publish that port on the host. `ports` creates a host-to-container port mapping. The legacy service has no `ports` entry, so clients outside the Docker network cannot access it through `localhost:8081`.

The proxy is the controlled entry point: it is connected to the public-facing `edge` network and the private `internal` network, while the legacy service is connected only to `internal`.

## 2. Explain the philosophy of TLS Termination in addition to JWT authentication

TLS Termination is the point at which a reverse proxy accepts an HTTPS connection from a client, performs the TLS handshake, presents the server certificate, and decrypts the incoming traffic. The proxy can then inspect and process the HTTP request before routing it to an internal service.

TLS Termination and JWT authentication solve different security problems:

- **TLS protects data in transit.** It encrypts the connection between the client and the TLS endpoint and helps protect the request from being read or modified while travelling across the network.
- **JWT authenticates the request.** The proxy verifies the token's signature and claims to determine whether the caller presents a valid token. A valid JWT does not itself encrypt HTTP traffic.

Centralizing TLS at a reverse proxy is useful for legacy systems because the backend service does not need to implement HTTPS or manage certificates. The proxy can manage certificates and apply security controls in one place, then forward the request to the backend. If the internal network also requires encryption, the proxy-to-backend connection should use HTTPS or another protected transport.

**Implementation note:** The assignment specifically requires the proxy to receive external requests on port `80` and forward authenticated requests to the internal service on port `8081`. Therefore, this submitted configuration publishes HTTP on port `80` and does not enable an HTTPS listener or perform TLS Termination at runtime. TLS Termination is explained here as the security concept requested in Question 2; it is not claimed as an implemented feature of this port-80 configuration. In a production deployment, HTTPS should be enabled at the edge, while preserving the internal-network isolation and JWT checks.

## 3. Implement a JWT authentication proxy

### 3.1 Components and network configuration

- **Legacy service:** `ghcr.io/mccutchen/go-httpbin:latest`, configured to listen on port `8081` inside Docker.
- **Reverse proxy:** a Go HTTP server listening on container port `8080`, mapped to host port `80` by Docker Compose.
- **Internal network:** the `internal` Docker network is marked `internal: true`. The legacy service is attached only to this network.
- **Edge network:** the proxy is also attached to `edge`, which provides the network side used by the published host port.
- **Token utility:** the `token` service creates short-lived JWTs for development and verification. It is a helper tool, not a public token-issuing endpoint.

### 3.2 JWT authentication flow

For requests routed to the legacy service, the proxy expects an HTTP header in this format:

```http
Authorization: Bearer <token>
```

The authentication middleware validates the JWT using the HS256 signing algorithm and checks the configured issuer (`iss`), audience (`aud`), and required expiration (`exp`). The issued-at (`iat`) claim is also validated when present. A short clock-skew allowance is configured.

If the token is missing, malformed, expired, incorrectly signed, has the wrong issuer or audience, or uses an unsupported signing algorithm, the proxy returns `401 Unauthorized` and does not forward the request.

After successful validation, the proxy sets `X-Authenticated-Subject` from the verified JWT subject. The reverse proxy removes the `Authorization` header before forwarding the request, so the bearer token is not passed to the legacy service.

The `GET /healthz` endpoint is an unauthenticated liveness check for the proxy process. Requests to the legacy service, such as `/get`, require a valid JWT.

### 3.3 Run the project locally

Requirements: Docker Desktop with Docker Compose and a terminal.

Run these commands from the `http-jwt-proxy` project directory.

1. Create a local environment file.

   **PowerShell:**
   ```powershell
   Copy-Item .env.example .env
   ```

   **Git Bash:**
   ```bash
   cp .env.example .env
   ```

2. Open `.env` and replace the `JWT_SECRET` placeholder with a private random secret containing at least 32 bytes. For example, with OpenSSL:

   ```bash
   openssl rand -hex 32
   ```

   Keep the generated value private. Do not commit `.env` or share the secret.

3. Build and start the services:

   ```bash
   docker compose up -d --build
   docker compose ps
   ```

   The proxy should publish host port `80`. The legacy service should not have a host-published port.

4. Check the proxy health endpoint:

   ```bash
   curl -i http://localhost/healthz
   ```

   Expected result: `200 OK`.

5. Check that a request without a token is rejected:

   ```bash
   curl -i http://localhost/get
   ```

   Expected result: `401 Unauthorized`.

6. Generate a short-lived test token:

   ```bash
   docker compose run --rm token -subject student -ttl 5m
   ```

   Copy the printed token into the following command in place of `<TOKEN>`:

   ```bash
   curl -i -H "Authorization: Bearer <TOKEN>" \
     "http://localhost/get?source=verification"
   ```

   Expected result: `200 OK`. The JSON response from go-httpbin should contain the query parameter `source` with the value `verification`.

7. Confirm that the legacy service is not directly published on host port `8081`:

   ```bash
   curl -i --connect-timeout 3 http://localhost:8081/get
   ```

   The connection should fail unless another unrelated process on the host is using port `8081`.

### 3.4 Automated checks

With Docker running and the services started, execute:

```bash
bash scripts/verify.sh
```

The script generates valid and expired test tokens, checks the health endpoint, verifies that missing, malformed, and expired tokens are rejected, checks that a valid token reaches go-httpbin, and verifies that the legacy service is not directly published on host port `8081`.

Go unit tests can be run with:

```bash
go test ./...
```

These commands are provided for verification. The results must be observed in the local environment; this README does not claim that the tests have already passed on every machine.

### 3.5 Stop the services

```bash
docker compose down
```

## AI assistance disclosure

AI assistance was used to help review the implementation, identify differences between the configuration and the assignment requirements, and improve the README documentation. The model used was **GPT-6**. The model's parameter count is **not publicly disclosed**.
