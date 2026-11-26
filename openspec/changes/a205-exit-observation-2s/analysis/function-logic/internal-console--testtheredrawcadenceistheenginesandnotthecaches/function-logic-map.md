# Function Logic Map: `TestTheRedrawCadenceIsTheEnginesAndNotTheCaches`

> **편집 후 지도(a205 GREEN 뒤).** 편집은 B1 의 기대값 리터럴 `5*time.Second` → `2*time.Second` 와 그 실패 문구 한 줄뿐 — 분기 수 · 위치 무변경(편집 전 base `e2921fc00` 은 B1@102 · B2@106 · B3@110 · B4@114 — 같은 파일 위쪽 `TestTheOverviewSpendsNothingHoweverOftenItReloads` 에 a205 주석 한 줄이 들어가 전부 +1, 순서 · 개수 같아 재번호 없음).

- Source: `internal/console/line_cadence_test.go` (102-119)
- AST evidence: `ast.json` — branches **4** / returns **0** / calls 7, source SHA-256 `442185be09849ca76ac067b31e105a75185d95ed39ff5a564c794fe80c1cf653`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/console/line_cadence_test.go --func TestTheRedrawCadenceIsTheEnginesAndNotTheCaches`.
- Risk scan: `risk-pattern-report.md`. a166 census 해당 없음(콘솔 시험 파일).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `lineRefreshInterval` | 엔진 관측 주기와 같은 값(2s) | `engine.DefaultExitObservationInterval` 사본(콘솔은 engine import 금지) | B1 `t.Errorf` |
| `holdingsTTL` | 재로드 주기보다 큼 | `internal/console/holdings.go` | B2 `t.Errorf` |
| 두 화면 `RefreshSeconds()` | `lineRefreshSeconds()` 와 같음 | `line_cadence.go:74` | B4 `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@103) | `lineRefreshInterval != 2*time.Second` **(a205 편집)** | `t.Errorf` | — | 이 시험 자신 — RED(5s 에서 실패) · GREEN(2s) |
| B2 (if@107) | `lineRefreshInterval >= holdingsTTL` | `t.Errorf` | — | 이 시험 — 2s < TTL 이라 거짓 갈래 |
| B3 (range@111) | positions · overview 두 화면 순회 | `RefreshSeconds()` 호출 | — | 이 시험 |
| B4 (if@115) | 화면 재로드 초 ≠ `lineRefreshSeconds()` | `t.Errorf` | — | 이 시험 — 2 == 2 라 거짓 갈래 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionsPage.RefreshSeconds` · `overviewPage.RefreshSeconds` | 화면이 실제로 쓰는 재로드 초 | 없음(순수) | AST calls@112 · @113 |
| `lineRefreshSeconds` | 기대값 | 없음 | AST calls@115 · @116 |
| `t.Errorf` | 판정 | — | AST |

## State mutations and fallbacks

- 없음(읽기만).

## Safety conclusion

- Safe edit boundary: B1 의 기대 리터럴 · 문구만. 분기 구조 · 다른 단언 무변경.
- High-risk impact: no(콘솔 표시 주기 시험). 생산 쪽 대응은 `line_cadence.go` 상수 하나.
