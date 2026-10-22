# HTTP Assignment

## Question 1 — HTTP Versions

### Key differences

- **HTTP/0.9:** The earliest, very simple version. It supported only `GET`, had no HTTP headers, and returned the resource body without a status line or response headers.
- **HTTP/1.0:** Added request and response headers, status codes, and support for methods such as `POST` and `HEAD`. Connections were generally closed after each request unless extra behavior was negotiated.
- **HTTP/1.1:** Introduced persistent connections by default, the `Host` header (important for hosting multiple websites on one IP address), chunked transfer encoding, and improved cache controls. Multiple requests could reuse a connection, although ordinary HTTP/1.1 did not multiplex responses in parallel on one connection.
- **HTTP/2:** Kept HTTP methods and status semantics but introduced binary framing, multiplexing multiple streams over one TCP connection, header compression (HPACK), and server push. This improves efficiency, though TCP packet loss can temporarily affect all streams on that connection.
- **HTTP/3:** Keeps HTTP semantics but runs over **QUIC**, which uses UDP. QUIC provides encrypted connections using TLS 1.3, stream multiplexing, and independent stream-level loss handling, reducing the cross-stream blocking associated with TCP.

### Why does HTTP/3 use UDP through QUIC?

HTTP/3 does not simply replace reliable TCP with unreliable UDP and leave reliability out. It uses **QUIC**, a transport protocol implemented over UDP that provides reliable delivery, congestion control, encryption, and independent streams. Using UDP lets QUIC implement these mechanisms without being constrained by TCP's single ordered byte stream. As a result, packet loss on one QUIC stream does not necessarily stop progress on other streams, and connection setup can be faster. UDP is therefore the carrier for QUIC, not the whole transport solution.

## Question 2 — Stateless and Stateful Communication

### A. Explain the distinction (maximum five lines)

A **stateless** interaction does not require the server to remember the previous request in order to understand the current one. Each request contains the information needed to process it. In a **stateful** interaction, the server keeps information about a client's session or earlier interactions and uses that information in later requests. Stateless designs are often easier to scale, while stateful designs can make ongoing sessions more convenient.

### B. Which category does HTTP belong to, and how are its limitations handled?

HTTP's request/response semantics are fundamentally **stateless**: each request can be understood independently. Websites commonly add state using **cookies and session IDs**, with session data stored on the server, or using **tokens** such as signed access tokens. The client sends the relevant cookie or token with later requests, allowing the application to associate requests with a user or session. This application-level state does not change HTTP's underlying stateless model.

## Question 3 — Client/Server Communication and Server-Initiated Updates

### A. What is the challenge?

In the usual HTTP request/response model, the client initiates a request and the server sends a response. A server cannot ordinarily send an unsolicited HTTP response over a connection when there is no outstanding request for it. Therefore, real-time alerts, chat messages, and webhook-style notifications need an additional mechanism or an established communication channel to deliver updates promptly.

### B. Possible solutions

**1. Webhooks.** A webhook is useful when one service needs to notify another service about an event. The receiving service exposes a callback URL, and the sending service makes an HTTP request—often a `POST`—to that URL when the event occurs. This is a good fit for service-to-service notifications because the receiver does not need to repeatedly poll for changes. The receiver must expose a reachable endpoint and should validate the sender, handle retries, and make processing idempotent where possible. A webhook is an event notification sent as a new HTTP request; it is not the server spontaneously sending a response to an earlier client request.

**2. WebSocket.** WebSocket starts with an HTTP-based handshake and then upgrades to a persistent, two-way channel. After the connection is established, either endpoint can send messages without waiting for a new HTTP request. This makes it appropriate for interactive chat and applications that need frequent, low-latency messages in both directions. The application must manage connection lifecycle, authentication, reconnection, and scaling across servers.

**3. Server-Sent Events (SSE).** SSE lets a browser open a long-lived HTTP connection and receive a stream of events from the server. It is a good option for one-way updates such as notifications, live status, or dashboards. It is simpler than WebSocket when the client mainly receives updates; client-to-server actions can still use ordinary HTTP requests. The server and infrastructure must support long-lived responses, and reconnect behavior should be considered.

**4. Long polling.** With long polling, the client sends a request and the server holds it open until an update is available or a timeout occurs. The client then processes the response and immediately sends another request. It can work when WebSocket or SSE is unavailable, but it creates repeated requests and may use more resources and add latency compared with a persistent event stream.

