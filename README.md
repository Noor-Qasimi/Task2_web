HTTP Assignment — Internet Engineering
Final answer draft for README.md
Question 1: HTTP Versions and HTTP/3 (3 points)
Key differences between HTTP versions
• HTTP/0.9: The earliest version of HTTP. It supported only the GET method and returned the requested document without HTTP headers or status codes. It was designed for simple hypertext transfer.
• HTTP/1.0: Introduced HTTP headers, status codes, and additional methods such as POST. It made HTTP more flexible by supporting metadata and different types of content. Connections were generally closed after each request unless extensions were used.
• HTTP/1.1: Introduced persistent connections by default, allowing multiple requests to reuse the same TCP connection. It also improved caching, content negotiation, virtual hosting, and transfer of dynamically generated content. However, multiple responses on one connection could still suffer from head-of-line blocking.
• HTTP/2: Introduced binary framing, multiplexing, and header compression through HPACK. Multiple request and response streams can share a single connection, reducing connection overhead and improving performance. However, because HTTP/2 uses TCP, packet loss can delay the delivery of data across all streams on that connection.
• HTTP/3: Uses QUIC instead of TCP as its transport protocol. QUIC runs over UDP and provides encryption, connection establishment, and independent streams. This reduces connection-establishment delays and prevents packet loss in one stream from directly blocking the delivery of data in other streams.
Why does HTTP/3 use UDP?
HTTP/3 uses UDP because UDP provides a lightweight transport interface over which QUIC can implement its own reliable delivery, congestion control, encryption, and stream management. Unlike TCP, QUIC supports independent streams, so packet loss affecting one stream does not necessarily block other streams. This can improve performance on unreliable networks and reduce connection-establishment latency. HTTP/3 does not gain reliability from UDP itself; reliability is provided by QUIC.
Question 2: Stateless and Stateful Communication (3 points)
A. Explain the difference
In stateless communication, the server does not need to retain the application state of previous requests to understand a new request. Each request contains the information required for the server to process it. In stateful communication, the system maintains information about an ongoing interaction, such as a user's session or a connection's state. Stateless designs are generally easier to scale because requests can be handled independently, while stateful designs can simplify continuous interactions but require state management.
B. Which category does HTTP belong to?
HTTP is fundamentally a stateless application-layer protocol. Each request is processed independently at the protocol level, and the server is not required to remember previous requests. However, many web applications need to preserve user-specific information across requests. They commonly achieve this by using cookies, session identifiers, or tokens such as JSON Web Tokens (JWTs). These mechanisms allow an application to maintain logical state without changing HTTP's underlying stateless nature.
Question 3: The Client/Server Communication Limitation (4 points)
A. Explain the challenge
In the traditional HTTP request-response model, the client initiates a request and the server returns a response. The server cannot normally initiate an independent HTTP response to a client when no request or suitable communication channel exists. This creates a challenge for applications that need to deliver real-time events, chat messages, alerts, or notifications. Without an additional mechanism, the client must repeatedly request updates or establish a persistent communication channel.
B. Possible solutions
1. Polling and Long Polling
With regular polling, the client sends requests at fixed intervals to check for new information. This approach is simple and works with ordinary HTTP infrastructure, but it can generate unnecessary traffic and introduce delays between updates. Long polling improves this approach: the server keeps a request open until new information becomes available or a timeout occurs, then responds and allows the client to send another request. Long polling reduces unnecessary empty responses but still requires repeated requests and server-side connection management.
2. Server-Sent Events (SSE)
Server-Sent Events allow a server to send a continuous stream of events to a client over a long-lived HTTP response. After the client establishes the connection, the server can send new events whenever they become available. SSE is particularly suitable for notifications, dashboards, and live updates where communication is primarily from server to client. It also supports automatic reconnection in compatible clients. However, SSE is not designed for general-purpose, bidirectional communication.
3. WebSocket
WebSocket provides a persistent, full-duplex communication channel between a client and a server. After an initial HTTP-based handshake, both sides can send messages independently without creating a new HTTP request for every message. This makes WebSocket suitable for interactive chat applications, multiplayer games, and other real-time systems requiring communication in both directions. Compared with ordinary request-response communication, it requires additional connection management and infrastructure support.
4. Webhooks
Webhooks solve a related but different problem. Instead of repeatedly asking a service whether an event has occurred, a client or application registers a callback URL. When the event occurs, the service sends an HTTP request, usually a POST, to that URL. Webhooks are commonly used for payment notifications, repository events, and communication between backend services. They are event-driven and efficient, but the receiving endpoint must be reachable by the sender and should verify the authenticity of incoming requests. Unlike WebSocket, a webhook is generally a server-to-server notification rather than a persistent interactive channel.
In summary, polling and long polling are useful when simple HTTP compatibility is important, SSE is suitable for one-way server-to-client updates, WebSocket is appropriate for continuous two-way communication, and webhooks are useful for event notifications between independent services.
Question 4: Identifying URL Components (3 points)
Given URL:
ss://asghar:1234!!@ss.myproxy.com:1234#shadowSocks1
The components are:
The @ character separates the user-information component from the host. The colon between the hostname and the final 1234 separates the host from the port. The fragment follows # and is not normally sent as part of the HTTP request target.
Note: This is a URI using the ss scheme, not an ordinary HTTP or HTTPS URL. The exact meaning of its user-information and fragment components depends on the application implementing that scheme.
Question 5: Choosing HTTP Status Codes (4 points)
1. The website's database has a problem, and login requests cannot be processed.
Recommended status: 503 Service Unavailable
This status is appropriate when the database failure makes the service temporarily unavailable. If the failure is unexpected and does not clearly represent temporary unavailability, 500 Internal Server Error may also be appropriate. The choice depends on how the application classifies and handles the failure.
2. The username entered by the user does not exist, and login fails.
Recommended status: 401 Unauthorized
The request has not been authenticated successfully. A 401 response is appropriate for invalid or missing authentication credentials. The application should avoid revealing unnecessary information about whether a particular username exists.
3. The website has moved to another domain.
Recommended status: 301 Moved Permanently
This status indicates that the resource has permanently moved to a new URL, which should be provided in the Location response header. If preserving the original HTTP method is essential, 308 Permanent Redirect is an alternative.
4. The user sends more requests than the permitted rate.
Recommended status: 429 Too Many Requests
This status indicates that the client has exceeded the server's rate limit. The server may include a Retry-After header to indicate when the client can retry.
5. A JWT refresh request succeeds, and the renewed token is returned in a response header.
Recommended status:  200 OK
The refresh operation completed successfully, and the renewed token can be returned in a response header. If the API intentionally returns no response body, 204 No Content may also be appropriate, provided the API's response design is consistent.
6. The user's IP address belongs to a range that is blocked from accessing the service.
Recommended status:  403 Forbidden
The server understands the request but refuses access because of the applicable access policy. A 403 response is appropriate when the IP-based restriction is the reason for denying access.
Question 6: Identifying Proxy Types (3 points)
1. Using internal university servers instead of central package servers to download packages through a package manager such as APT
Recommended type: Caching proxy or package mirror
An internal caching proxy stores copies of packages requested by users and reuses them for later requests. A package mirror instead maintains a synchronized copy of a package repository. Both approaches can reduce external bandwidth usage and improve download speed. If the university provides a complete, synchronized repository, a package mirror is the more precise term.
2. Distributing requests according to the geographical origin of the requests
Recommended type: Reverse proxy with geographic routing
A reverse proxy receives client requests and forwards them to appropriate backend servers. With geographic routing, the proxy or an associated routing layer selects a backend based on the client's approximate geographic location. This can reduce latency and distribute traffic across regional infrastructure.
3. Adding authentication to a very old service
Recommended type: Reverse proxy acting as an authentication gateway
A reverse proxy can be placed in front of a legacy service that does not support authentication. The proxy validates incoming credentials or tokens before forwarding authorized requests to the backend. This adds an authentication layer without requiring changes to the legacy application itself. The backend should also be isolated from direct external access so that clients cannot bypass the proxy.
References
1. MDN Web Docs, *Evolution of HTTP*: https://developer.mozilla.org/en-US/docs/Web/HTTP/Guides/Evolution_of_HTTP
2. RFC 9110, *HTTP Semantics*: https://www.rfc-editor.org/rfc/rfc9110
3. MDN Web Docs, *HTTP Response Status Codes*: https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status
4. MDN Web Docs, *Server-Sent Events*: https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events
AI Assistance Disclosure
AI assistance was used to help organize the answers and improve the wording of the explanations. 
