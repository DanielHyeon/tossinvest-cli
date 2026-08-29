# Function Logic Map: `markerRefusesStart`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a056-autostart-survives-container-recreate
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 change로 바뀌지 않음 | as before | as before | as before |

## Branches and early returns

새로 도입한 규칙 함수. 신선한 자문 마커가 기동을 거부해도 되는 조건은 **다른 관측이 동의할 때**뿐이다 — 프로세스가 실제로 보이거나, 열거가 실패해 부재를 주장할 수 없거나. 인라인 조건이 아니라 이름 있는 함수인 이유는 다음 편집이 이유까지 함께 읽게 하기 위해서다. 이 함수가 대체한 모양은 순차 `return` 두 개였고, 자문 검사가 먼저라 그 아래 프로세스 검사는 도달 불가였다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the one path through this function | 없음 — 순수 함수, 인자만 읽는다 | unchanged behaviour | `TestAGhostMarkerDoesNotRefuseAStart`, `TestAFreshMarkerWithALiveProcessStillRefuses`, `TestEnumerationFailureKeepsTheRefusal`, `TestMarkerRefusesStartOnlyWithCorroboration`, `TestNoPathRefusesOnMarkerFreshnessAlone` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 0 call sites, contract unchanged by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- 없음 — 순수 함수, 인자만 읽는다.
- 새 브로커 호출·config key·audit 레코드·POST 라우트 없음.

## Safety conclusion

- Safe edit boundary: 엔진 기동 전 안내 판정.
- 이 change는 주문 제출·손절·익절·사이징·Guardian 판정·원장·대사·인증·체결 어느 경로도 바꾸지 않는다. 바꾸는 것은 엔진 기동 **전 안내** 판정 하나이며, 배타는 그대로 journal flock이 담당한다. 방향은 보수적이다 — 엔진이 떠 있어야 exit 루프가 돌므로, 유령 마커로 인한 감시 공백을 없애는 것은 손절 즉시성을 강화한다.
