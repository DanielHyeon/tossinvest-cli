# Voice B — 증거 리뷰 (5.6.2.1 · 5.2.2.1, 대상 00e1b9bd)

> 기록 경위: 리뷰어의 /tmp 파일 쓰기가 권한 시스템에 거부되어(리뷰어는 우회하지 않음),
> Manager 가 hand-back 으로 수신한 원문을 이 정본 위치에 그대로 옮김 (2026-09-30).

**판정: BLOCK** (증거 수준의 P1 셋. 생산 동작 변화나 안전 불변식 위반은 찾지 못했음. 오늘 생산에는 서명 매니페스트가 0건이라 모든 시장이 `dispatchHandoffs` B1(시장 단위) 갈래를 탐.)

사본: `/tmp/claude-1000/a112-rev-2j4N`(리뷰 트리 = 00e1b9bd 전체), 편집 전 사본 `/tmp/claude-1000/a112-rev-pre-X753`, 변이 작업 폴더 `/tmp/claude-1000/a112-rev-2j4N-mut`. 셋 다 삭제했음. 실제 저장소에는 쓰지 않았음. `git diff --stat` 에 보이는 3파일은 병행 세션(a094 등)의 미커밋 편집이고 내 것이 아님. 하네스는 git 없는 사본에서 돌도록 `ROOT`/archive 부분만 바꾼 것을 scratchpad 에서 썼음(`lot_mutate_local.py`, `extra_mutate.py`). 모든 판에서 무변이 대조군 GREEN 을 먼저 확인했고, 원복은 sha256 으로 확인(restored=True).

## 발견

