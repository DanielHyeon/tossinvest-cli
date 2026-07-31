# Review

## Proposal-freeze review

- Date: 2026-07-31
- Change: `fix-console-origin-fallback`
- Base: `c06192799fa9faa744bf194c8d9636bf9441195c`
- Risk: High-risk authentication/session boundary; no trading-path symbols in
  the CodeGraph impact set.
- Voices: operator outcome, product scope, adversarial engineering/security,
  developer/operations experience.

### Findings and decisions

1. **P0 accepted: common gate makes every setting unusable.** The failure URL is
   already the exact configured HTTPS origin; the suffix `/restart` is the form
   action and is never part of origin comparison.
2. **P0 accepted: response and request contracts conflict.** Remote responses
   set `Referrer-Policy: no-referrer`, while `sameOrigin` rejects the absence of
   both `Origin` and `Referer`. A privacy-preserving browser can therefore never
   reach any POST handler.
3. **Security boundary retained.** Explicit Origin has final precedence,
   Referer is considered only when the Origin key is absent, and direct
   TLS+Host is considered only for CSRF-protected console mutations when both
   header keys are absent. Empty, whitespace, malformed, or multiple explicit
   values are final rejection evidence.
4. **Forwarding headers rejected.** The fix reads neither forwarded host nor
   forwarded protocol. TLS termination remains inside TossOS.
5. **CSRF retained.** Headerless canonical mutation requests still must pass
   the process-local CSRF gate before any handler.
6. **Scope rejected: multiple origins/wildcards.** This change does not allow
   `localhost` and `127.0.0.1` simultaneously and does not change VPN CIDR,
   certificate, Host, or token policy.
7. **Compatibility login remains strict.** `/login` continues to require a
   valid explicit Origin or Referer and does not receive TLS+Host fallback.
   Its token, rate-limit, session, and audit behavior are unchanged.

### Proposal-freeze verdict

SUPERSEDED by the independent adversarial review below. Rollback remains the
previous image/commit; no config, DB, journal, LIVE order, engine or operating
toggle migration exists.

## Independent adversarial proposal review

- Reviewer: isolated read-only Codex process
- Initial verdict: **REJECT**

### Blocking findings and resolutions

1. `sameOrigin` also protects `/login`, which has no CSRF gate. **Accepted.**
   The TLS+Host fallback is now isolated in `sameOriginForMutation`; `/login`
   keeps strict explicit-header validation.
2. `Header.Get` cannot distinguish an absent key from an explicitly empty key.
   **Accepted.** The target predicate checks canonical header-map key presence,
   requires exactly one non-empty value, and never falls through from invalid
   explicit evidence.
3. The first Function Logic Map mixed current AST branches with proposed
   branches. **Accepted.** Current and target branch sets are now separate, and
   post-GREEN AST refresh is mandatory.
4. The test matrix omitted empty/multiple headers, contradictory evidence,
   login behavior, and forwarding headers. **Accepted.** Each is now an
   explicit scenario and Branch Test Map row.

Because requirements changed materially, implementation remains blocked until
a second independent proposal review returns APPROVE.

### Second independent review

- Verdict: **REJECT**
- Blocking findings accepted:
  1. The current-branch map inverted AST branch B1 and omitted the
     missing-scheme/missing-host subconditions in B2. The map now records the
     exact current conditions and separates target IDs by strict (`SO-*`) and
     mutation-only (`SM-*`) predicates.
  2. Valid Origin plus invalid Referer was ambiguous. The spec now states that
     presence of Origin makes Referer unevaluated, and tests cover both
     contradiction directions.
  3. “Forwarding-only” coverage was vague. The planned test now names both
     `X-Forwarded-Host` and `X-Forwarded-Proto`.

Implementation remains blocked until the revised contract receives an
independent APPROVE verdict.

### Third independent review

- Verdict: **APPROVE**
- Blockers: none.
- Confirmed: the current B1/B2 map matches the captured AST and source hash;
  both Origin/Referer contradiction directions are unambiguous and mapped to
  tests; both `X-Forwarded-Host` and `X-Forwarded-Proto` are named in the
  rejection coverage.
- New P0/P1 findings: none.

The proposal-freeze and Pre-Edit gates are now approved for RED tests.

## Pre-Edit Gate

- change id / task id: `fix-console-origin-fallback` / 2.1–3.2
- 대상 심볼(패키지.함수): `console.remoteRuntime.sameOrigin`,
  `console.Console.mutating`; new leaf `sameOriginForMutation`
- CodeGraph definition/callers/callees/impact: definition
  `internal/console/remote.go:242`; callers `Console.mutating` and
  `remoteRuntime.loginPost`; depth-3 impact 11 console-only symbols; no
  trading/engine/Guardian/journal dependency.
- CodeGraphContext 후보와 evidence reconciliation:
  `analysis/code-context/codegraphcontext-context.md` and
  `evidence-reconciliation.md`; no conflicting candidate.
- 기존 동작 파악 근거: current HEAD, `remote_test.go` independent-gate tests,
  deployed reproduction, CodeGraph and Go AST.
- Function Logic Map / Branch Test Map:
  `analysis/function-logic/internal-console--remoteruntime.sameorigin/` and
  `analysis/function-logic/internal-console--console.mutating/`.
- upstream 상속 테스트 영향: yes; wrong-origin, wrong-Host, non-TLS and CSRF
  rejections stay covered.
- 실패 테스트 선행 작성: yes, task 2.1.
- 설정·DB·journal 변경과 rollback: none; revert function or previous image.
- 안전 불변식 §0 위반 여부 검토: 통과. No order, engine, toggle, exit, Guardian,
  credential or journal code is modified.

## Independent implementation/security review

- Reviewer: isolated read-only Codex process
- Verdict: **APPROVE**
- P0/P1 blockers: none.
- Confirmed:
  - Go `net/http` canonicalizes inbound header keys and preserves repeated
    fields as slices, matching the absent/empty/multiple implementation;
  - Origin precedence is final, `/login` remains strict, and only
    `Console.mutating` calls the mutation-only fallback;
  - headerless fallback requires direct TLS and exact configured Host:port and
    ignores forwarded host/protocol headers;
  - gate order remains POST → origin → form parse → CSRF → handler;
  - tests have no operational side effects.
- The reviewer's own full console rerun was socket-blocked by its read-only
  sandbox; focused tests/race/vet passed there. The Manager context separately
  ran and passed both the full console suite and full repository suite.
