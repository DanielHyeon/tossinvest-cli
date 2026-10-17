# Function Logic Map: `newProductionRiskGuardian`

- Source: `internal/app/engine/guardian_wiring.go` (71-103)
- Qualified: `newProductionRiskGuardian`
- AST evidence: `ast.json`(현재) — 현재 AST branches 3 · returns 4
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6acfab5929684cce…) — 편집 전 갈래 수 2
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 Guardian 생성 — 정책 유도 · 검증 · 생성.

**a138 편집.** Validate 앞에서 `config.WholeOpenPositions` 로 해석 값을 판정해, 운영자가 적은 값을 문구에 담아 ErrLimitsRequired — 새 분기 하나(freeze 리뷰 P2-3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if v := gate.EffectiveMaxOpenPositions(); !wholeOpenPositions(v) {` (if 79) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'ProductionPolicy|A136|A138'` · `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` |
| B2 | `if err := policy.Validate(); err != nil {` (if 84) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'ProductionPolicy|A136|A138'` · `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` |
| B3 | `if err != nil {` (if 99) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'ProductionPolicy|A136|A138'` · `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` |

Exact AST return positions: 80 · 85 · 100 · 102

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 8종 — 대표: `costs.DefaultModel`, `execgw.NewManualBuyBoard`, `execgw.NewRiskGuardian`, `fmt.Errorf`, `gate.EffectiveMaxOpenPositions`, `policy.Validate`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
