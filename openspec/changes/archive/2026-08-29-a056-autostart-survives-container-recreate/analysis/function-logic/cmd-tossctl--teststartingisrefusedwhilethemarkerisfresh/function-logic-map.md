# Function Logic Map: `TestStartingIsRefusedWhileTheMarkerIsFresh`

- Source: `cmd/tossctl/engineproc_test.go`
- AST evidence: `ast.json` (revision: base)
- Change: a056-autostart-survives-container-recreate
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 change로 바뀌지 않음 | as before | as before | as before |

## Branches and early returns

삭제됐다. (마커 fresh, 프로세스 미관측) 칸을 거부로 고정하던 테스트이고, 이 change가 뒤집는 칸이 정확히 그 칸이다. `TestAGhostMarkerDoesNotRefuseAStart`가 대체하며 그 근거는 대체 테스트의 주석과 branch-test-map.md에 적혀 있다. 아래 증거는 base revision — 이 함수가 존재한 마지막 revision이다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if at line 155 | 없음 — 테스트 코드 | unchanged behaviour | `TestStartingIsRefusedWhileTheMarkerIsFresh` (this function is the test) |
| B2 | if at line 163 | 없음 — 테스트 코드 | unchanged behaviour | `TestStartingIsRefusedWhileTheMarkerIsFresh` (this function is the test) |
| B3 | if at line 166 | 없음 — 테스트 코드 | unchanged behaviour | `TestStartingIsRefusedWhileTheMarkerIsFresh` (this function is the test) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| see `ast.json` calls | 12 call sites, contract unchanged by this change | no retry introduced | AST + CodeGraph |

## State mutations and fallbacks

- 없음 — 테스트 코드.
- 새 브로커 호출·config key·audit 레코드·POST 라우트 없음.

## Safety conclusion

- Safe edit boundary: 엔진 기동 전 안내 판정.
- 이 change는 주문 제출·손절·익절·사이징·Guardian 판정·원장·대사·인증·체결 어느 경로도 바꾸지 않는다. 바꾸는 것은 엔진 기동 **전 안내** 판정 하나이며, 배타는 그대로 journal flock이 담당한다. 방향은 보수적이다 — 엔진이 떠 있어야 exit 루프가 돌므로, 유령 마커로 인한 감시 공백을 없애는 것은 손절 즉시성을 강화한다.
