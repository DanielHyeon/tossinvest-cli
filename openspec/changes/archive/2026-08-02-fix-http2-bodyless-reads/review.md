# Review

- Date: 2026-08-01
- Trigger: deployment canary failure on TLS HTTP/2
- Voices: main release/debug review, specialized code-debugger root-cause review
- Finding: `Body != http.NoBody` is transport-implementation identity, not wire-body evidence
- Decision: accept known-empty `ContentLength == 0` only when any preserved header is digit-only zero; reject one-byte probing because it can block until timeout
- Scope: `internal/httpapi` only; no UI, mutation, order, LIVE, gate, or lifecycle changes
- RED evidence: actual TLS HTTP/2 bodyless GET/HEAD/stream all returned stable 400 before the fix
- GREEN evidence: the same test passes after protocol-neutral field/header classification; positive, malformed, signed-zero and overflow preserved headers fail closed
- Test/maintainability review: APPROVE WITH SUGGESTIONS, blocker 0; unknown-length stream, no-dispatch counter and header boundary coverage accepted
- Security review: APPROVE/CLEAN, blocker/Critical/High/Medium/Low 0; two Low findings were fixed and re-reviewed
- Security notes accepted: preserved positive/malformed headers and signed-zero syntax are rejected before read/SSE handler dispatch
- Verification: full repository tests, `internal/httpapi` race, vet, Windows cross-build/test compile, logic-map and strict OpenSpec validation pass
- Residual: raw malformed-frame tests rely on Go `net/http` protocol enforcement; no custom proxy/parser hop exists
- CI: origin/main commit `32fbb7af46489412d95413e9f9bc05768ae5bd69`, GitHub Actions run `30725043818` passed
- Deployment: Compose `tossos` and `httpapi` rebuilt and healthy; certificate-verified HTTP/2 negotiated as version 2
- Canary: engine/positions/orders/candidates/settings/optimization returned 200 twice; console health returned 200; mutation absence returned 405
- Degraded dependency parity: performance and SSE returned stable 503 under both HTTP/1.1 and HTTP/2 because the pre-existing `performance.db` is absent; neither returned `BODY_NOT_SUPPORTED`
- Status: implementation, independent reviews, CI, deployment and protocol-parity canary complete
