# Function Logic Map: `Assess`

- Source: `internal/protectionreadiness/readiness.go` (103-170)
- Revision: current — HEAD `648df8ef`; source_sha256 `d95f486a1cef853f912c9ab6330896f930cdc6b73f2efdce5aa138bd0510337c`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 17 · returns 1 · calls 16
- Exact AST return positions: 169:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| policy/state/time | sealed and monotonic | pinned policy and durable state | per-market UNWIRED |
| market evidence | exact signed KR or US scope | attestation + supervisor binding | peer market unchanged |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 109:2 | `if stateValid {` | entered 23/32 |
| B2 | if at 112:2 | `if stateValid && timeValid && !timeRollback {` | entered 23/32 |
| B3 | if at 114:3 | `if result.NextState.TrustedTimeFloor.IsZero() \|\| input.Time.Now.After(result.NextState.TrustedTimeFloor) {` | entered 22/32 |
| B4 | range at 118:2 | `for _, market := range []Market{MarketKR, MarketUS} {` | entered 24/32 |
| B5 | if at 120:3 | `if !present {` | entered 18/32 |
| B6 | switch at 124:3 | `switch {` | evaluated 24/32 |
| B7 | case at 125:3 | `case !policyValid:` | NOT entered 0/32 |
| B8 | case at 127:3 | `case !stateValid:` | entered 1/32 |
| B9 | case at 129:3 | `case !timeValid:` | entered 1/32 |
| B10 | case at 131:3 | `case timeRollback:` | entered 2/32 |
| B11 | case at 133:3 | `default:` | entered 23/32 |
| B12 | if at 136:4 | `if code == RefusalNone {` | entered 16/32 |
| B13 | if at 150:3 | `if market == MarketKR {` | entered 23/32 |
| B14 | else at 152:10 | `} else {` | entered 10/32 |
| B15 | if at 156:2 | `if result.StateCommitAllowed {` | entered 23/32 |
| B16 | else at 161:9 | `} else {` | entered 3/32 |
| B17 | if at 158:3 | `if result.NextState.seal != input.State.seal {` | entered 23/32 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1-B3 | durable state/time floor validity | pure next-state only | fail closed | state tests |
| B4-B14 | each market evidence and verification result | per-market verdict | typed refusal | KR/US isolation tests |
| B15-B17 | state commit allowed | reseal or preserve preimage | assessment result | rollback tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyAttestation` | signature/scope/supervisor proof | no retry or external mutation | CodeGraph + AST |

## State mutations and fallbacks

- No external mutation; result contains a pure durable-state successor and immutable paired snapshot.

## Safety conclusion

- Safe edit boundary: add exact account/profile/supervisor provenance to already-verified verdicts only.
- High-risk impact: yes; every new field is included in the snapshot seal.
