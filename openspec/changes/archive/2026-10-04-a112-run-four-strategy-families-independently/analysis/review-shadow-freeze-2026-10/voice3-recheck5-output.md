# SHADOW re-freeze 재검 5라운드 voice3 — SubagentHandback 원문 (agent ac4fbb2d05a76746e, 저장 2026-10-05)

**5라운드(보이스 ③, 좌표 4d22d726) 결과: PASS.** 내 남은 표적 셋(P1-C, P1-D, 좌표 정정)은 모두 CLOSED이고, v3.2가 새로 연 P0/P1은 없다. 다만 codex·보이스 1이 4라운드에 낸 「record 전 실패 뒤 이전 SHADOW 잔존」 P1을 v3.2의 폐기가 닫는지는 **부분 종결**로 본다. 남은 갈래 하나를 아래 「잔여 노트」에 적었고, 등급은 Manager가 판단해야 한다.

**안전 규칙:** `~/.codex`는 읽지도 검색하지도 않았다. 리뷰 디렉터리의 `_codex/`도 열지 않았다. 저장소 파일은 편집하지 않았다. 모든 셸 첫 줄은 `set -euo pipefail`로 시작했다. 기존 사본 `…/scratchpad/v3r3`(archive 4d22d726)에서 판독만 했고, 사본 쓰기나 시험 실행은 하지 않았다.

**입력 확인:** 브리프 v3.2 sha256 `68512bc36b58…09af1` 일치.

| 항목 | 판정 | v3.2 좌표 | 근거 |
|---|---|---|---|
| P1-C (활성화 출처) | **CLOSED** | §5 칸 `{wave, batch, activation}`을 한 임계 구역에서 함께 씀 · §2 ③ 금지 집합 | activation은 evaluate가 이미 받는 `promotion`이다(strategy_lane_runtime.go:199-201 일치). shadow 단계의 활성화는 칸의 복사본뿐이다. 금지 집합 좌표 `familyGateFor` :115, `loadFamilyActivation` :154(strategy_family_activation.go)는 모두 맞다. 활성화 축에서 「같은 파도」가 구성으로 성립한다. |
| P1-D (`forMarket` 본문이 주문 경로 앞에서 핀 밖) | **CLOSED** | §5 핀 ① 타입 규칙 · `forMarket` shape 핀 · 인자 index 5 | shadow 타입 식은 evaluate의 `Args[5]` 부분 트리 하나로 한정되고, 그 안의 유일한 호출이 `forMarket`이다. `forMarket` 본문은 형제 함수(strategy_proposal_authority.go:167-175, 확인함)와 같은 모양(시장 비교 후 필드 반환, 호출 0)으로 AST에 고정된다. evaluate의 현재 인자는 5개(index 0~4)이므로 새 인자는 마지막인 index 5가 되고, 잠금 복구 세대 핀 `Args[2]`(a112_lane_latch_durability_test.go:335-348)는 그대로 보존된다. 4라운드 문면의 「호출 0」과의 자기모순도 해소됐다. |
| 좌표 정정 | **CLOSED** | §0 · 머리말 | `gate.admit(`는 strategy_market_coordinator.go:91, 머리말은 4d22d726으로 맞다. |

**v3.2가 새로 인용한 좌표 실측:**
- 모두 맞음: strategy_lane_runtime.go :199-201(evaluate 서명) · :208-209(`recoverMarketLanes` 오류 반환) · :244-247(panic 재던짐) · strategy_entry_supervisor.go:1079(`invokeStrategyCycle`, recover는 :1080-1088) · strategy_proposal_authority.go:270-276(collect의 recover) · 상수 `DefaultStrategyCycleLimit = 5s`, `MaximumStrategyCycleLimit = 30s`(strategy_entry_supervisor.go:29-30).
- :1035는 `evaluationState`(시작 :1032) 안의 `worker.latched` 조건 줄이다. 인용으로 허용 범위다.

**사실 하나 정정(비등급):**
- v3.2 §5는 「panic은 시장을 잠가(latchMarket) 주기 자체가 멈춘다」고 쓴다. 이것은 **effective worker에만** 맞다.
- 오늘 생산의 dormant 갱신 worker(`refreshOnly = !effective && RefreshesAuthority`, :899-901)는 오류와 panic(abnormal 오류)을 `recordSwallowedCycleError` 후 `continue`로 삼킨다(:922-951). 잠그지 않고 다음 폴에서 계속 돈다.
- defer 폐기는 두 경우 모두 작동하므로 설계 결론은 바뀌지 않는다. 문장만 고치면 된다.

