# Function Logic Map: `TestTheCacheIntervalsAreTheirOwnConstants`

> **편집 후 지도(a205 GREEN 뒤).** 편집은 B4 가 순회하는 리터럴 목록의 한 원소 `"const engineRuntimeInterval = 5 * time.Second"` → `2 * time.Second` 뿐.
> 같은 파일의 패키지 수준 `wantRuntimeInterval` 5s → 2s 는 함수 본문 밖(B2 의 비교 대상 값). 분기 수 · 위치 무변경(base 도 B1@652 · B2@655 · B3@658 · B4@662 · B5@666).

- Source: `internal/console/position_policy_cache_test.go` (651-672)
- AST evidence: `ast.json` — branches **5** / returns **0** / calls 6, source SHA-256 `95da37bb65ca91b23527ed19529d3ebd06d9e68c371cc981eda07da1becea912`(파일 단위).
  명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/console/position_policy_cache_test.go --func TestTheCacheIntervalsAreTheirOwnConstants`.
- Risk scan: `risk-pattern-report.md`(패턴 0).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `engineLifecycleInterval` | 30s(쓰기 커넥션 경합 — 무변경) | `position_policy_cache.go` | B1 `t.Errorf` |
| `engineRuntimeInterval` | 2s — 엔진 관측 주기 이하(operator-console SHALL) | `position_policy_cache.go`, 엔진 상수 사본 | B2 `t.Errorf` |
| 두 간격이 서로 다름 | 30s ≠ 2s | design a081 D2 | B3 `t.Error` |
| 소스 리터럴 | 두 상수가 유도식이 아닌 리터럴 | `packageFiles` 가 읽은 비시험 소스 | B5 `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@652) | lifecycle ≠ 30s | `t.Errorf` | — | 이 시험(거짓 갈래) |
| B2 (if@655) | runtime ≠ `wantRuntimeInterval`(2s, a205) | `t.Errorf` | — | 이 시험 — RED(5s) · GREEN |
| B3 (if@658) | 두 간격이 같음 | `t.Error` | — | 이 시험(거짓 갈래) |
| B4 (range@662) | 리터럴 두 줄 순회 — 둘째 원소 **(a205 편집)** | — | — | 이 시험 |
| B5 (if@666) | 소스에 리터럴이 없음 | `t.Errorf` | — | 이 시험 — RED(`2 * time.Second` 리터럴 없음) · GREEN |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles` | 비시험 소스 읽기(`static_test.go:36`) | 읽기 실패면 `t.Fatalf` | AST calls@661 |
| `strings.Contains` | 리터럴 존재 | 없음 | AST calls@666 |
| `t.Errorf` · `t.Error` | 판정 | — | AST |

## State mutations and fallbacks

- 없음(읽기만).

## Safety conclusion

- Safe edit boundary: B4 리터럴 원소 하나(+ 함수 밖 기대 상수). 분기 구조 무변경.
- High-risk impact: no(콘솔 캐시 간격 시험). 생산 쪽 대응은 `position_policy_cache.go` 상수 하나.
