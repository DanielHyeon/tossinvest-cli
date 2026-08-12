# Branch Test Map: `NewRuntime`

**GREEN 칸은 호출 지점을 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이다: AST 분기 12 · 이탈 7.

기존 커버리지의 정본은 `TestTheRuntimeRefusesWiringItCannotSupervise`
(`internal/app/engine/runtime_test.go:450`)의 일곱 케이스 표다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:184` 루프가 하나도 없으면 거부 | `TestTheRuntimeRefusesWiringItCannotSupervise` | no | **yes (기존)** |
| B2 | `:189` 루프 집합을 훑는다 | 같은 테스트 | no | **yes (기존)** |
| B3 | `:191` `switch` — 조건 없는 분기 | 같은 테스트 | no | **yes (기존)** |
| B4 | `:192` 이름 없는 루프를 거부 | 같은 테스트 | no | **yes (기존)** |
| B5 | `:195` 이름이 겹치는 루프를 거부 | 같은 테스트 | no | **yes (기존)** |
| B6 | `:197` `Run`이 nil인 루프를 거부 | 같은 테스트 | no | **yes (기존)** |
| B7 | `:201` `Health`가 nil이면 트리거·계좌를 안 본다 | `TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical` (`runtime_test.go:127`) — `Health` 없는 루프로 조립한다 | no | **yes (기존)** |
| B8 | `:204` 열거에 없는 트리거를 거부 | `TestTheRuntimeRefusesWiringItCannotSupervise` | no | **yes (기존)** |
| B9 | `:209` 승격 가능한 루프에 계좌가 없으면 거부 | 같은 테스트 | no | **yes (기존)** |
| B10 | `:222` 시계가 nil이면 시스템 시계 | `TestFiveConsecutiveFailuresEscalateOnceAndTheLoopKeepsRunning` (`runtime_test.go:276`)는 **가짜 시계를 넘긴다** — 즉 이 분기의 **거짓 쪽**만 덮인다 | no | **부분** |
| B11 | `:225` 임계 0이면 기본값 | 같은 테스트가 명시 임계를 넘긴다 — **거짓 쪽만** | no | **부분** |
| B12 | `:228` 간격 0이면 기본값 | `TestTheSupervisorPollsHealthWhileTheLoopsRun` (`runtime_test.go:405`)도 명시 간격을 넘긴다 — **거짓 쪽만** | no | **부분** |
| 이탈 `:231` | 정상 조립 | 위 전부 | no | **yes (기존)** |
| **신설** | **이름 없는 보조 실행자를 거부한다** | **a098 R — §3에 등록** | **예정 — 오늘은 필드가 없어 컴파일이 안 된다** | no |
| **신설** | **보조 실행자 이름이 감독 루프 이름과 겹치면 거부한다** | **a098 R — §3에 등록** | **예정 — 같음** | no |
| **신설** | **`Run`이 nil인 보조 실행자를 거부한다** | **a098 R — §3에 등록** | **예정 — 같음** | no |

> **⚠ 신설 셋의 RED는 「컴파일 실패」다.** 필드가 없으므로 테스트가 안 빌드된다.
> 그것은 **요구가 틀렸다는 빨강이 아니라 코드가 없다는 빨강**이다.
> born-GREEN을 피하려면 **필드만 먼저 더하고 검증은 안 넣은 상태**에서
> 세 케이스가 `ErrRuntimeUnavailable`을 못 받는 것을 보고, 그다음 검증을 넣는다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

B1~B9와 `:231` — 열 자리 전부 기존 테스트가 덮고 **a098은 그 조건을 안 바꾼다.**
§5의 VERIFY는 그 열 자리의 **판정이 안 바뀌었다**만 확인한다.

## 덮이지 않은 것을 이름으로 적는다

- **B10·B11·B12의 참 쪽**(0값 → 기본값 보정) — 기존 테스트가 전부 명시값을 넘기므로
  **기본값이 실제로 채워지는 경로를 아무도 단언하지 않는다.**
  a098은 그 셋을 **안 건드리므로 이 change의 RED 대상이 아니다**(`not-applicable`).
  다만 회귀가 나면 여기를 본다.
- **`Recover`·`Alerts`·`Escalate`·`Announcer`·`Log`의 nil 허용** — 검증 자체가 없으므로
  분기가 없고 테스트도 없다. a098은 그 정책을 **안 바꾼다.**
