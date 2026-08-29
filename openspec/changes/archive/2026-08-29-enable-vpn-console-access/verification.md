## Verification evidence

Date: 2026-07-31

### Behavior and regression checks

- RED compilation failures were observed before the remote access types, token
  loader and CLI flags existed; focused tests are now GREEN.
- Remote tests cover all-or-nothing configuration, exact listener bind, CIDR and
  ignored forwarding headers, certificate hostname, Host/Origin/Referer/CSRF,
  distinct Secure sessions, IP/UA binding, idle/absolute expiry, logout,
  restart handoff, bounded rate limiting, audit failure and minimal health.
- Race-enabled `internal/console` and `cmd/tossctl` tests passed.
- Full `make test`, `make vet`, `make validate`, and strict validation of
  `enable-vpn-console-access` passed.
- Function Logic Map checking passed with current AST/risk evidence and complete
  RED/GREEN branch maps.
- Dockerfile analysis scored 100/100. Compose fails when its VPN host bind is
  absent and renders one exact `host_ip` when complete.
- A built dummy-secret container became healthy, returned `ok` over TLS, logged
  in with HTTP 303, opened the authenticated console with HTTP 200, wrote a 0600
  access audit, and was then stopped and removed.
- Production Go/deployment/docs files have zero high/critical secret-scanner
  findings. Repository-wide heuristic results were false positives in hashes,
  lockfiles, identifier text, and explicit fake test credentials.
- Independent implementation review resolved address-family ambiguity, host
  secret ownership, unbounded rate-limit audit/session state, logout audit
  failure, and initial remote-auth complexity.
- `make sdd-sync` and `make sdd-check` passed. GBrain freshness emitted its
  documented advisory busy warning; CodeGraph hard evidence matched the worktree.

### Container evidence

- Built image declares `USER 10001:10001`, port 37085 and the TLS healthcheck.
- Smoke runtime used non-root UID/GID, read-only root, `cap_drop: ALL`,
  no-new-privileges, PID 128, memory 512 MiB and one CPU.
- Runtime layers contain the binary, entrypoint and CA bundle only; source, VCS
  metadata and secret files are excluded.
- Compose secret sources are read-only mounts and copied with mode 0600 into a
  private tmpfs; the entrypoint refuses UID 0.

### Safety evidence

- No LIVE gate or trading toggle was changed.
- The engine was not started.
- No real broker session was loaded for the container smoke.
- No order, cancellation, amendment, adoption, verification approval, or other
  account mutation was submitted.
- Remote authentication grants only the existing console route graph. It adds no
  direct order route, no automatic verify nonce, and no bypass of the engine
  startup interlock.

## Trusted-network revision evidence

Date: 2026-07-31

- RED compilation failed before `TrustedNetwork` and `trustedNetwork` existed;
  focused tests subsequently passed.
- Direct trusted-network `GET /` returns 200 without redirect or application
  session cookie. The compatibility login page is not exposed in this mode.
- Outside-CIDR requests, wrong-origin POSTs, and missing-CSRF POSTs return 403.
- `go test ./internal/console ./cmd/tossctl` and the race-enabled equivalent
  passed after the final banner and no-login assertions.
- Full `make test`, `make vet`, `make validate`, strict change validation, PM
  generated-tracker check, and `docker compose config` passed.
- Rendered Compose contains `--trusted-network` and contains neither
  `--remote-token-file` nor a remote-token secret.
- Repository-wide heuristic scanning reported historical/test/hash false
  positives, while the revised production Go, Compose, Docker, deploy, and
  operations files contained zero high/critical findings.
- A Go-specific `code-reviewer` quality run over `internal/console` reported
  zero issues; manual review verified the retained middleware/gate ordering.
- Affected trusted-network AST and risk evidence were refreshed. The repository
  logic-map checker still reports only maps owned by the concurrently modified
  optimization and engine-autostart work; no trusted-network target remains
  stale or missing.
- `make sdd-sync` completed with the documented GBrain-busy advisory and current
  CodeGraph evidence. `make sdd-check` passed.
- Before deployment, persisted `engine.autostart` and `engine.gate_enabled` were
  both false and no engine marker existed. No engine or account mutation was
  invoked during verification.
- The final `tossos:local` image built successfully as non-root UID/GID
  `10001:10001`; its binary advertises `--trusted-network`, and the running
  container was deliberately not recreated.
- The final change gate reached Function Logic Map step 4/8 and stopped only on
  concurrently modified optimization/engine-autostart functions whose three
  active changes share the same frozen base commit.
- After the operator explicitly repeated the no-auth deployment request, Compose
  recreated the service with image
  `sha256:79ad6013347674e065317a768c8fb018073bd7ac42b032111b8f2252d31bc639`.
  The service became healthy with `restart: unless-stopped`, its command contains
  `--trusted-network`, and the obsolete remote-token mount is absent.
- Post-deployment HTTPS smoke returned 200 from `/` with no redirect, no session
  cookie, no login text, and the expected HSTS/CSP headers. `/login` returned
  404, `/healthz` returned `ok`, and the engine marker remained absent.
