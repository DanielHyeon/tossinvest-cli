## 1. Contract and evidence

- [x] 1.1 Validate Story↔OpenSpec 1:1 registration, strict OpenSpec artifacts, and capture the implementation base commit.
- [x] 1.2 Record CodeGraph hard evidence, CodeGraphContext supporting context, and evidence reconciliation for `remoteRuntime.sameOrigin` and `Console.mutating`.
- [x] 1.3 Create and complete the Go AST, Function Logic Map, Branch Test Map, risk report, and Pre-Edit Gate for `remoteRuntime.sameOrigin` and `Console.mutating`.
- [x] 1.4 Record the proposal-freeze security/engineering review and rollback boundary.

## 2. RED regression coverage

- [x] 2.1 Add a failing test proving headerless canonical TLS+Host POSTs reach the handler only with valid CSRF.
- [x] 2.2 Add failing/retained tests for explicit/empty/contradictory Origin/Referer, strict headerless login rejection, non-TLS mutation requests, wrong host/port, forwarding-header distrust, and path-independent Referer comparison.

## 3. Minimal implementation

- [x] 3.1 Keep strict Origin→Referer precedence in `remoteRuntime.sameOrigin`, add a mutation-only direct TLS+Host fallback, and wire only `Console.mutating` to it.
- [x] 3.2 Confirm explicit mismatches never fall through and forwarding headers remain ignored.

## 4. Verification and review

- [x] 4.1 Run focused console tests, full tests, race/vet/validate, OpenSpec strict validation, `make sdd-sync`, and `make sdd-check`.
- [x] 4.2 Run an independent diff/security review and resolve all blocking findings.
- [x] 4.3 Complete verification evidence and prepare all inputs for `make gate CHANGE=fix-console-origin-fallback`.

## 5. Delivery and closure readiness

- [x] 5.1 Freeze the reviewed commit scope and confirm delivery will not touch LIVE orders or operating toggles.
- [x] 5.2 Record the Compose rebuild/recreate, health/restart-policy, and safe deployed POST verification procedure.
- [x] 5.3 Prepare the post-gate OpenSpec archive, Story-path/PM sync, episodic retain, and final SDD check procedure.
