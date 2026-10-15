# Function Logic Map: `LimitsDigest`

- Source: `internal/execgw/riskguardian.go` (328-336)
- Qualified: `RiskGuardian.LimitsDigest`
- AST evidence: `ast.json`(현재) — 현재 AST branches 0 · returns 1
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 7bf1bbe7623c179e…) — 편집 전 갈래 수 0
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 바인딩이 결속하는 Guardian 한도 digest.

**a138 편집.** digest 입력에 `\nmax_open_positions=N` 을 더함(분기 무변경).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 직선 함수 — 분기 없음 | — | 정상 경로 | `go test ./internal/execgw -run 'LimitsDigest|A136|A138'` · `TestA138TheLimitsDigestBindsTheOpenPositionMaximum` |

Exact AST return positions: 335

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 3종 — 대표: `hex.EncodeToString`, `sha256.Sum256`, `strconv.Itoa`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
