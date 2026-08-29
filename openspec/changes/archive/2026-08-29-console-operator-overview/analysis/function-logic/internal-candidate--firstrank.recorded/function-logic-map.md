# Function Logic Map: `FirstRank.Recorded`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. `Baseline.Recorded()`의 쌍둥이이고, 두 절반을 **둘 다** 요구하는 것이 요점이다 — 목록 길이 없는 순위는 정규화할 수 없고 `rank/0`은 +Inf라 비교되는 모든 임계를 통과한다(§1 7절이 같은 실패를 `rank_total == 0`에서 이미 만났다).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `f.Rank` | 1-based 위치. 0은 '이 삶에 기록된 순위 관측이 없다'이고 '목록 맨 아래'가 아니다 | `decodeFirstRank` 또는 `NoteFirstRank` | 0이면 false |
| `f.Total` | 목록 길이 | 같은 곳 | 0이면 false — 분모 0을 술어 단계에서 끊는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (무분기) | — | 없음 | `f.Rank > 0 && f.Total > 0` | `TestARankOfZeroIsNotAFirstSighting` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls null — 한 줄 순수 술어 |

## State mutations and fallbacks

- 상태 변경 없음.
- '부재 ≠ 0'을 타입 수준에서 지키는 자리다. 제로값 `FirstRank{}`는 **미기록**이고, D10에 따라 미기록은 통과가 아니다. §3 리뷰 P0-1이 같은 파일군에서 제로값이 '측정됨'으로 읽히던 결함을 실행으로 재현했다.

## Safety conclusion

- Safe edit boundary: `Rank > 0`만 보도록 좁히면 `rank/0 = +Inf`가 백분위로 흘러 모든 임계를 통과한다 — 금지
- High-risk impact: no — 순수 술어. 다만 `seen_late` veto의 측정 여부를 결정하므로 느슨해지는 방향만 위험하다.
