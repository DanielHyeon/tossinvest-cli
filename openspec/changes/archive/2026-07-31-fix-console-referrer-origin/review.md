# Review

## Risk classification

- Classification: High-risk security boundary, narrowly scoped.
- Reason: `remoteRuntime.security` fronts all remote console routes and
  `Console.render` plus shared templates define the effective document policy,
  although the change does not alter authentication, session, origin
  evaluation, CSRF, action audit, or any trading path.
- Required voices: proposal/architecture, adversarial engineering/security,
  implementation correctness, and deployed browser QA.

## Pre-Edit Gate

- change id / task id: `fix-console-referrer-origin` / 2.1-2.2
- 대상 심볼(패키지.함수): `internal/console.remoteRuntime.security`,
  `internal/console.Console.render`; non-function surface
  `internal/console.pageTemplates`
- CodeGraph definition/callers/callees/impact: definition at
  `internal/console/remote.go:276`; sole caller `Console.routes`; 49-symbol
  wrapper impact; calls response header setters, peer/CIDR/Host checks, and
  downstream handler. `Console.render` is at `internal/console/pages.go:376`,
  has two indexed impacts, one template-error branch, and later sets its own
  response headers. `pageTemplates` has shared and restart meta declarations.
- CodeGraphContext 후보와 evidence reconciliation:
  `analysis/code-context/`; broad console connectivity only, reconciled in
  favor of CodeGraph and current HEAD
- 기존 동작 파악 근거: current HEAD, existing remote security-header/peer/Host
  tests, real Chrome invalid-CSRF form reproduction, commit history
- Function Logic Map / Branch Test Map:
  `analysis/function-logic/internal-console--remoteruntime.security/` and
  `analysis/function-logic/internal-console--console.render/`
- upstream 상속 테스트 영향: no; console-only response header, verified by
  full inherited test suite
- 실패 테스트 선행 작성: yes; exact remote/rendered header and shared/restart
  meta assertions will fail against `no-referrer`; explicit `Origin: null`
  remains pinned as rejected
- 설정·DB·journal 변경과 rollback: none; rollback is previous image and all
  four prior policy literals in `remote.go`, `pages.go`, and `templates.go`
- 안전 불변식 §0 위반 여부 검토: 통과; no LIVE order, engine, risk, toggle,
  journal, or account side effect

## Proposal-freeze review

### Attempt 1 — REJECT (2026-07-31)

- P0: a one-line `remoteRuntime.security` change would be overwritten by
  `Console.render` and both HTML meta policies, while a `/login`-only assertion
  could produce a false positive.
- P1: the planned tests did not pin explicit `Origin: null` rejection.
- P1: scope, Function Logic Map, Pre-Edit Gate, and rollback omitted
  `Console.render` and both template policies.
- Resolution: proposal, design, delta spec, tasks, evidence, rollback, tests,
  and logic maps were expanded to cover all four effective policy surfaces.

### Attempt 2

REJECT (2026-07-31).

- P1: the named `origin_null_is_opaque` table case did not exist, and direct
  `sameOrigin` coverage alone would not prove TLS+Host fallback and handler
  ordering.
- Resolution: task 2.1, design, and the Branch Test Map now require a full
  `Console.mutating` regression with explicit `Origin: null`, canonical direct
  TLS and Host, valid CSRF, the origin-refusal response, and zero handler calls.

### Attempt 3

APPROVE (2026-07-31).

- The full `Console.mutating` regression plan includes explicit
  `Origin: null`, canonical direct TLS/Host, valid CSRF, origin-refusal text,
  and zero handler calls.
- Delta spec and both Branch Test Maps are consistent.
- Freeze base equals HEAD (`a8d43a5`), and production/test files were unchanged
  at approval time.
- Verdict: no unresolved P0/P1; production editing is unblocked.

## Independent implementation/security review

### Round 1 — BLOCK (2026-07-31)

- Code/security boundary passed: exactly two response-header and two meta-value
  substitutions; CSP, origin/CSRF/Host/TLS/peer logic unchanged; unit security
  coverage present.
- P1: post-implementation deployed Chrome evidence was still pending.
- Resolution: rebuilt and force-recreated the Compose service, confirmed
  healthy status, then submitted the actual `/restart` form with an
  intentionally invalid CSRF value. Chrome sent canonical Origin/Referer and
  received the CSRF-specific 403, never the origin refusal.

### Round 2

BLOCKED by reviewer sandbox environment (2026-07-31).

- The reviewer confirmed the recorded form probe satisfies every application
  condition and found no new code P0/P1.
- Its only blocker was inability to open `/var/run/docker.sock`, so it could not
  independently rerun `docker compose ps`.
- Resolution: the implementing host already captured `healthy`, failing streak
  zero, and the exact published port immediately before the browser probe.
  Round 3 evaluates the application evidence and this captured host state
  without requiring Docker socket access from the read-only reviewer sandbox.

### Round 3

APPROVE (2026-07-31).

- The recorded canonical Origin/Referer and CSRF-specific 403 from the
  intentionally invalid token resolve the missing post-deploy browser evidence.
- Combined with round 1's code/security boundary verdict, no unresolved P0/P1
  remains.
