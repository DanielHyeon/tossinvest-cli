## 1. Evidence and RED contracts

- [x] 1.1 Capture the implementation base, CodeGraph/CodeGraphContext reconciliation, Function Logic Maps, Branch Test Maps, risk reports, and High-risk Pre-Edit declaration.
- [x] 1.2 Add failing console tests for saved-ready, missing, rejected, environment-managed rejection, transient/refresh failure, secret-safe form, deterministic outer-body 413 plus preserved token-session or trusted-network access mode/origin/CSRF gates, serialized concurrent requests, truthful partial failures, and successful login-to-soak continuation.
- [x] 1.3 Add failing `cmd/tossctl` tests for credential source/load/probe classification, isolated replacement-token validation, fresh environment-generation validation plus old-cache invalidation, validate-before-save, fixed-detail audit ordering, persistent save, normal-cache invalidation, 0600 marker creation, non-destructive existing-marker reuse, cross-instance restart blocking, save/cache/audit/read/remove failure closure, valid-resubmission clearing, exact environment precedence, and no real process/API side effects.

## 2. Console onboarding workflow

- [x] 2.1 Add narrow Open API preflight/setup seams, gated GET/POST routes, blank secret-safe form rendering, and a route-specific body limit installed before mutation form parsing.
- [x] 2.2 Serialize restart/setup credential generations, make soak restart preflight saved credentials, redirect only missing/auth-rejected file states to setup, classify transient and environment-managed failures, and continue successful setup directly into soak restart.
- [x] 2.3 Update static route/capability guards so the write-only credential setup exception remains common-access-mode+CSRF gated and no credential-read or order route is introduced.

## 3. CLI wiring and persistence

- [x] 3.1 Wire console preflight to the existing official credential loader and read-only validation probe with no dashboard polling.
- [x] 3.2 Wire setup to validate with an isolated temporary 0600 token cache, persist a secret-free 0600 incomplete-generation marker before save, atomically replace and verify a regular 0600 credential file, invalidate the normal cache, emit a fixed secret-free audit event, then clear the marker; after old soak exit invalidate the cache again immediately before spawn, preserve access-token renewal and Docker config-mount behavior, and refuse ineffective file replacement for environment-managed credentials.

## 4. Verification and delivery

- [ ] 4.1 Run focused tests, console race tests, full tests, vet, validate, secret scans, direct-TLS plus token-session and trusted-network peer/Host/origin/CSRF browser-contract QA, deployed missing-credential fail-closed QA, persistent marker path/mode/recovery tests, and rollback-procedure review without starting a real soak, engine, toggle, or LIVE order.
- [ ] 4.2 Complete proposal-freeze and implementation security/QA reviews, refresh SDD indexes, and pass `make sdd-check` plus `make gate CHANGE=automate-soak-openapi-onboarding`.
- [ ] 4.3 Commit and push the feature branch, build/recreate the Docker service, and verify the deployed one-click browser flow with an isolated empty config or otherwise blocked side effects.
