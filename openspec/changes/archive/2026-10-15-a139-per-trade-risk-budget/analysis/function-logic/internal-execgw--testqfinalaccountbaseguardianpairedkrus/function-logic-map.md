# Function Logic Map: `TestQFinalAccountBaseGuardianPairedKRUS`

- Source: `internal/execgw/riskguardian_account_base_testseam_test.go` (25-72)
- Qualified: `TestQFinalAccountBaseGuardianPairedKRUS`
- AST evidence: `ast.json`(현재) — 현재 AST branches 8 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` e23432f073a3c6a7…) — 편집 전 갈래 수 8
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — KR · US 짝 계약(같은 권위 경로로 같은 q_final).

**a139 편집.** rig 정책 예산을 옛 값으로 명시 — 1/3 예산은 US 위험 상한만 묶어 짝을 깸(시험이 재는 것이 예산 값이 아님).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, test := range []struct {` (range 26) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run PairedKRUS` |
| B2 | `if err != nil {` (if 43) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run PairedKRUS` |
| B3 | `if err != nil {` (if 48) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run PairedKRUS` |
| B4 | `if precheck.QCandidate() != 20 \|\| precheck.QFinal() != 10 \|\| rig.collections != 0 {` (if 51) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run PairedKRUS` |
| B5 | `if err != nil {` (if 55) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run PairedKRUS` |
| B6 | `if err != nil {` (if 59) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run PairedKRUS` |
| B7 | `if err != nil {` (if 63) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run PairedKRUS` |
| B8 | `if limits.AccountBaseFX == nil \|\| limits.AccountBaseFX.QuoteCurrency != request.Currency \|\|` (if 66) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run PairedKRUS` |

Exact AST return positions: (없음)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 18종 — 대표: `bindPairedAccountBaseFXForTest`, `context.Background`, `execgw.DecodeLimits`, `fixedIDs`, `fx.Digest`, `guardianPolicy`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a139 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a139 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
