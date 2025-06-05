# mask-443
Multiplexed Access with Secure Knock-Authentication on 443

## Overview
**mask-443** (Multiplexed Access with Secure Knock-Authentication on 443) is a sophisticated system designed to provide stealthy, secure access to backend services (such as SSH) while simultaneously serving standard web traffic, all over a single TCP port: 443. This allows services to be accessed through restrictive network environments that typically only permit HTTPS/WSS traffic, effectively "masking" the presence of the hidden services.
The core principle of mask-443 is an "authenticate first, then obfuscate and multiplex" paradigm. It achieves this through two main components: a custom server application and a corresponding client utility.

Key Features:
 * Secure Knock-Authentication (SPA-like): Before any service tunnel is established, clients must perform a cryptographic "knock" – a Single Packet Authorization sequence. This pre-authentication step ensures that the tunneling endpoint remains entirely hidden and inaccessible to unauthorized users and automated scanners.
 * Obfuscated Tunneling via Secure WebSockets (WSS): Once authenticated by the "knock," mask-443 establishes a Secure WebSocket (wss://) tunnel between the client and server. Traffic for backend services (e.g., SSH) is then routed through this tunnel, appearing as legitimate encrypted web traffic on port 443.
 * Multiplexed Public Services: Concurrently, the mask-443 server can handle standard HTTPS requests and/or public (non-hidden) Secure WebSocket connections on the same port 443. This allows a single public-facing endpoint to serve both regular web applications and provide a hidden gateway for authorized users.
 * Port Unification: All interactions – the secure knock, the obfuscated tunnel, and public web traffic – occur exclusively over TCP port 443, maximizing compatibility with strict firewall policies.
 * Enhanced Security & Stealth: By hiding service entry points until a cryptographic knock is received and then channeling traffic through an encrypted WSS tunnel, mask-443 provides a robust defense-in-depth strategy, significantly reducing the attack surface and enhancing the stealth of your backend services.

The mask-443 server is the intelligent gatekeeper, listening on port 443, differentiating SPA knocks from regular traffic, and managing both public services and the SPA-triggered WSS tunnels. The mask-443 client is responsible for generating the secure knock and establishing the WSS tunnel to interface with local applications (e.g., an SSH client via ProxyCommand).

This system is ideal for users and administrators who need secure, covert access to services in environments where direct connections are blocked or heavily monitored, without sacrificing the ability to host standard web services on the conventional HTTPS port.
