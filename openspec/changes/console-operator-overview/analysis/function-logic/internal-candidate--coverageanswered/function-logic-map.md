# Function Logic Map: `coverageAnswered`

- Source: `internal/candidate/scan.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

본문이 바뀐 기존 함수다. 세 번째 인자 `panel` → `heard` 하나가 §5 리뷰 P1-1의 실제 수리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `sources` | 후보를 올린 적 있는 원천의 누적 목록 | `candidates.sources` 컬럼(누적, 병합) | 빈 목록이면 `present == 0`이라 false |
| `responded` | 응답하고 행을 나른 원천 | `Collect` | 여기 없는 지지자가 `heard`에 있으면 즉시 false |
| `heard` | 패널 + 일정이 넘긴 원천 | `Collect` | 여기 없는 지지자는 '없어진 원천'이라 판정에서 제외한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 지지자 순회 | 지역 카운터 | — | `TestOneSurvivingSupporterIsNotEnoughToCool` |
| B2 | `!heard[id]` — 설정에서 사라진 원천 | `present` 증가 안 함 | `continue` | `TestASupporterThatLeftThePanelDoesNotBlockCoolingForever`, `TestASourceTheSchedulePassedOverIsNotASourceThatIsGone` |
| B3 | `!responded[id]` — 들을 수 있었는데 답하지 않음 | 없음 | `false` 즉시 | `TestAScanDoesNotCoolASymbolItDidNotLookFor`, `TestASourceThatWasNotAskedDoesNotVouchForTheCandidatesItRaised` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | — | — | ast.json calls 없음 — 순수 함수 |

## State mutations and fallbacks

- 상태 변경 없음. 순수 술어다.
- 마지막 `return present > 0`: 지지자가 **전부** 없어진 후보는 냉각하지 않는다. 이번 스캔이 그 후보를 볼 위치에 없었기 때문이다.

## Safety conclusion

- Safe edit boundary: `heard`를 좁히거나 `present > 0` 요구를 없애는 방향은 §2-5와 §5 P1-1 두 결함을 동시에 되살린다
- High-risk impact: no — 순수 술어, 부작용 없음. 결과가 냉각을 좌우하므로 파괴 방향(느슨하게)만 위험하다.
