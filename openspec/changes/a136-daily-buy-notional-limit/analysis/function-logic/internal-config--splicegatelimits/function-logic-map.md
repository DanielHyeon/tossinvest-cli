# Function Logic Map: `spliceGateLimits`

- Source: `internal/config/limits_io.go` (207-249, 편집 뒤)
- Qualified: `spliceGateLimits`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 9d08b616b9a8c621…, 편집 전 분기 수 11) — 현재 branches 11 · returns 8
- Risk scan: `risk-pattern-report.md`

**역할.** automation_gate 블록 안에 멤버를 키별로 바꾸거나 넣음.

**a136 편집.** 추가 멤버(지움)를 받을 수 있게 가변 인자 `extra ...gateMember` 를 더함 — 멤버 순회 루프 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `data` | 유효 JSON | 디스크 | 오류 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 209) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B2 | `for _, m := range members {` (range 216) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B3 | `if err != nil {` (if 218) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B4 | `if !found {` (if 221) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B5 | `if data, err = insertEmptyGate(data); err != nil {` (if 222) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B6 | `if gStart, gEnd, found, err = gateValueSpan(data); err != nil {` (if 225) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B7 | `} else if !found {` (else 227) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B8 | `} else if !found {` (if 227) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B9 | `if err != nil {` (if 233) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B10 | `if kFound {` (if 236) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |
| B11 | `if data, err = insertKey(data, gStart, gEnd, m.key, m.value); err != nil {` (if 244) | — | 분기 본문 참조 | `go test ./internal/config -run 'Splice / Limit / TestA136'` |

Exact AST return positions: 210 · 219 · 223 · 226 · 228 · 234 · 245 · 248

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `gateMembersOf` · `gateValueSpan` · `valueSpan` · `insertKey` | 키 위치 · 삽입 | 오류 | AST |

## State mutations and fallbacks

- 새 바이트 슬라이스.

## Safety conclusion

- Safe edit boundary: 멤버 목록 앞에 덧붙임만.
- High-risk impact: yes(설정 쓰기 경로)