**Choice:** Use webhooks for event notifications between services; WebSocket for interactive, two-way chat; SSE for one-way live updates to a browser; and long polling as a compatibility fallback. The best choice depends on whether communication is service-to-service or browser-based, whether it must be bidirectional, and what the deployment environment supports.

## Question 4 — Identify the Parts of the URL

Given URL:

```text
ss://asghar:1234!!@ss.myproxy.com:1234#shadowSocks1
```

| Component | Value | Meaning |
|---|---|---|
| Scheme | `ss` | Identifies the Shadowsocks URI scheme. |
| User information | `asghar:1234!!` | The user-info portion before `@`; it contains `asghar` as the user-info name and `1234!!` after the colon. In a Shadowsocks URI, this is scheme-specific encoded configuration, so it should not automatically be interpreted as a conventional website username and password. |
| Host | `ss.myproxy.com` | The proxy server's hostname. |
| Port | `1234` | The port specified for the server. |
| Fragment | `shadowSocks1` | A client-side label or fragment; it is not normally sent to the server as part of the request target. |

**Note:** This is a scheme-specific URI, not a normal `http://` or `https://` URL. The exact interpretation of its user-info depends on the Shadowsocks URI format used by the relevant client.

## Question 5 — Appropriate HTTP Status Codes

More than one code can be reasonable depending on the API's exact behavior. The choices below state the assumptions used.

| # | Scenario | Recommended status | Reason |
|---|---|---|---|
| 1 | The website's database has failed and login requests cannot be processed. | **`500 Internal Server Error`** or **`503 Service Unavailable`** | Use `500` for an unexpected server-side failure. Use `503` when the service is temporarily unavailable, for example because a dependency is down or the service is overloaded. If possible, include `Retry-After` for a temporary outage. |
| 2 | The submitted username does not exist and login fails. | **`401 Unauthorized`** | The request lacks valid authentication credentials. In a login API, the response should avoid revealing whether a particular username exists; a generic invalid-credentials message is safer. |
| 3 | The website has moved to another domain. | **`301 Moved Permanently`** | Appropriate for a permanent move; include the new URL in the `Location` header. If the move is temporary, use `302 Found` instead. |
| 4 | The user has sent more requests than the permitted rate. | **`429 Too Many Requests`** | This status specifically represents rate limiting. A `Retry-After` header can tell the client when to try again. |
| 5 | A JWT refresh request succeeds and the new token is returned in a header. | **`200 OK`** | The refresh operation succeeded and the response carries the result. The refreshed token may be returned in a response header as stated in the scenario. |
| 6 | The user's IP range is blocked from accessing the service. | **`403 Forbidden`** | The server understands the request but refuses access under its policy. `403` is a suitable general choice for a denied IP range. |

## Question 6 — Proxy Types

### 1. Use university internal servers instead of central package servers (for example, Ubuntu `apt`)

**Caching proxy / package mirror.** The university can operate a local package mirror or caching proxy that stores packages retrieved from upstream repositories. Users download packages from the nearby university server, reducing repeated external traffic and often improving download speed. A complete mirror copies repository content locally; a caching proxy stores content as it is requested. For package managers, the university may configure an internal mirror as the repository source.

### 2. Route requests based on the geographic origin of the request

**Geographic load balancing (often implemented with DNS-based or edge/reverse-proxy routing).** A routing layer directs clients to an appropriate regional server or data centre based on the client's approximate location, latency, availability, or policy. This can reduce latency and distribute load across regions. The exact implementation may use GeoDNS, an edge network, or a reverse proxy/load balancer; therefore, “geographic load balancing” is the most precise description of the function rather than one single proxy protocol.

### 3. Add authentication to a very old service

**Reverse proxy / authentication gateway.** Place a reverse proxy in front of the legacy service and require users to authenticate at the proxy. After successful authentication, the proxy forwards permitted requests to the older backend, which can remain unchanged. This approach centralizes access control and can add modern authentication in front of a service that does not support it. The backend should be protected so clients cannot bypass the proxy, and the proxy must forward identity information safely if the backend needs it.

## References

- IETF, **RFC 9114: HTTP/3** — https://www.rfc-editor.org/rfc/rfc9114.html
- MDN, **HTTP response status codes** — https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status
- MDN, **HTTP overview** — https://developer.mozilla.org/en-US/docs/Web/HTTP/Overview
- MDN, **WebSocket API** — https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API
- MDN, **Server-sent events** — https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events
- IETF, **RFC 9110: HTTP Semantics** — https://www.rfc-editor.org/rfc/rfc9110.html

## Use of AI Assistance

AI assistance was used to help organize the answers and improve the wording of the explanations.
