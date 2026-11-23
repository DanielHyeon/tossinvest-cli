# 1.5 — `AutomationStatus` 필드를 읽는 자리 전수 (타입 기반)

- 방법: `census.go.txt`(go/types — 선택자가 가리키는 객체가 `engine.AutomationStatus` 의 필드인 자리만, 비시험 Go 전체 `./...` 133 패키지 타입 검사).
  grep 은 `.Verified` 를 이름으로만 찾아 다른 타입(전략 activation 의 `Verified()` · verifylive)을 섞고 다른 변수 이름(`status` · `automation`)을 놓침.
- 원문: `sites.txt`(55줄 — read · write · literal · status-value). HEAD `e4165fa3d`(코드 = `9738a0c32`).
- `AutomationStatus` 값 자체가 복사·투영되는 자리(`status-value`)도 셈 — `Context.Automation` 밖으로 나가는 구조체 복사는 0, 필드 투영은
  `StrategyEntryProductionSnapshot`(아래 ⑦) 하나. console · httpapi · cmd 의 다른 파일은 `AutomationStatus` 를 읽지 않음(0 자리).

## `Verified` — 읽는 자리 8 (쓰는 자리 1: interlock.go:406 `runInterlock`)

| # | 자리 | 함수 | 지금의 뜻 | 분류 | 제안 |
|---|---|---|---|---|---|
| ① | cmd/tossctl/engine.go:232 | `runEngineRun` B7 | 검증 아니면 `errEngineGateOff` — 런타임 조립 여부 | **청산 쪽** | 새 표지로 전환(design D3) |
| ② | internal/app/engine/engine.go:592 | `NewContext` B16 | 검증 아니면 Guardian=nil | **청산 쪽**(exit 관측이 Guardian 의 `ReductionIssuer` 를 요구) | 새 표지로 전환(D3) |
| ③ | internal/app/engine/exitwiring.go:323 | `Context.ExitObserver` B2 | 검증 아니면 거절 | **청산 쪽** | 새 표지로 전환(D3) |
| ④ | internal/app/engine/reconcileloop.go:368 | `Context.ReconcileDriver` B2 | 검증 아니면 거절(대사 + 편입) | **청산 쪽** | 새 표지로 전환(D3) |
| ⑤ | internal/app/engine/strategy_entry_supervisor.go:469 | `Context.productionStrategyWorker` | `Journal·Gateway·Guardian != nil && Verified` 아니면 휴면 worker | **진입 쪽** | **엄격 유지**(D3 가 든 그 자리) — 청산 전용에서 Guardian 은 비-nil 이 되므로 이 조건의 마지막 항(`Verified`)이 유일한 차단이 됨 → 시험으로 고정 권고 |
| ⑥ | internal/app/engine/interlock.go:461 | `logGateDecision` | 구조 로그 필드 `gate_verified` | **보고** | 그대로(false 가 참) + 새 표지 필드 하나 더함 권고 |
| ⑦ | internal/app/engine/interlock.go:479 | `logGateDecision` | `Verified && Protection != WIRED` 면 「보호는 프로세스 수명」 문장 | **보고** — 단 정본 「보호 미배선 기동」(운영자에게 기동 시점에 전달 SHALL)이 걸림 | 청산 전용 기동도 보호 미배선 런타임이므로 **문장을 내야 함** → 조건을 `(Verified ‖ 청산 전용)` 으로. design 의 넷 밖 — Manager Q |
| ⑧ | internal/app/engine/strategy_entry_supervisor.go:392 | `Context.NewPairedStrategyEntryProductionAssembly` | `StrategyEntryProductionSnapshot.AutomationVerified` 투영 | **보고**(투영 — 이 필드의 읽는 자리 0, 스냅숏은 `Schedule` 만 쓰임 grep) | 그대로(엄격 쪽) |

**판정:** design D3 의 넷(①~④)은 청산 쪽 자리 **전부**다. 넷 밖의 넷(⑤~⑧)은 진입 1 · 보고 3 이고 「판단 불가」 0 — 멈춤 조건 아님.
⑤ 는 design 이 이미 엄격 유지로 지목한 자리, ⑦ 은 설계에 없던 보고 자리로 결정이 필요(Q).

## `EntryPermitted` — 읽는 자리 2 (쓰는 자리 1: interlock.go:409 `runInterlock`)

| 자리 | 함수 | 분류 | 제안 |
|---|---|---|---|
| interlock.go:464 | `logGateDecision` 필드 `entry_permitted` | 보고 | 그대로(청산 전용에서 false — D5) |
| strategy_entry_supervisor.go:393 | 스냅숏 투영 `EntryPermitted` | 보고(투영, 읽는 자리 0) | 그대로 |

## `Verified` 를 읽지 않고 진입을 여는 경로가 있는가 (claim-check (b) 의 근거)

- 진입(노출 증가)의 생산자는 게이트웨이 `Place` · `Amend` · `PlaceClaimedStrategy`(→ `place` → `submit`) 뿐이고, 셋 다 `submit` 의 `checkProtection`(2a)·`checkEntry`(2b)
  를 지난다(gateway.go:586·592, 디스패치 직전 재검사 662·669). `checkEntry` 는 `g.entry == nil || !plan.raisesExposure` 면 통과(AST B1@860) — 청산 전용 기동도
  조항 8(게이트웨이 배선 — `Wiring().Entry`)을 통과해야 하므로(D2) `g.entry` 는 비-nil.
- `ReplayInDoubt`(재전송)는 `checkEntry` 를 안 지나지만 생산 배선이 `Attested: nil`(internal/app/engine/gateway.go:338) → `ReasonReplayNotAttested` 로 먼저 반환(replay.go:260). 지금 빌드에서 경로 없음.
- Guardian 이 비-nil 이 되어 새로 움직이는 것: `ManualBuyRefresher`(manual_buy_refresh.go:65 — 읽기만, 일일 매수 한도 보드 갱신), 전략 조립의
  `guardian, _ := c.Guardian.(*execgw.RiskGuardian)`(strategy_entry_supervisor.go:365 — first-leg bridge 재료; 발급은 worker 의 사이클에서만이고 worker 는 ⑤ 로 휴면).
- 결론: `Verified` 를 읽지 않는 진입 경로는 전부 게이트웨이 한 곳을 지난다 → 그 게이트웨이의 `EntryGate` 에 래치가 서 있는 한 fail-closed. 래치 세우기 누락이 유일한 구멍
  이므로 변이 「래치 세우기 삭제」(design §3-10)가 핵심 변이다.
