## 1. Contract and hard evidence

- [x] 1.1 Capture the implementation base, validate the OpenSpec change strictly, and record PM bootstrap tracking.
- [x] 1.2 Complete proposal-freeze Security/Eng adversarial review with the tailored STRIDE/DREAD model and resolve every DREAD ≥7 finding in design/spec ownership.
- [x] 1.3 Build pre-edit Function Logic Maps for `Listen`, `Serve`, `ListenAndServe`, `loopbackOnly`, `Console.routes`, `Console.session0`, `Console.mutating`, `Console.grantSession`, `runConsole`, `newConsoleCmd`, and restart handoff/relaunch functions.

## 2. RED tests for remote trust boundaries

- [x] 2.1 Add RED construction/listener tests for incomplete remote config, exact bind verification, invalid/global CIDR, TLS certificate hostname, and unchanged loopback defaults.
- [x] 2.2 Add RED authentication tests for non-URL credentials, token mismatch, secure distinct session cookie, IP/UA binding, idle/absolute expiry, logout, handoff, and audit failure.
- [x] 2.3 Add RED request-boundary tests for actual peer CIDR, ignored forwarding headers, exact Host, same Origin/Referer plus CSRF, rate limit, security headers, and fixed health endpoint.
- [x] 2.4 Add RED CLI characterization tests for all-or-nothing flags, 0600 token-file loading, no secret in argv/banner, and relaunch flag preservation.

## 3. Remote console implementation

- [x] 3.1 Implement validated immutable remote listener configuration and TLS 1.3 serving while preserving the local `Listen(port)`/loopback refusal path.
- [x] 3.2 Implement bounded login attempts, append-only access audit, server-side IP/UA-bound expiring sessions, secure cookies, logout, and remote handoff session issuance.
- [x] 3.3 Implement peer-CIDR, Host, Origin/Referer, CSRF ordering, security-header middleware, HTTP timeouts, and minimal `/healthz`.
- [x] 3.4 Wire `tossctl console` remote flags and secret/certificate validation without adding a non-interactive verify/order approval path.

## 4. Container deployment

- [x] 4.1 Add a pinned-version multi-stage non-root Dockerfile and `.dockerignore` with no source/build tools/secrets in runtime layers.
- [x] 4.2 Add `compose.yaml` and `.env.example` with required VPN host-IP publish, required CIDR/public URL/secret paths, persistent config/data, read-only secrets, and hardened bounded runtime settings.
- [x] 4.3 Document native VPN and Compose setup, token/certificate rotation, container image update/rollback, firewall checks, and explicit prohibition of public/wildcard host publish.
- [x] 4.4 Validate Compose missing-variable failure, rendered config, image build/inspect, and dummy-secret TLS health smoke without loading a real broker session or making an account request.

## 5. Verification and review

- [x] 5.1 Refresh every affected Function Logic Map AST/risk report and map all RED/GREEN branches.
- [x] 5.2 Run focused and race tests, full `make test`, `make vet`, strict OpenSpec validation, `make validate`, Docker/Compose checks, and high/critical secret scan.
- [x] 5.3 Run independent adversarial implementation review, resolve findings, then run `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=enable-vpn-console-access`.
- [x] 5.4 Record verification evidence and episodic memory without changing LIVE gate, starting the engine, or submitting an account mutation.

## 6. User-approved trusted-network revision

- [x] 6.1 Record the user's explicit decision that loopback and authenticated VPN membership are the application access boundary, and revise proposal/design/spec/threat ownership without weakening TLS, CIDR, Host/Origin, CSRF, audit, or LIVE approval invariants.
- [x] 6.2 Add RED tests for explicit trusted-network configuration, unauthenticated dashboard access, absent login/token/session requirements, preserved CIDR/Host/Origin/CSRF rejection, and unchanged engine/verify interlocks.
- [x] 6.3 Implement the minimal trusted-network bypass while retaining the authenticated mode as an opt-in compatibility path and keeping partial or implicit remote configuration fail-closed.
- [x] 6.4 Switch Compose and operations documentation to explicit trusted-network mode and remove the deployed remote-token secret requirement.
- [x] 6.5 Refresh affected Function Logic Maps, run focused/race/full/Compose/security checks, independently review the revision, and record pre-deployment evidence without starting the engine or placing an order.
