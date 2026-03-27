# Research & Analysis: mask-443

## 1. Introduction
The **mask-443** project aims to provide stealthy, secure access to backend services (e.g., SSH) by multiplexing traffic on TCP port 443. Its core innovation is an **in-band, same-connection Single Packet Authorization (SPA)** followed by **WebSocket Secure (WSS) tunneling**, all while concurrently serving standard HTTPS/WSS traffic.

## 2. Competitive Landscape: Similar Projects

| Category | Project | Comparison with mask-443 |
| :--- | :--- | :--- |
| **Multiplexers** | `sslh`, `sshttp` | These are **passive**. They route traffic based on protocol signatures (e.g., SSH banner vs. TLS ClientHello) but **do not require authentication** before routing. A scanner can easily discover the hidden service. |
| **Stealth Proxies** | `Trojan`, `V2Ray`, `Shadowsocks` | These use **passive mimicry** to look like HTTPS. While very effective against DPI, they typically don't use a "pre-authentication" knock to hide the service from unauthorized probes before the TLS handshake. |
| **SPA Tools** | `fwknop` | The industry standard for SPA, but usually **out-of-band**. It requires a separate UDP/TCP packet to open a firewall rule, then a new connection. `mask-443` performs SPA and tunneling on the **same TCP connection**. |
| **Kernel Stealth** | `TCP Stealth` | Extremely stealthy by hiding the knock in TCP Sequence Numbers (ISN). However, it is **fragile** (broken by NAT/proxies) and requires kernel modifications. `mask-443` is application-layer and more portable. |

## 3. The "mask-443" Advantage: Key Differentiators
1.  **In-Band Authentication:** By performing SPA on the same connection as the subsequent TLS handshake, `mask-443` eliminates the latency and complexity of out-of-band "knocking."
2.  **Service Concealment:** Unlike `sslh`, the hidden service is cryptographically invisible. Probing port 443 without the correct SPA packet will only reveal the "Public" HTTPS service.
3.  **DPI Resilience:** Using WSS (WebSockets over TLS) ensures the traffic is indistinguishable from standard encrypted web traffic, allowing it to bypass strict corporate firewalls.
4.  **Zero Trust Architecture:** Authentication happens *before* the TLS stack is engaged, protecting the server from vulnerabilities in the TLS library itself.

## 4. Discussion: Validity, Usability, and Feasibility

### 4.1 Technical Validity
*   **Security:** The "Authenticate First" approach is a proven security best practice (Zero Trust). AES-GCM for the SPA packet provides strong encryption and authentication.
*   **Stealth:** The use of a buffered reader to "peek" at initial bytes is a technically sound way to implement multiplexing without protocol-level interference.

### 4.2 Usability
*   **Client Complexity:** The need for a custom client to generate the SPA packet is a slight hurdle. However, support for **OpenSSH ProxyCommand** makes the experience seamless for power users once configured.
*   **Environment Compatibility:** Port 443 is universally open. WSS is a standard protocol. The main limitation is non-transparent proxies that might expect immediate TLS, though most modern environments handle this gracefully.

### 4.3 Feasibility
*   **Implementation:** The current Go skeleton using `BufferedConn` is the correct architectural path. The logic for WSS proxying and SPA parsing is well-defined and achievable using standard libraries (e.g., `gorilla/websocket`).

## 5. Conclusion
**mask-443** is a highly valid and specialized tool for secure, stealthy remote access. It successfully bridges the gap between basic port multiplexers (which lack security) and complex VPNs (which lack stealth). Its architecture is robust, and its use cases—bypassing restrictive firewalls while maintaining a low profile—are highly relevant for modern network environments.
