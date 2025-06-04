Project Requirements Document: SPA-Protected Multiplexed Tunneling System (SPAMTS)
1. Introduction
 * 1.1. Project Goal: To develop a secure system (SPAMTS) that provides access to a backend service (e.g., SSH) via a tunnel. This tunnel will be obfuscated as standard web traffic (e.g., Secure WebSocket - WSS) on port 443. Access to initiate this tunnel will be gated by a Single Packet Authorization (SPA) mechanism, also on port 443. The system must also allow standard HTTPS and/or public WSS traffic on port 443 to be served or proxied concurrently without requiring SPA.
 * 1.2. System Scope: The system comprises two main components:
   * SPAMTS Server: A custom server-side application.
   * SPAMTS Client: A client-side application or wrapper responsible for performing SPA and establishing the tunnel.
 * 1.3. Core Paradigm: "Authenticate First (via SPA on port 443), then Obfuscate/Multiplex (tunneling and public services on port 443)."
 * 1.4. Target Use Case: Secure, stealthy remote access to a service, bypassing restrictive firewalls that only allow standard web traffic on port 443, while also providing standard web services on the same port.
2. General System Requirements
 * 2.1. Port Unification: All external inbound network communication for SPA, tunneling, and public web services shall occur exclusively over TCP port 443.
 * 2.2. Operating Environment:
   * Server: The SPAMTS Server shall be deployable on common Linux distributions supporting both x86_64 and ARM-based (e.g., arm64/aarch64) architectures.
   * Client: The SPAMTS Client shall be compatible with common client operating systems (Linux, macOS, Windows) supporting both x86_64 and ARM-based (e.g., arm64/aarch64, Apple Silicon) architectures.
 * 2.3. Security Principles: The system shall adhere to principles of defense-in-depth, least privilege, and aim for maximum stealth of the tunneling capability. All cryptographic operations shall use strong, industry-standard algorithms and practices.
 * 2.4. Protocol Obfuscation: The tunneling protocol shall be Secure WebSockets (wss://) to emulate legitimate web traffic.
3. Server-Side Application (SPAMTS Server) Requirements
 * 3.1. Port Listening & Initial Connection Handling
   * 3.1.1. The server shall listen for incoming TCP connections exclusively on port 443 on a designated network interface.
   * 3.1.2. For each new incoming connection, the server shall buffer initial data packet(s) to determine the client's intent (SPA knock or standard service request) before committing to a specific protocol handler.
   * 3.1.3. The server shall implement a short timeout for receiving initial data to prevent resource exhaustion from idle connections.
 * 3.2. Single Packet Authorization (SPA) Module
   * 3.2.1. Packet Interpretation:
     * The server shall inspect the initial data from a new connection to detect a potential SPA packet signature.
     * The SPA packet format shall be clearly defined (e.g., fixed/variable length fields, specific markers).
   * 3.2.2. Cryptography:
     * The SPA packet payload shall be encrypted (e.g., using AES-GCM).
     * The SPA packet shall be authenticated using a strong Message Authentication Code (e.g., HMAC-SHA256 or the tag from AES-GCM).
     * The server shall use pre-shared keys (PSK) for SPA encryption and authentication. Secure methods for key generation, storage, and distribution must be considered (though key distribution is outside the scope of the application runtime itself, secure loading is).
   * 3.2.3. Anti-Replay Mechanisms:
     * The SPA packet shall include a timestamp. The server shall validate this timestamp against its current time within a configurable tolerance window.
     * The SPA packet shall include a nonce. The server shall maintain a store of recently accepted nonces for a configurable period to detect and reject replays.
   * 3.2.4. Authentication & Authorization Logic:
     * Upon successful decryption, MAC validation, timestamp validation, and nonce validation, the server shall consider the SPA knock authentic.
     * The server shall authorize the client (e.g., based on their source IP address from the TCP connection) to proceed with establishing a tunnel.
     * The server shall log all SPA attempts (successful and failed) with relevant details (timestamp, source IP, reason for failure).
   * 3.2.5. Transition to Tunneling:
     * Upon successful SPA from a client, the server shall seamlessly transition the same TCP connection to expect a TLS Client Hello for the WSS tunnel. The server must be prepared to act as a TLS server for this connection.
     * Alternatively (if same-connection transition is too complex for PoC), the SPA success could authorize the client's IP to make a new connection to port 443 within a short time window, which will then be identified (e.g., via a specific secret path or pre-arranged signal) as a request for the WSS tunnel. [Decision Point: Same connection preferred for elegance, new connection might be easier initially.]
 * 3.3. Public Service Handling Module (Non-SPA Traffic)
   * 3.3.1. If initial data inspection does not indicate an SPA knock (e.g., a TLS Client Hello is received immediately), the server shall treat the connection as a standard request for a public service.
   * 3.3.2. TLS Termination: The server shall be capable of acting as a TLS server, performing TLS handshakes with clients using a configurable server certificate and private key.
   * 3.3.3. HTTPS Request Handling:
     * The server shall be able to identify HTTPS requests (after TLS).
     * It shall either serve basic static HTTP/S content (e.g., a decoy page) or proxy valid HTTPS requests to a backend web server.
     * Support for SNI (Server Name Indication) should be included to allow hosting or proxying for multiple hostnames.
   * 3.3.4. Public WSS Request Handling:
     * The server shall be able to identify WebSocket Secure (wss://) upgrade requests (after TLS) destined for public WebSocket services.
     * It shall either handle these public WSS connections directly (if it's a simple WebSocket service) or proxy them to a backend public WebSocket server.
 * 3.4. Tunneling Service Module (Post-SPA Authorization & WSS Based)
   * 3.4.1. WSS Endpoint Activation: This module handles connections pre-authorized by the SPA module for tunneling.
   * 3.4.2. TLS Handling for Tunnel:
     * The server shall complete a TLS handshake for the WSS tunnel (this might be the same TLS session if transitioning from SPA on the same connection, or a new one if a new connection is made post-SPA).
   * 3.4.3. WebSocket Handshake:
     * The server shall perform a WebSocket handshake (HTTP Upgrade) with the authorized client over the established TLS connection.
     * This handshake may occur on a specific, non-obvious path, potentially communicated or known via the SPA process.
   * 3.4.4. Data Framing & Proxying:
     * Once the WSS tunnel is established, the server shall receive binary WebSocket messages from the client.
     * The server shall extract the payload from these messages and forward it over a new TCP connection to the configured internal backend service (e.g., localhost:22 for SSHD).
     * The server shall read data from the backend service connection, wrap it in binary WebSocket messages, and send it to the client over the WSS tunnel.
     * It shall manage the lifecycle of the connection to the backend service, tying it to the WSS client connection.
   * 3.4.5. The WSS endpoint used for tunneling shall not be accessible or discoverable without prior successful SPA.
 * 3.5. Logging & Monitoring
   * 3.5.1. The server shall generate detailed logs for all significant events, including SPA attempts (success/failure), tunnel establishments, public service accesses, errors, and connection lifecycles.
   * 3.5.2. Log format shall be configurable (e.g., plain text, JSON).
   * 3.5.3. Log verbosity shall be configurable.
 * 3.6. Configuration Management
   * 3.6.1. The server shall be configurable via a configuration file.
   * 3.6.2. Configurable parameters shall include: listening interface/port, SPA secrets (key paths), SPA timing/nonce parameters, TLS certificate/key paths, public HTTPS/WSS backend details (if proxying), and internal backend service details for tunneling (e.g., target_host:target_port).
 * 3.7. Security Hardening
   * 3.7.1. The server shall be developed with secure coding practices to prevent common vulnerabilities.
   * 3.7.2. It shall handle network errors and unexpected client behavior gracefully without crashing.
4. Client-Side Application (SPAMTS Client) Requirements
 * 4.1. SPA "Knock" Generation
   * 4.1.1. The client shall construct the SPA packet payload including a current timestamp, a unique nonce, and any other required data as per the defined SPA packet format.
   * 4.1.2. The client shall encrypt the SPA payload and compute an HMAC using pre-shared keys.
   * 4.1.3. The client shall send the composed SPA packet as the initial data over a new TCP connection to the server's port 443.
   * 4.1.4. The client shall keep the TCP connection open after sending the SPA packet to proceed with the WSS tunnel establishment on the same connection. (This assumes the "same connection transition" for Option C).
 * 4.2. Tunnel Establishment (WSS Based, Post-SPA on Same Connection)
   * 4.2.1. Immediately after sending the SPA packet, the client shall initiate a TLS handshake over the same TCP connection, acting as a TLS client.
   * 4.2.2. The client shall perform server certificate validation (configurable: e.g., trust system CAs, allow pinning a specific server certificate/CA).
   * 4.2.3. After successful TLS handshake, the client shall initiate a WebSocket handshake (HTTP Upgrade request) over the TLS connection to the server-defined path (if applicable).
   * 4.2.4. The client shall handle WebSocket handshake responses from the server.
 * 4.3. Local Application Interfacing
   * 4.3.1. The client shall be usable as an OpenSSH ProxyCommand. It will receive unencrypted data from the local SSH client via stdin and send data back to the SSH client via stdout.
   * 4.3.2. (Optional) The client may offer an alternative mode where it listens on a local TCP port, accepting connections from local applications and tunneling them.
 * 4.4. Data Tunneling
   * 4.4.1. Once the WSS tunnel is established, the client shall read data from the local application (e.g., stdin if ProxyCommand).
   * 4.4.2. It shall wrap this data into binary WebSocket messages and send them to the server.
   * 4.4.3. It shall receive binary WebSocket messages from the server, extract the payload, and write it to the local application (e.g., stdout if ProxyCommand).
 * 4.5. Configuration
   * 4.5.1. The client shall be configurable via command-line arguments and/or a configuration file.
   * 4.5.2. Configurable parameters shall include: server address/port, SPA secrets (key paths), local listen port (if applicable), server certificate validation options.
 * 4.6. User Interface/Feedback
   * 4.6.1. The client shall provide informative output regarding connection status, SPA attempt, tunnel establishment, and errors.
   * 4.6.2. A verbose/debug mode shall be available for troubleshooting.
5. Non-Functional Requirements
 * 5.1. Security: No cleartext secrets in logs (unless in debug mode explicitly warning about it). Protection against common network attacks.
 * 5.2. Performance: Low latency for the established tunnel, efficient handling of concurrent connections (server-side).
 * 5.3. Reliability: Stable operation, graceful error handling.
 * 5.4. Maintainability: Well-structured, commented code.
 * 5.5. Usability: Clear configuration, informative feedback for users.
