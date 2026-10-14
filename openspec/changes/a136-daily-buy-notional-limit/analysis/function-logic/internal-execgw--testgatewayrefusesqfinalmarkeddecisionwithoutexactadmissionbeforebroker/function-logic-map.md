# Function Logic Map: `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker`

- Source: `internal/execgw/riskguardian_qfinal_test.go` (215-257, 편집 뒤)
- Qualified: `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 0b59cca08c88715e…, 편집 전 분기 수 6) — 현재 branches 6 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험 함수(생산 코드 아님).

**a136 편집.** 시험 fixture 에 a136 의 새 필수 칸을 채움 — EXPOSURE_RAISING 예약에 넉넉한 `DailyBuyLimit`(이 시험이 재는 것과 무관) (재추출 2026-10-14: a137 a1cfe0c2 가 같은 파일을 고쳐 소스 해시만 바뀜 — 이 함수의 분기 · 반환 · 호출 모양은 그대로.)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 219) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B2 | `if err != nil {` (if 223) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B3 | `if err != nil {` (if 227) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B4 | `if err != nil {` (if 244) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B5 | `if err != nil {` (if 248) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |
| B6 | `if !errors.As(err, &rejected) \|\| rejected.Reason != execgw.ReasonGuardianRiskBucketMismatch \|\| places != 0 {` (if 254) | — | 분기 본문 참조 | `go test ./internal/execgw` (이 함수를 쓰는 기존 시험) |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture 구성 | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 추가 — 시험이 단언하는 값 무변경.
- High-risk impact: no(시험 전용, 비례 원칙)
