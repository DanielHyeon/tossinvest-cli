# Review record

## Proposal freeze — attempt 1

Date: 2026-07-31

Independent read-only Codex review covered product, UX, adversarial
engineering/security, and operator-DX concerns.

Verdict: **REJECT**

Blocking findings:

- P0: the body limit must wrap outside `mutating`, because that middleware calls
  `ParseForm` before the handler.
- P0: submitted replacement credentials require an isolated token cache or an
  old valid access token can falsely authenticate them.
- P0: persistence, audit, and restart failures must never produce a "started"
  notice; save-success audit ordering must be truthful.
- P1: rejected environment-managed credentials cannot be repaired by saving a
  lower-precedence file and require explicit container-environment guidance.

Resolution:

- Design, delta specs, and tasks now specify the outer body-limit order,
  isolated temporary validation cache and cleanup, post-save normal-cache
  invalidation, truthful save/audit/restart ordering, and environment-managed
  credential refusal.

## Proposal freeze — attempt 2

Date: 2026-07-31

Verdict: **REJECT**

Blocking findings:

- P1: concurrent setup/restart requests could validate one credential generation
  and restart with another.
- P1: post-persistence failure semantics needed to state that credentials remain
  saved while soak was not started.
- P1: the outer body wrapper must be proven not to bypass the existing
  session/origin/CSRF authorization chain.

Advisories:

- Specify deterministic oversize status, fixed secret-free refresh errors, and
  concrete environment variable names.

Resolution:

- Design, delta specs, and tasks now require one console mutex spanning each
  full credential-to-restart transaction, deterministic 413 behavior while
  preserving all existing gates, an explicit "saved, not started" partial
  state, fixed refresh classification, and environment variable names without
  values.

## Proposal freeze — attempt 3

Date: 2026-07-31

Verdict: **APPROVE**

Findings: P0 none, P1 none, P2 none.

The independent reviewer approved the serialized credential generation,
outer-body-limit ordering, isolated validation cache, truthful partial-state
semantics, environment-source handling, automatic token refresh, and
secret-free boundaries.

Implementation route note: GET and POST use `/openapi/login` and
`/openapi/login/save` respectively, allowing the static route table to prove
that the write path is explicitly mutation-gated. This does not change the
approved one-action behavior or any requirement.

Implementation middleware note: the route-specific limit is an optional
parameter on the existing `mutating` wrapper and is installed before its
`ParseForm`. Existing routes omit it. This preserves the approved ordering while
letting static guards continue to recognize the mutation wrapper directly.

## Pre-Edit Gate

- change id / task id: `automate-soak-openapi-onboarding` / 1.1-3.2
- 대상 심볼: `internal/console.(*Console).routes`,
  `internal/console.(*Console).handleSoakRestart`, `cmd/tossctl.runConsole`;
  remaining implementation is new narrow leaf code.
- CodeGraph definition/callers/callees/impact: current route→mutation→handler
  chain, CLI seam assembly, official loader/validator/token cache, and restart
  process seam captured under `analysis/code-context/`.
- CodeGraphContext 후보와 evidence reconciliation:
  `analysis/code-context/codegraphcontext-context.md` and
  `evidence-reconciliation.md`; no unresolved conflict.
- 기존 동작 파악 근거: base `fa57d98`, current HEAD, CodeGraph source,
  `restart_test.go`, `remote_test.go`, `static_test.go`, `openapi_test.go`,
  `soakproc_test.go`, and token/credential tests.
- Function Logic Map / Branch Test Map:
  `analysis/function-logic/` for all three modified existing functions;
  `check_analysis.py` passes before edit.
- upstream 상속 테스트 영향: no order/trading path; full inherited test suite
  remains mandatory.
- 실패 테스트 선행 작성: yes, tasks 1.2 and 1.3.
- 설정·DB·journal 변경과 rollback: credential file and token cache only;
  no schema change. Rollback deploys previous image and retains the compatible
  0600 credential file.
- 안전 불변식 §0 위반 여부 검토: 통과. No LIVE order, engine start, toggle
  flip, or real soak/API execution is permitted in tests.

## Implementation review — remediation rounds

Date: 2026-07-31

### Attempt 1 — REJECT

P1 findings:

- Plaintext loopback could reach credential ingress.
- A saved-but-not-started generation could bypass ordinary restart.
- Whitespace environment credentials did not use the exact source precedence
  used by the child process.

Resolution:

- Both credential routes now require direct TLS, pending generations reopen
  setup without spawn, and raw non-empty environment pairs exactly match the
  official loader.

### Requirements re-review — REJECT, then APPROVE

P1 findings covered HTTPS-mode scope, save-error marker retention, canonical
persistent marker path, rollback quarantine, environment-marker dormancy, and
explicit test coverage.

Resolution:

- The contract now scopes guided onboarding to configured HTTPS, retains the
  marker after any attempted save, fixes its Compose path, documents rollback
  quarantine and dormant environment behavior, and enumerates the required
  tests. A separate final requirements review returned **APPROVE**, with no
  P0/P1/P2 findings.

### Attempt 2 — REJECT

P1 finding:

- The draft application-session wording contradicted the already approved
  trusted-network deployment, where authenticated VPN membership is the access
  boundary and no separate application login is required.

Resolution:

- Token mode still requires its application session; trusted-network mode
  consumes allowed VPN peer membership through `session0`, while TLS, peer,
  Host/origin, method, CSRF, and audit boundaries remain mandatory.

### Attempt 3 — REJECT

P1 findings:

- Retrying marker creation could destructively rewrite an existing fail-closed
  marker.
- A changed environment pair could be falsely accepted through a normal token
  cached for an older pair.

Resolution:

- Marker creation is exclusive and treats an existing marker as immutable.
- Environment preflight validates through an isolated temporary cache, removes
  it, and invalidates the normal cache on success.

### Attempt 4 — REJECT

P1 findings:

- An old soak could recreate its credential generation in the normal token
  cache after preflight invalidation but before it was stopped.
- Overwriting an existing permissive credential file did not guarantee final
  0600 mode.

Resolution:

- `restartSoak` waits for all old processes, applies a second cache
  invalidation fence immediately before spawn, and starts nothing if that fence
  fails.
- `official.SaveCredentials` writes and fsyncs a new 0600 temporary file,
  atomically replaces the target, verifies a regular 0600 result, and fsyncs
  the parent directory. Any error remains fail-closed through the onboarding
  marker.

### Final implementation and requirements review — APPROVE

The independent read-only reviewer examined the complete base-to-worktree diff,
all active artifacts, security/access-mode gates, failure ordering, token
generation fence, atomic credential publication, and isolated tests.

Verdict: **APPROVE**

Findings: P0 none, P1 none.

The reviewer invoked no real API, soak, engine, toggle, or order path.