| # | 등급 | 주장 | 증거 | 권고 |
|---|---|---|---|---|
| 1 | **P1** | `ownerScope` 에 **금지 축**(Horizon)이나 전략군 축(LaneID)을 넣어도 스위트 전체가 통과함. `handoff.go` 주석은 "`Horizon` 은 절대 넣지 않음"(넣으면 같은 종목의 두 전략군이 둘 다 건너감)이라고 하지만, 그것을 못 박는 시험이 없음. 이 방향은 비보수(범위당 상한이 뚫림)임. | X03 `symbol: …+"\|"+string(result.Lineage.Horizon)` → 필터 세트 SURVIVED. 전체 태그 스위트(`go test -tags tossos_testseams ./internal/app/engine ./internal/strategyhandoff ./internal/strategyworker`, 대조군 GREEN)에서도 **SURVIVED**. X04(LaneID)도 필터 세트 SURVIVED. 원인: `owner_scope_test.go` 의 duplicate 경우는 두 선택의 Horizon·LaneID 가 같음(영값). | duplicate 경우에 "같은 계좌·시장·종목·세대 + **다른 Horizon/LaneID** → OverCapacity" 를 더할 것. |
| 2 | **P1** | 구조 못 `TestTheProductionCycleDeliversEveryOwnerScopeHandoff` 가 우회됨. 이 못이 `runProductionStrategyMarketCycle` 의 유일한 검출기인데(진입 0), 세 가지 우회가 모두 살아남음. | 전체 태그 스위트 **SURVIVED**: X08 `marketWideOnly{…}.dispatchHandoffs()`(이름만 같은 메서드가 시장 단위 `[]{m.a.dispatchHandoff()}` 를 돌려줌. selector 이름만 봄) · X09 몸통 closure 안 `if calls++; calls > 1 { return nil }`(둘째 범위부터 조용히 버림. Args[0] 만 봄) · X10 `deliverEachStrategyHandoff := deliverOnlyFirstStrategyHandoff`(지역 섀도. Ident 이름만 봄). 반대로 동작이 같은 리팩터 X11(`handoffs := …dispatchHandoffs()` 변수 경유)은 CAUGHT(:222) → 못이 거짓 양성에도 약함. | 못은 식별자 해소(`types.Info.Uses` → 패키지 수준 `deliverEachStrategyHandoff` 함수 객체 · 수신자 타입이 `strategyProposalMarketAuthority` 인 `dispatchHandoffs`)로 세고, 몸통 closure 가 캡처 변수에 쓰는 것을 금지할 것. 근본적으로는 이 함수를 도는 행동 시험(이월 표의 5.6.2.2)이 필요함. |
| 3 | **P1** | 5.6.2.1 의 "변이 E01~E11 **11/11 CAUGHT**" 는 착지 커밋에서 재현되지 않고, 원장도 커밋돼 있지 않음. | 착지 트리에서 재실행 결과 **E10 `NOT-APPLIED old occurs 0 times`**. 앵커 `"\t\tReasonStrategyCentralIntegrity,\n\t}"` 는 766a8456(a094, 20:10)이 뒤에 `ReasonOppositePendingOrder` 를 덧붙이면서 사라졌음. `git show <c>:internal/execgw/failclosed.go` 로 보면 3260f4eb 에서는 `}` 가 바로 뒤, **36ade9b2 · 00e1b9bd 에서는 빈 줄 + a094 줄**임. 36ade9b2 판 하네스에도 같은 앵커가 있음. `lot-5.6.2-5.2.2/` 에 `mutation-5.6.2.1.tsv` 가 없어서 어느 HEAD 에서 쟀는지 확인할 수 없음. 앵커를 고친 X01 은 CAUGHT(`TestReasonCodeEnumIsStable` failclosed_test.go:48) → 성질 자체는 성립함. | 앵커를 `"\t\tReasonStrategyCentralIntegrity,\n\n"` 으로 고치고, 착지 트리에서 다시 돌려 원장(TREE · CONTROL 줄 포함)을 커밋할 것. |
| 4 | P2 | 오늘-동등성 핀의 **대조 절반**을 유일하게 잡는 변이가 원장에 없음(판별력 미증명). | 원장 F01~F15 중 대조 절반(:132/:141)으로 죽은 것은 0. 내가 만든 X12(1차 레그 관문 `len(proposal.entries) >= 1`)는 CAUGHT이고, `TestTwoOwnerScopes…` 의 **대조 줄 :132 에서만** 실패("control: a single-scope market was refused by the same sentence") → 판별력은 있음. | X12 류 변이를 원장에 추가할 것. |
| 5 | P2 | `TestAScopeFaultStopsTheCycleBeforeTheNextScope` 의 둘째 절반(handoff 하나일 때 감싸지 않은 같은 값)을 유일하게 잡는 변이가 없음. | X13(`fmt.Errorf("scope: %w", err)`)은 첫 절반 :178 과 closure 시험(fmt import)이 잡음. 둘째 절반 :183 만 실패시키는 변이는 원장에 없음. | 선택 사항. |
| 6 | (T) | 증명 안 된 주석이 둘: `deliverEachStrategyHandoff` 의 "거절된 handoff 는 … 다음 범위로 넘어감", `ownerScope` 의 market 축. 둘 다 현재 생산자 아래에서는 동등 변이임. | X05(거절에서 멈춤) SURVIVED: `AdmitEachOwnerScope` 는 [거절 하나] 아니면 전부 Admitted 만 내므로 섞인 목록에 도달할 수 없음. X02(market 축 제거) SURVIVED: 엔진은 한 시장 권한의 항목만 넘기고, 축을 빼는 것은 보수 방향임. | 주석에 "도달 불가 · 미시험" 을 적거나 단위 시험 하나를 둘 것. |
| 7 | (T) | 편집 번들 `runproductionstrategymarketcycle` 의 FLM·BTM 머리 `File SHA-256: 22855de0…` 가 틀림. 번들은 "Pinned revision current — this worktree's file" 라고 적음. | 실제 해시: 00e1b9bd 파일 `1f4f2096…`(ast.json 과 일치), 36ade9b2 `da4fa6d1…`, 3260f4eb `9d97e59c…`. `22855de0` 은 어느 커밋과도 맞지 않음(4f49a8eb/44d0fb58 에서 유입). 편집 전 번들(`pre-edit-5.2.2.1/`) md 도 같은 값. 산문의 `455:12` · `479:15` · `502:12` · `456:3` · `451:16` 도 옛 좌표(기존 부채). | 머리 SHA 와 산문 좌표를 정정할 것. |
| 8 | (T) | 이동 스크립트가 **기존 오염**을 계속 옮김. `invokeboundedstrategycycle` BTM 의 시험 옆 괄호가 원래 갈래 줄 `(889-890)`/`(891-892)` 이었는데, 44d0fb58 에서 함수 범위로 오염됐음. `text.replace(rng,new_rng)` 가 이번에도 `(1047-1067)→(1051-1071)` 로 옮김. `newstrategyentrysupervisor` BTM 은 분기 좌표 +4(625:2)와 측정 블록 좌표(621.52-623.3, 36ade9b2 소스 기준) 사이에 설명 없이 한 행에 두 기준이 섞임. | `git log` 추적(472ce405…00e1b9bd). | 기준을 표기하거나 측정 블록을 재측정할 것. |

## 필수 항목별 판정

