# Function Logic Map: `candidateIntervals`

- Source: `cmd/tossctl/candidate.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> 분기가 없는 구성 함수다. ast.json의 `branches`는 null이고 `returns`는 line 383의 한
> 항목뿐이다. 이 change는 map 리터럴에서 항목 하나를 뺐고, 그 외에는 아무 것도 바꾸지
> 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | 없음 (params=0) | — | — |
| `official` 지역값 | `Every 15s / Floor 5s` | `add-candidate-discovery` D13의 per-source cadence 계약 | 값 자체는 이 change가 바꾸지 않았다 |
| `wts` 지역값 | `Every 5s / Floor 3s` | 같은 계약 | 같음 |
| 반환 map의 **키 집합** | `candidatesrc.Panel`이 만들 수 있는 id의 부분집합이어야 한다 | `internal/candidatesrc`의 `Panel` + 이 change가 추가한 `candidateschedule_drift_test.go` | 위반해도 런타임은 실패하지 않는다 — 호출을 만들지도 오류를 내지도 않는다. 그것이 이 결함이 조용한 이유이고, 이제 테스트가 잡는다 |
| 패널에는 있고 이 map에는 없는 원천 | **허용된다** | `internal/candidate/source.go`의 `unconfiguredFloor` | 위반이 아니다. 그 원천은 가장 보수적인 알려진 간격(15초)으로 읽힌다. 등호를 요구하면 그 설계가 테스트 위반이 된다 |

**이 change가 건드린 불변식**: 없음. 키가 하나 줄었고, 남은 키의 값은 그대로다.
`candidate.SourceOfficialGainers`는 같은 change에서 `candidatesrc.Panel`에서도 빠졌으므로
"패널에 있으면서 일정에 없다"는 상태를 만들지 않는다 — 그 상태였다면 `unconfiguredFloor`가
적용되어 폴링이 **잦아졌을** 것이다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 함수 전체가 하나의 경로다 (ast.json `branches`: null) | 지역 map 리터럴 생성뿐. 전역 상태·파일·네트워크 없음 | line 383의 `return`, 오류 없음 | `TestNoIntervalNamesASourceNoPanelBuilds` — 반환된 키 집합이 `candidatesrc.Panel`이 만들 수 있는 id 집합에 포함됨을 단언 |

**early return 없음.** 오류 반환 자체가 없다(results=1, error 아님).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | ast.json의 `calls`가 null이다. 이 함수는 아무 것도 호출하지 않는다 | — | ast.json `calls` |

**live config binding**: 없다. 간격은 리터럴이며 설정 파일·환경변수·플래그로 덮이지 않는다.
소비자는 `candidate.NewSchedule(candidateIntervals())` 두 곳(`scan`, `watch`)이고, `Schedule`이
`unconfiguredFloor`와 `engineYieldFactor`를 적용한다 — 그 두 값은 이 change가 건드리지 않았다.

## State mutations and fallbacks

- mutation 없음. 매 호출이 새 map을 만들어 돌려주며, `NewSchedule`이 다시 복사한다
  (`source.go`의 `NewSchedule`).
- fallback은 이 함수가 아니라 `Schedule.every`에 있다: 키가 없는 원천은
  `unconfiguredFloor`(15초)로 읽힌다. 이 change는 **패널과 일정에서 같이** 뺐으므로 그
  fallback이 이 원천에 적용될 일이 없다(D10).
- 이 change가 만든 fallback은 없다.

## Safety conclusion

- Safe edit boundary: map 리터럴의 항목 집합. 분기·호출·mutation이 전부 없는 함수이므로
  편집이 바꿀 수 있는 것은 반환 값의 키/값뿐이다.
- High-risk impact: **no.** 이 map은 발굴 폴링 간격만 정한다. 주문·손절·사이징·인증·체결
  경로에 닿지 않으며, 항목을 빼는 방향은 호출량을 늘리지 않는다 — 단, 패널에 남은 원천의
  항목을 뺐다면 `unconfiguredFloor`가 적용되어 오히려 잦아질 수 있었다. 그래서 이 편집은
  `candidatesrc.Panel`의 편집과 **같은 change 안에서** 이루어졌고, 두 집합의 관계는 이제
  `candidateschedule_drift_test.go`가 지킨다.