**공통 질문:**
- **defer와 `invalidateShadow`가 주기 경로를 바꾸는가:** 바꾸지 않는다.
  - recover하지 않는 defer는 panic이 펼쳐지는 동안 실행되고, panic은 그대로 `invokeStrategyCycle`의 recover로 간다. 반환값도 그대로 전달된다. 단, 구현 시 「nil 반환 여부」는 결과값을 받아 둔 플래그로 판정해야 한다. `recover()`로 판정하면 안 된다.
  - 펼침 중 `runtime.mu` 교착 가능성도 확인했다. defer 없이 잠그는 구간은 strategy_lane_latch.go :156-169, :243-245, :259-261이다. 이 구간에는 map 순회·append·대입만 있고, `latches` map은 생성자에서 초기화된다. 그래서 잠금을 쥔 채 panic이 날 거리가 없고, `invalidateShadow`가 잠금을 기다리며 멈출 위험도 사실상 0이다.
- **CAS 잠금 순서:** `invalidateShadow`와 게시 CAS는 `runtime.mu`만 잡는다(epoch와 wave 비교). 레인 잠금은 잡지 않는다. projection의 「런타임 → 레인」 순서와 역전이 없다. 단 CAS 본문이 레인 접근자를 부르지 않도록 `invalidateShadow`와 같은 shape 핀을 두기를 권한다.
- **나이 상한 유도:** 2×(5s + 30s + 2s) = 74s는 세 상수에서 나오는 식이다. 30s를 다 쓰는 건강한 주기에서 「거부 안 함」 대조와 경계 핀(정확히 MaxAge / MaxAge−1ns)도 있다. 성립한다.

**잔여 노트 (codex·보이스 1 P1의 종결 판단용, 내 등급은 P2):** supervisor 마감(30s)으로 버려진 주기는 v3.2 폐기의 범위 밖이다.
- 무슨 일이 일어나는가:
  - `invokeBoundedStrategyCycle`는 마감이 지나면 `ErrStrategyCycleDeadline`(abnormal)를 돌려주고 그 주기를 버린다(:1057-1076). 이 감시자 ctx는 주기 goroutine의 ctx를 **취소하지 않는다**(`watchdogCtx`는 Sleep 전용).
  - 그래서 버려진 주기 goroutine은 계속 돌다가 늦게 **nil**로 돌아올 수 있다. 그러면 클로저의 defer는 폐기하지 않는다.
  - 오히려 shadow 단계가 시작되어 CAS를 통과한다. epoch는 그대로이고, wave는 클로저가 그 순간 칸에서 복사한 값이기 때문이다.
- 결과:
  - effective worker라면 이미 `latchMarket`으로 잠긴 시장에 새 SHADOW가 게시되고 최대 74s 동안 보인다.
  - refreshOnly worker라면 다음 주기와 겹쳐 돌 수 있다. 다만 칸의 묶음과 파도가 짝으로 복사되므로 정합성은 유지된다.
- 왜 P2로 보는가: 노출은 0이고, 가시성 문제이며, v3.2 자신의 나이 상한(74s)으로 시간이 묶인다.
- 닫는 법(한 줄): 클로저가 주기 시작부터 경과 시간이 `MaximumStrategyCycleLimit` 이상이면 실패로 취급한다(폐기하고 shadow는 시작하지 않음). 핀 (iv)에 「마감+δ 뒤 늦은 nil」 사례를 추가한다.
- codex·보이스 1의 「가시성 P1」 기준을 적용하면 P1이 될 수 있으므로 Manager 판정에 넘긴다.

**저장소 무변경 확인:**
- 시작과 끝 모두 `rev-parse HEAD` = `4d22d72663dcf3a42be928b3ecf3db817ffc6ac5`.
- `status --short`도 같다: ` M docs/ROADMAP.md`, ` M …/tasks.md`, `?? .reticle-setup-crash.log`, `?? …/review-shadow-freeze-2026-10/`, `?? …/shadow-2026-10/`, `?? w4.log`.