1. **RED 를 거치지 않은 시험들의 판별력**
   - **(a) 원장(F01~F15)이 유일하게 잡는 변이가 있는가**
     - 반복 헬퍼 둘: F09 · F10 이 각각 유일함 ✓
     - 구조 못: F11 유일 ✓. 다만 #2 우회가 살아남음.
     - 표기 정규화 둘: F06 · F07 유일 ✓
     - 활성화됐지만 닫힌 시장: F13 유일 ✓(scoped 단언은 Ready=true 라 닫힘 단언만 실패함)
     - 오늘-동등성 핀 대조: ✗ 원장에 없음(#4, X12 로 내가 보임).
   - **(b) 사후 RED** — 편집 전 소스(36ade9b2 의 handoff.go · strategy_dispatch_handoff.go · strategy_entry_supervisor.go)에 RED 로그와 같은 스캐폴드를 얹어 돌림.
     - 빨간 것: 구조 못(:222 "…=0, …mentions=2"), census(:657), `TestEveryAdmitted…`.
     - 초록인 것(편집을 증명하지 않음): `TestMarketLevelRefusalsStayOneHandoff`(정규화 둘 포함 — 편집 전 `Admit` 도 2개면 OverCapacity), 헬퍼만 편집 전 동등물(`handoffs[0].Deliver`)로 둔 격리판의 `TestAScopeFaultStops…`(delivered=[a], err=fault 로 PASS), 닫힌-시장 단언(편집 전 `dispatchHandoff()` → MarketClosed 로 통과, 코드로 추론). 이들의 판별력은 전적으로 F06 · F07 · F10 · F13 에 기댐. 그 변이들이 원장에 있으므로 수용 가능함.
   - **(c) 내가 만든 변이**
     - SURVIVED: X03(Horizon) · X08 · X09 · X10(못 우회). 넷 다 전체 태그 스위트에서 확인 → **P1**(#1, #2).
     - X04(LaneID) SURVIVED(필터 세트) → #1 에 포함.
     - X02 · X05 SURVIVED: 동등 변이 → (T)(#6).
     - X06(헬퍼 역순) · X07(`AdmitEachOwnerScope` 역순) · X11 · X12 · X13 CAUGHT.
2. **하네스 유효성**
   - 5.2.2.1 세트: 사본은 archive 방식으로 병행 세션의 미커밋 편집을 막음 ✓. 적용 여부는 `count(old)==1` 로 확인하고 NOT-APPLIED 를 기록함 ✓. 원장 TREE 는 b74875e7 이지만, 00e1b9bd 사본에서 다시 돌리니 **15/15 가 똑같이 재현**됐고 대조군 GREEN ✓.
   - `-run` 필터: 5.2.2.1 의 새 시험 전부와 `Handoff|Seam` 을 포함함. 필터 밖 시험이 X03 · X08~X10 을 대신 잡지도 않음(전체 스위트로 확인).
   - 5.6.2.1 세트: 재실행 결과 E01~E09 · E11 CAUGHT, **E10 NOT-APPLIED** → #3.
   - 약점: 원복 검증이 없음(사본을 버리므로 영향 작음). 이름 없는 실패(panic · timeout)도 CAUGHT 로 셈.
3. **RED 로그의 진정성**
   - `red-5.2.2.1.log`: 실패 이유가 "handoffs=1, want one per owner scope" 로 기능 부재 ✓. 조립 · 컴파일 오류 아님. 스캐폴드는 머리말에 적혀 있음. 로그의 줄 번호(:63/:87/:70)가 지금(:65/:96/:74)과 다른 것은 GREEN 뒤 시험을 더했다는 서술과 일치함.
   - `red-5.6.2.1.log`: 넷 다 기능 부재(blocker <nil> · err <nil> · blocks=map[] · swallowed) ✓. 다만 명령 · 트리 · 스캐폴드(EntryGate 옵션 존재) 머리말이 없어 출처가 약함 (T).
4. **FLM/BTM**
   - `go run ./tools/logic-map --func Context.runProductionStrategyMarketCycle` 결과가 번들 ast.json 과 **완전히 같음**. 같은 파일의 12개 번들 전부 재추출 결과도 SAME ✓.
   - 좌표 · 호출 표 · return 목록 ✓. 이동 번들 표본(refreshPaired… · runMarket · invokeBounded… · newStrategyEntrySupervisor)은 자기 AST 좌표만 +4 로 옮겼고, 이번 로트가 새로 잘못 옮긴 것은 없음. 기존 오염의 전파와 머리 SHA 는 #7, #8.
   - "진입 0": 태그 스위트 커버리지(`-coverpkg=./internal/app/engine`)를 재측정한 결과 496–565 의 모든 블록이 count 0 ✓(trimpath 비호환 a111 시험 둘은 로트와 무관한 실패). 새 헬퍼 `dispatchHandoffs` · `deliverEachStrategyHandoff` 는 시험에서만 진입함.

Recommendation: 착지 상태로 두지 말고 작은 증거 수리 로트를 낼 것 — (1) Horizon/LaneID 가 다른 같은 범위 → OverCapacity 시험을 추가하고, (2) 구조 못을 타입 해소 기반으로 바꾸고 closure 캡처 쓰기를 금지하며, (3) E10 앵커를 고쳐 착지 트리에서 E 원장을 커밋하고 X12 를 5.2.2.1 원장에 더할 것. 이유: 생산 동작은 오늘 동등하지만, 주석이 금지한 축과 "모든 범위를 건넨다" 는 주장을 지키는 핀이 뚫려 있고, 11/11 이라는 주장은 그것을 주장하는 커밋에서 재현되지 않음.
