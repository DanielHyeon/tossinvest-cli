# Function Logic Map: `LoadRawEngineGate`

- Source: `internal/config/limits_io.go` (42-78)
- Qualified: `Service.LoadRawEngineGate`
- AST evidence: `ast.json`(현재) — 현재 AST branches 4 · returns 4
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6c8fdc5d37028679…) — 편집 전 갈래 수 4
- Risk scan: `risk-pattern-report.md`

**역할.** 콘솔이 보여 줄 적힌 그대로의 게이트 블록 읽기.

**a138 편집.** `MaxOpenPositions` 한 칸.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if errors.Is(err, os.ErrNotExist) {` (if 44) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B2 | `if err != nil {` (if 47) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B3 | `if err := json.Unmarshal(data, &doc); err != nil {` (if 56) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |
| B4 | `if raw.Enabled != nil {` (if 74) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|A138'` · `TestA138OpenPositionMaximumIsParsedFromTheGateBlock` |

Exact AST return positions: 45 · 48 · 57 · 77

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 4종 — 대표: `errors.Is`, `fmt.Errorf`, `json.Unmarshal`, `os.ReadFile`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
