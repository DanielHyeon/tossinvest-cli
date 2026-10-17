# Function Logic Map: `marketVerdictSeal`

- Source: `internal/protectionreadiness/dispatch.go` (97-108)
- Revision: current — HEAD `648df8ef`; source_sha256 `246de6db4db431e80be57e878820374db9633e32f6d34d37913f8119183ed999`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 0 · returns 1 · calls 9
- Exact AST return positions: 98:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| verdict and provenance | all fields that can authorize a protection mutation, including quantity bounds and capability scope | verified attestation | omitted or modified field changes seal and snapshot identity |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | branchless happy path at 97:1 | `func marketVerdictSeal(release string, verdict Verdict) [32]byte {` | called 29/32 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | all canonical fields | hashes length-prefixed preimage | deterministic SHA-256 | seal tamper tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `hashStrings` | collision-safe length-prefixed digest | pure/no retry | current HEAD |

## State mutations and fallbacks

- Pure; no external state. Field order is protocol and changes require dispatch tests.

## Safety conclusion

- Safe edit boundary: append every authority-bearing scope field to the market seal and paired snapshot seal.
- High-risk impact: yes — omission permits scope substitution.
