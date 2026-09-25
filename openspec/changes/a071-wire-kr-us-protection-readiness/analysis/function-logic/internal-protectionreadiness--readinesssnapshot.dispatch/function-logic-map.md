# Function Logic Map: `ReadinessSnapshot.Dispatch`

- Source: `internal/protectionreadiness/dispatch.go` (37-90)
- Revision: current — HEAD `648df8ef`; source_sha256 `246de6db4db431e80be57e878820374db9633e32f6d34d37913f8119183ed999`
- AST evidence: `ast.json` (`tools/logic-map` 추출기 출력)
- Risk scan: `risk-pattern-report.md` (ast-grep, `tools/logic-map/sgconfig.yml`)
- Extractor counts: AST branches 11 · returns 9 · calls 16
- Exact AST return positions: 41:3, 45:3, 53:4, 59:4, 64:3, 76:3, 81:3, 85:3, 89:2
- 2026-09-25 task 5.1 갱신: 분기 표는 `analysis/harness/51_refresh.py` 가 ast.json 에서 생성함. 아래 산문 절은 a071 저자 기록(2026-08-04) 그대로이며 base 뒤 다른 change 가 이 함수에 더한 동작은 기술하지 않음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| sealed snapshot | release plus independent KR/US market seals | `Assess`/`DefaultSnapshot` | corrupt selected market returns `state_corrupt`; peer market remains usable |
| dispatch scope | exact account/profile/market and attested order/session/quantity/trigger/replace/capability contract | persisted entry plan plus sealed supervisor contract | any missing/substituted field returns typed refusal before transport |
| current time | non-zero UTC instant inside attestation lifetime | gateway clock | invalid/future/expired returns typed refusal |

## Branches and early returns

| Branch | AST anchor | Source text at anchor | Coverage at HEAD (untagged package suite, per-test runs) |
|---|---|---|---|
| B1 | if at 39:2 | `if snapshot.release != ReadinessRelease {` | NOT entered 0/32 |
| B2 | if at 43:2 | `if scope.AccountID == "" \|\| scope.ProfileID == "" \|\| !validMarket(scope.Market) \|\| now.IsZero() {` | NOT entered 0/32 |
| B3 | switch at 48:2 | `switch scope.Market {` | evaluated 5/32 |
| B4 | case at 49:2 | `case MarketKR:` | entered 5/32 |
| B5 | if at 51:3 | `if snapshot.krSeal != marketVerdictSeal(snapshot.release, verdict) {` | entered 2/32 |
| B6 | case at 55:2 | `case MarketUS:` | entered 3/32 |
| B7 | if at 57:3 | `if snapshot.usSeal != marketVerdictSeal(snapshot.release, verdict) {` | NOT entered 0/32 |
| B8 | if at 63:2 | `if verdict.State != Wired \|\| verdict.Code != RefusalNone {` | entered 1/32 |
| B9 | if at 67:2 | `if provenance.AccountID != scope.AccountID \|\| provenance.ProfileID != scope.ProfileID \|\| provenance.Ser...` | entered 2/32 |
| B10 | if at 79:2 | `if provenance.IssuedAt.IsZero() \|\| provenance.ExpiresAt.IsZero() \|\| now.Before(provenance.IssuedAt) {` | NOT entered 0/32 |
| B11 | if at 83:2 | `if !now.Before(provenance.ExpiresAt) {` | entered 1/32 |

## Author concept rows (2026-08-04 — N 번호는 AST 분기 번호가 아님)

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| N1 | release mismatch | none | `state_corrupt` | corrupt snapshot test |
| N2 | invalid dispatch identity/contract | none | `invalid_attestation` | scope substitution matrix |
| N3 | selected KR or US market seal mismatch | none | `state_corrupt` | market-isolation corruption test |
| N4 | verdict is not exactly WIRED/none | none | stored typed refusal | missing/expired evidence tests |
| N5 | provenance or exact contract mismatch | none | `scope_mismatch` | order/session/quantity/trigger/replace/capability substitution tests |
| N6 | issued/expiry window invalid | none | invalid/expired | expiry tests |
| N7 | all checks pass | none | allowed with sealed snapshot ID | valid dispatch test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `marketVerdictSeal` | authenticate the selected immutable market verdict | pure/no retry | CodeGraph + AST |
| `validDigest` | validate build/evidence/supervisor/capability digests | pure/no retry | current HEAD |

## State mutations and fallbacks

- Pure decision function; no file, durable-state, toggle or broker mutation.
- A corrupt KR sub-snapshot must not invalidate a separately sealed US sub-snapshot and vice versa.

## Safety conclusion

- Safe edit boundary: extend the already sealed dispatch preimage; do not add fallback/default contract values.
- High-risk impact: yes — this is the final exposure-raising protection gate.
