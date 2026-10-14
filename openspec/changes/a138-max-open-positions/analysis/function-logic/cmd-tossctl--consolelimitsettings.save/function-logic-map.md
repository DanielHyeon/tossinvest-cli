# Function Logic Map: `Save`

- Source: `cmd/tossctl/limitsettings.go` (49-78)
- Qualified: `consoleLimitSettings.Save`
- AST evidence: `ast.json`(현재) — 현재 AST branches 6 · returns 2
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 2e04bcffd343777e…) — 편집 전 갈래 수 5
- Risk scan: `risk-pattern-report.md`

**역할.** 콘솔 한도 저장 seam — 쓰기 + 필드별 audit.

**a138 편집.** 저장 뒤 파일 값으로 `recordOpenPositionsSave` 한 줄(best-effort, 0 = 「미설정 — 기본 6」).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := s.svc.SaveEngineGateLimits(l); err != nil {` (if 51) | — | 분기 본문 참조 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B2 | `if beforeErr == nil {` (if 57) | — | 분기 본문 참조 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B3 | `} else {` (else 59) | — | 분기 본문 참조 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B4 | `if after, err := s.svc.LoadRawEngineGate(); err == nil {` (if 64) | — | 분기 본문 참조 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B5 | `if beforeErr == nil {` (if 66) | — | 분기 본문 참조 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |
| B6 | `if beforeErr == nil {` (if 72) | — | 분기 본문 참조 | `go test ./cmd/tossctl -run 'A136|A138|Limit'` · `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` |

Exact AST return positions: 52 · 77

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 6종 — 대표: `before.Limits`, `recordDailyBuySave`, `recordLimitSave`, `recordOpenPositionsSave`, `s.svc.LoadRawEngineGate`, `s.svc.SaveEngineGateLimits`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
