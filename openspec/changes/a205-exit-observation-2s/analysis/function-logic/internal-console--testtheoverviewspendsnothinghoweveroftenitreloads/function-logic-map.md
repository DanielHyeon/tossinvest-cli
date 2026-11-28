# Function Logic Map: `TestTheOverviewSpendsNothingHoweverOftenItReloads`

> **편집 후 지도(a205 구현 리뷰 r1 #1 처분 뒤).** 편집은 `const redraws = 10` → `16` 과 그 위 한국어 주석 한 줄뿐. 분기 개수 · 순서 무변경
> (편집 전 B1@74 · B2@79 · B3@84 · B4@91 — 주석 한 줄로 전부 +1, 재번호 없음). 편집 전 AST 는 같은 4 분기 · 같은 호출 10.

- Source: `internal/console/line_cadence_test.go` (69-96)
- AST evidence: `ast.json` — branches **4** / returns **0** / calls 10, source SHA-256 `442185be09849ca76ac067b31e105a75185d95ed39ff5a564c794fe80c1cf653`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/console/line_cadence_test.go --func TestTheOverviewSpendsNothingHoweverOftenItReloads`.
- Risk scan: `risk-pattern-report.md`(패턴 0).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `redraws × lineRefreshInterval` | `holdingsTTL`(30s) 초과 — 그래야 get 은 창 안에서 한 번 갱신하고 peek 는 0 이라 둘이 갈림 | `holdings.go` TTL · `line_cadence.go` 주기 | B1 `t.Fatalf` |
| 개요 화면 브로커 호출 수 | 0 | console-operator-overview D4(peek) | B3 `t.Errorf` |
| 포지션 화면 1회 렌더 뒤 호출 수 | ≥1(계수기가 브로커에 닿는다는 양성 대조) | holdingsCache.get | B4 `t.Error` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@75) | 창 ≤ TTL — **a205 이전 10회 × 2s = 20s 로 참이 되어 실패**(r1 #1), 16회 × 2s = 32s 로 거짓 | `t.Fatalf` | 시험 중단 | 이 시험 — RED(attempt1) · GREEN |
| B2 (for@80) | 재로드 16회 | `/dashboard` 렌더 · 시계 2s 전진 | — | 이 시험 |
| B3 (if@85) | 개요가 브로커를 부름 | `t.Errorf` | — | 이 시험(거짓 갈래) |
| B4 (if@92) | 포지션 화면도 호출 0(계수기 미배선) | `t.Error` | — | 이 시험(거짓 갈래 — 양성 대조) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newOverviewHarness` · `seedJournal` | 가짜 브로커 · 원장 | 실패면 하네스가 Fatal | AST calls@70 · @71 |
| `h.page` | 화면 렌더 | 하네스 | AST calls@81 · @91 |
| `h.clock.advance` | 재로드 간격만큼 가짜 시계 전진 | — | AST calls@82 |
| `h.holdings.count` | 브로커 호출 계수 | — | AST calls@85 · @92 |

## State mutations and fallbacks

- 가짜 시계 · 가짜 브로커 계수기만(시험 하네스 내부).

## Safety conclusion

- Safe edit boundary: 횟수 상수 하나. 단언 · 분기 무변경 — 창 > TTL 조건(B1)은 그대로 두고 그 조건을 만족하도록 횟수를 늘림(조건을 약화하지 않음).
- High-risk impact: no(콘솔 브로커 비용 시험).
