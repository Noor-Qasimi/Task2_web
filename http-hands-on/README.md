# HTTP Hands-on: JWT Reverse Proxy

This folder contains a local demonstration of an authentication gateway in front of the legacy `go-httpbin` service.

## Architecture

```text
Client (this computer)
        |
        | HTTP to 127.0.0.1:8080, with a Bearer JWT
        v
Envoy reverse proxy (container port 80)
        |
        | Only requests with a valid JWT are forwarded
        v
go-httpbin (container port 8080, private Docker network)
```

- **Internal-only backend:** `go-httpbin` has no published host port. It is reachable by service name from the Compose network.
- **Authentication:** Envoy's JWT authentication filter checks the signature, issuer, audience, and expiration before routing a request.
- **Local binding:** The proxy is published as `127.0.0.1:8080`, so this setup is intended for testing from the same computer. Envoy itself listens on port 80 inside its container.
- **Demo secret:** The example uses a fixed HS256 secret to make local testing straightforward. It is not suitable for production.

## Requirements

- Docker Desktop installed and running.
- Docker Compose v2 (included with current Docker Desktop).
- PowerShell on Windows for the test commands below.

## Start the services

Open PowerShell in the repository root, then run:

```powershell
Copy-Item .\http-hands-on\.env.example .\http-hands-on\.env
docker compose -f .\http-hands-on\compose.yaml up -d
docker compose -f .\http-hands-on\compose.yaml ps
```

The first start downloads the Envoy and go-httpbin images. If Docker Desktop's engine is not running, start/fix Docker Desktop before continuing.

View logs:

```powershell
docker compose -f .\http-hands-on\compose.yaml logs -f proxy go-httpbin
```

Press Ctrl+C to stop following logs; this does not stop the containers.

## Test 1: no token

```powershell
curl.exe -i http://localhost:8080/get
```

Expected: an HTTP 401 response from the JWT filter. A request without a token must not reach the backend.

## Test 2: create a short-lived local JWT

Run the following in PowerShell. It creates a token locally; it does not send the secret to an external website.

```powershell
function ConvertTo-Base64Url([byte[]]$Bytes) {
    [Convert]::ToBase64String($Bytes).TrimEnd('=').Replace('+', '-').Replace('/', '_')
}

$header = @{ alg = "HS256"; typ = "JWT"; kid = "demo-key" } | ConvertTo-Json -Compress
$now = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
$payload = @{
    iss = "http-hands-on"
    aud = "http-hands-on-client"
    sub = "student-demo"
    iat = $now
    exp = $now + 600
} | ConvertTo-Json -Compress

$headerPart = ConvertTo-Base64Url([Text.Encoding]::UTF8.GetBytes($header))
$payloadPart = ConvertTo-Base64Url([Text.Encoding]::UTF8.GetBytes($payload))
$unsignedToken = "$headerPart.$payloadPart"
$key = [Text.Encoding]::UTF8.GetBytes("local-demo-secret-change-before-use")
$hmac = [System.Security.Cryptography.HMACSHA256]::new($key)
$signaturePart = ConvertTo-Base64Url($hmac.ComputeHash([Text.Encoding]::UTF8.GetBytes($unsignedToken)))
$token = "$unsignedToken.$signaturePart"
```

The token expires after ten minutes. Its issuer, audience, key ID, and signing secret must match the Envoy configuration.

## Test 3: valid token

Use the token created in the previous step:

```powershell
curl.exe -i -H "Authorization: Bearer $token" http://localhost:8080/get
```

Expected: HTTP 200 and a JSON response from go-httpbin.

## Test 4: invalid token

```powershell
curl.exe -i -H "Authorization: Bearer not-a-valid-token" http://localhost:8080/get
```

Expected: an HTTP 401 response. If the results differ, inspect the proxy logs and confirm the JWT filter started successfully.

## Stop the services

```powershell
docker compose -f .\http-hands-on\compose.yaml down
```

## Question 1: keep a service internal

Do not publish the legacy backend's port. In this Compose setup, only the proxy has a host port mapping, and the backend is attached to the internal Docker network. This reduces the chance that clients can bypass authentication by connecting directly to the legacy service.

For a real deployment, also enforce network rules at the host, cloud, or orchestration layer and avoid exposing the Docker host to untrusted networks.

## Question 2: TLS Termination

TLS Termination means the reverse proxy accepts an HTTPS connection, presents a certificate, and decrypts the request before forwarding it to the upstream service. It centralizes certificate management and can provide one place to enforce transport settings.

**This Compose example does not configure HTTPS/TLS certificates.** It uses HTTP on localhost for the JWT demonstration. For a production deployment, configure a TLS listener and a trusted certificate, and protect traffic between the proxy and backend when the network or policy requires it. Do not expose this HTTP-only demo directly to the public internet.

## Security limitations

- The embedded HS256 key is a demonstration key. Replace it and the matching JWKS value before any shared or deployed use.
- Keep signing secrets out of source control in real deployments. Use a managed secret store and rotate keys.
- A valid JWT must be signed by a trusted issuer. This example uses a shared HMAC secret, so anyone who obtains it can mint tokens.
- The sample only binds the published proxy port to localhost. Changing the binding to `0.0.0.0` would make it reachable on other interfaces and should not be done casually.
- The configuration has no HTTPS listener, rate limiting, production observability, or key rotation.

## Files

- `compose.yaml` — runs Envoy and go-httpbin on an internal Docker network.
- `proxy/config.yaml` — Envoy listener, JWT validation rules, and upstream cluster.
- `.env.example` — example local JWT settings; copy it to `.env` for reference.

## AI assistance

AI assistance was used to help organize the implementation notes and explain the configuration. Review the code, run the tests, and make sure you understand each component before submitting it.
