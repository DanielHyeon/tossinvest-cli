# Branch Test Map: `NewRuntime`

**GREEN 칸은 호출 지점을 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이다: **편집 후** AST 분기 17 · 이탈 10 (편집 전 12 · 7).

> **⛔ 아래 표의 B10·B11·B12 는 편집 후 B15·B16·B17 이다.** 신설 검증을 가운데
> 넣어 뒤의 id 가 밀렸다 — **판정은 한 자도 안 바뀌었고 번호만 밀렸다.**
> 표를 번호로 읽으면 「B10 = 시계 기본값」이 **신설 검증**을 가리키게 된다.
> 신설 셋의 실측 자리는 **B10~B14**(`range`·`switch`·`case`×3)이다.

기존 커버리지의 정본은 `TestTheRuntimeRefusesWiringItCannotSupervise`
(`internal/app/engine/runtime_test.go:450`)의 일곱 케이스 표다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:192` 루프가 하나도 없으면 거부 | `TestTheRuntimeRefusesWiringItCannotSupervise` | no | **yes (기존)** |
| B2 | `:197` 루프 집합을 훑는다 | 같은 테스트 | no | **yes (기존)** |
| B3 | `:199` `switch` — 조건 없는 분기 | 같은 테스트 | no | **yes (기존)** |
| B4 | `:200` 이름 없는 루프를 거부 | 같은 테스트 | no | **yes (기존)** |
| B5 | `:203` 이름이 겹치는 루프를 거부 | 같은 테스트 | no | **yes (기존)** |
| B6 | `:205` `Run`이 nil인 루프를 거부 | 같은 테스트 | no | **yes (기존)** |
| B7 | `:209` `Health`가 nil이면 트리거·계좌를 안 본다 | `TestAGracefulCancelStopsEveryLoopAndRaisesNoCritical` (`runtime_test.go:127`) — `Health` 없는 루프로 조립한다 | no | **yes (기존)** |
| B8 | `:212` 열거에 없는 트리거를 거부 | `TestTheRuntimeRefusesWiringItCannotSupervise` | no | **yes (기존)** |
| B9 | `:217` 승격 가능한 루프에 계좌가 없으면 거부 | 같은 테스트 | no | **yes (기존)** |
| B15 | `:249` 시계가 nil이면 시스템 시계 (편집 전 B10) | `TestFiveConsecutiveFailuresEscalateOnceAndTheLoopKeepsRunning` (`runtime_test.go:276`)는 **가짜 시계를 넘긴다** — 즉 이 분기의 **거짓 쪽**만 덮인다 | no | **부분** |
| B16 | `:252` 임계 0이면 기본값 (편집 전 B11) | 같은 테스트가 명시 임계를 넘긴다 — **거짓 쪽만** | no | **부분** |
| B17 | `:255` 간격 0이면 기본값 (편집 전 B12) | `TestTheSupervisorPollsHealthWhileTheLoopsRun` (`runtime_test.go:405`)도 명시 간격을 넘긴다 — **거짓 쪽만** | no | **부분** |
| 이탈 `:260` | 정상 조립 | 위 전부 | no | **yes (기존)** |
| B10 | **신설** `:227` — `range opts.Auxiliary` (편집 전 B10 은 시계 기본값이었다) | 같은 테스트 전체 | **yes — 뮤테이션 F** | **yes (2026-08-12)** |
| B11 | **신설** `:229` — `switch` (조건 없음) | — | — | — |
| B12 | **신설** `:230` — **이름 없는 보조 실행자를 거부한다** | `TestTheRuntimeRefusesAMisWiredAuxiliaryExecutor/이름이_없다` | **yes — 뮤테이션 F** (아래) | **yes (2026-08-12)** |
| B13 | **신설** `:233` — **보조끼리, 그리고 감독 루프와도 이름이 겹치면 거부한다** | 같은 테스트의 하위 둘 | **yes — 뮤테이션 F · E** | **yes** |
| B14 | **신설** `:236` — **`Run`이 nil인 보조 실행자를 거부한다** | `…/Run_이_없다` | **yes — 뮤테이션 F** | **yes** |

> **⚠ 신설 셋의 RED는 「컴파일 실패」다.** 필드가 없으므로 테스트가 안 빌드된다.
> 그것은 **요구가 틀렸다는 빨강이 아니라 코드가 없다는 빨강**이다.
> born-GREEN을 피하려면 **필드만 먼저 더하고 검증은 안 넣은 상태**에서
> 세 케이스가 `ErrRuntimeUnavailable`을 못 받는 것을 보고, 그다음 검증을 넣는다.
>
> ## ✅ 그 상태를 실제로 만들어서 봤다 — 뮤테이션 F (2026-08-12)
>
> 필드와 검증을 **한 커밋에 같이** 넣었으므로 위 문단이 요구한 중간 상태가 시간축에
> 없었다. 그래서 **검증 블록만 도로 지워** 같은 상태를 만들었다: `Auxiliary` 필드는
> 있고 `NewRuntime`은 그것을 안 본다.
>
> **네 하위 케이스가 전부 FAIL**했다 — `NewRuntime accepted the wiring`.
> 그리고 뮤테이션 **E**(`seen`을 새 맵으로)는 **하나만** FAIL 시킨다. 둘을 같이
> 해야 *"검증이 있다"*와 *"그 검증이 같은 이름 공간을 쓴다"*가 갈린다.
>
> **시간축의 중간 상태와 뮤테이션으로 만든 중간 상태는 같은 증거인가.** 관측되는
> 사실은 같고(검증이 없으면 넷이 빨갛다), 다른 것은 **누가 그 상태를 만들었는가**다.
> 뒤엣것은 **재현 가능**하다는 점에서 오히려 낫다 — 이 문단을 읽는 사람이 같은
> 한 줄을 지워 같은 빨강을 다시 볼 수 있다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

B1~B9와 `:260` — 열 자리 전부 기존 테스트가 덮고 **a098은 그 조건을 안 바꾼다.**
§5의 VERIFY는 그 열 자리의 **판정이 안 바뀌었다**만 확인한다.

## 덮이지 않은 것을 이름으로 적는다

- **B15·B16·B17의 참 쪽**(편집 전 B10~B12)(0값 → 기본값 보정) — 기존 테스트가 전부 명시값을 넘기므로
  **기본값이 실제로 채워지는 경로를 아무도 단언하지 않는다.**
  a098은 그 셋을 **안 건드리므로 이 change의 RED 대상이 아니다**(`not-applicable`).
  다만 회귀가 나면 여기를 본다.
- **`Recover`·`Alerts`·`Escalate`·`Announcer`·`Log`의 nil 허용** — 검증 자체가 없으므로
  분기가 없고 테스트도 없다. a098은 그 정책을 **안 바꾼다.**
