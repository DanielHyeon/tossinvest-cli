# Function Logic Map: `engineLogPath`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json` (revision: base)
- Change: a056-autostart-survives-container-recreate
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 change로 바뀌지 않음 | as before | as before | as before |

## Branches and early returns

이 change가 편집하지 않았다. `markerRefusesStart`를 바로 위에 삽입하면서 diff hunk가 인접 줄까지 닿아 대상 목록에 들어왔다. 본문은 base와 동일하다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | the one path through this function | 없음 | unchanged behaviour | `TestAGhostMarkerDoesNotRefuseAStart`, `TestAFreshMarkerWithALiveProcessStillRefuses`, `TestEnumerationFailureKeepsTheRefusal`, `TestMarkerRefusesStartOnlyWithCorroboration`, `TestNoPathRefusesOnMarkerFreshnessAlone` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 1 call sites, contract unchanged by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- 없음.
- 새 브로커 호출·config key·audit 레코드·POST 라우트 없음.

## Safety conclusion

- Safe edit boundary: 엔진 기동 전 안내 판정.
- 이 change는 주문 제출·손절·익절·사이징·Guardian 판정·원장·대사·인증·체결 어느 경로도 바꾸지 않는다. 바꾸는 것은 엔진 기동 **전 안내** 판정 하나이며, 배타는 그대로 journal flock이 담당한다. 방향은 보수적이다 — 엔진이 떠 있어야 exit 루프가 돌므로, 유령 마커로 인한 감시 공백을 없애는 것은 손절 즉시성을 강화한다.
