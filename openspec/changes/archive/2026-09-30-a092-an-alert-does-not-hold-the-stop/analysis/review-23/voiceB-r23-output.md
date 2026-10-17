<!-- 23라운드 보이스 B(산출물 무기화 렌즈) — Claude code-inspector-tester 서브에이전트(독립 컨텍스트, 같은 하네스). 프롬프트 prompt-common.md + 호출 렌즈.
     대상 a88e7079(HEAD 7002270f 와 internal/ cmd/ openspec/specs/ 바이트 동일). 원문 그대로. -->

**판정: BLOCK**

보이스 B(산출물 무기화 렌즈)입니다. 대상은 23판(`a88e7079`)입니다.
- HEAD `7002270f`와 `a88e7079`는 `internal/` `cmd/` `openspec/specs/` 범위에서 바이트 동일합니다.
- 작업 트리의 미커밋 a066 파일(`risk_relaxation_command.go` 등)은 쓰지 않았습니다. 코드는 모두 `git show a88e7079:` 사본으로 읽었습니다.
- MANIFEST의 파일별 sha256 13개가 `a88e7079`와 전부 일치합니다.
- `check_values.py` 실패 60건은 모두 23판 절(design 727-826, tasks 13-37, proposal 1-11) 밖입니다. spec 델타에는 0건입니다.

경로 약칭: `ES` = `openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md`, `EP` = 같은 change의 `specs/exit-policy/spec.md`.

## 발견

| # | 등급 | 주장 | 증거(파일:줄) | 권고 |
|---|---|---|---|---|
| 1 | **P1** | 23판이 「알림기를 거치는」을 「그 입구를 거치는」으로 바꾸면서 동기 `Notify` 경로가 규범 밖으로 빠졌습니다. 동기 경로는 `n.mu` 아래에서 `ClaimAlertForDelivery`를 부르고, 새 key를 넣으며, 재알림 창 1h로 정착 행을 재무장합니다. (가) 이 경로는 「나눌 수 없는 하나」 SHALL의 대상에서 빠집니다. (나) 문자대로 읽으면 「입구를 거치지 않고 원장에 쓰는 기록자」라서 먼저 잠그라는 SHALL이 걸립니다. (다) 첫 SHALL 「critical 기록의 정식 입구는 기록 전용 입구」는 동기 발송을 유지하는 범위 밖 호출자(`ES:45`)와 충돌합니다. (라) 같은 요구의 Scenario는 여전히 「알림기를 거치지 않는 기록자」라고 적어, 한 요구 안에서 부류 이름이 둘입니다. (마) K6 census가 `EnqueueAlert` · `RecordAlert`만 세고 `ClaimAlertForDelivery`는 세지 않습니다. `Acknowledge` 주석이 지키려는 재무장 경합(`notifier.go:836-839`)이 바로 이 경로입니다. | `ES:65` · `ES:67` · `ES:116`, 22판(`7cf80832`)의 「알림기를」 3곳, `notifier.go:251-262` · `:350-355`, tasks.md:28 | 부류를 「알림기 배제 잠금(`n.mu`) 아래에서 기록하는 모든 경로(기록 전용 입구 + 동기 claim)」로 정의합니다. census 핀에 `ClaimAlertForDelivery`(비시험 호출자 1개, `notifier.go:262`)를 더합니다. Scenario 문구를 통일합니다. |
| 2 | **P1** | ADDED의 새 무조건 SHALL 「운영 모드 전이의 통지는 상태를 바꾼 전이마다 한 번 나가야 한다」를 착지한 코드가 어깁니다. `CRITICAL_ALERT_UNDELIVERED` 승격은 announcer nil로 **통지하지 않습니다.** 동기 경로 `:228` → `escalate` `:382-383`이 그렇고, a124 배달 실행자 `alertdelivery.go:442-452`도 의도적으로 nil을 넘깁니다. 23판의 예외(`ES:77`)는 durable 기록 실패의 경우 하나뿐입니다. 이대로 archive하면 정본 SHALL을 HEAD 코드가 위반합니다. 23판 design K12는 같은 `:382-383`을 durable 실패 경로만의 근거로 인용합니다. | `ES:324`, `notifier.go:223-229` · `:360-368` · `:382-383`, `alertdelivery.go:447-452` | SHALL을 「announcer를 넘기는 전이」로 한정하거나, 전달 실패 승격(동기·실행자 양쪽)의 무통지를 근거(`notifier.go:362-368`)와 함께 예외로 적습니다. |
| 3 | **P1** | K2 수단을 적힌 대로 구현할 수 없습니다. design은 「`deliver`의 세 래치 자리가 `escalate`의 반환값으로 (ii)를 판정한다」고 적었지만, `escalate`는 `deliver`가 반환한 **뒤**, `claimAndDeliver` 밖에서 돕니다. 또 `:520` 자리의 판정은 `lost=true` → `owed=false`라 승격이 아예 없습니다. 이 판정은 a124 정본이 「승격을 포함하지 않는 판정(시도 기록이 행을 찾지 못함)은 승격을 만들어서는 안 된다」고 한 바로 그 경우입니다. 그런데 tasks 23.3 K2는 「세 래치 자리 각각」 RED를 요구합니다. `:520` RED는 만들 수 없거나, 만들려면 a124 SHALL NOT을 어기게 됩니다. 델타 「모든 발송자」는 a124의 두 조항만 옮겼고 이 SHALL NOT은 빠졌습니다. | design.md:755, `notifier.go:223-229` · `:309-316` · `:519-523`, 정본 engine-safety `:1460-1463`, tasks.md:24, `ES:69` | (ii)는 `notifyCritical`에서 `escalate` 실패 뒤 `Block`으로 판정합니다. RED는 `:484` · `:571`로 한정합니다. a124의 SHALL NOT을 이식합니다. `AccountRef` 빈값(`notifier.go:379`)을 「승격 미포함」으로 정의합니다. |
| 4 | **P1** | 착지한 a066 `notifyRelaxation`이 원장에 직접 쓰면서 먼저 잠그지 않습니다. 따라서 23판 SHALL(`ES:67`)과 K6·K7 핀(tasks 23.3)은 **HEAD에서 이미 빨갛습니다.** 완화 통지에는 먼저 세울 「자기 사유」도 없어서, 실현 가능한 길은 입구로 옮기는 것뿐입니다. 23.4는 「확인」일 뿐 착수·archive 선행 조건이 아닙니다. 그 파일은 지금 다른 세션이 미커밋으로 고치고 있습니다(git status `M internal/app/engine/risk_relaxation_command.go`). | `risk_relaxation_command.go:40` · `:151-173` · `:158`, `ES:67`, tasks.md:28-29 · `:37` | a066을 입구로 옮기는 것을 a092 착수(또는 archive) 조건으로 명시하거나, SHALL 범위를 조정합니다. |
| 5 | P2 (T) | K3 핀의 「Commit과 Project 사이에 반환 없음」은 HEAD 양성 대조군에서 실패합니다. commit 실패 갈래의 `return`이 `:469`에 있어 `:468`과 `:476` 사이에 놓입니다(AST returns). 투영기 내부의 `go`는 핀 범위 밖입니다. | head-ast-21 `…transitionoperatingmode.json` returns `469:3`, `operating_mode.go:468-477`, design.md:749, tasks.md:25 | 핀을 「커밋 성공 경로에서」로 한정하고, 투영기 몸체에 `go` 0 핀을 더합니다. |
| 6 | P2 (T) | 새 잔여 「프로세스 밖 기록자(flatten)」는 공집합입니다. `ReplayInDoubt`의 비시험 호출자는 `reconcile/recovery.go:351`(인터페이스) 하나입니다. `Replayer`는 `runtime_wiring.go:184`(`Context.Recovery`)에서만 대입되고, 그 호출자는 `cmd/tossctl/engine.go:645` 하나입니다. flatten CLI는 `Recovery`를 만들지 않습니다(`flatten.go:199-272`). `Saga.Notifier`도 nil입니다(`flatten.go:247-263`). | 위 좌표 | 잔여를 정적 census로 닫고 도달 집합 핀을 둡니다. 23.2의 FLM 항목은 핀으로 대체합니다. |
| 7 | P2 (T) | `ES:67`의 「진입 게이트를 넘기지 않는 조립은 생산 조립이 아니다」는 순환 정의입니다. 이 정의 아래에서 「생산 조립은 항상 넘긴다」는 정의상 참이 되어, 정본이 나중에 핀을 약화할 근거가 됩니다. | `ES:67` | 생산 조립을 「`tossctl` main에서 도달하는 비시험 경로」로 정의합니다. |
| 8 | P2 (T) | MODIFIED 「배달 실행자의 정지」에 낡은 좌표가 하나 남았습니다. `risk-management :102-108`(`ES:229`)이며, 정본 `:102-108`은 한도 완화 Scenario입니다. 23판은 「좌표 넷을 고쳤다」고 하지만 같은 인용 두 곳 가운데 하나(`ES:198`)만 고쳤습니다. 또 23판이 새로 넣은 HEAD 좌표(`engine.go :680`, `retry.go:532-539`)가 정본에 들어가는데, a092 자신의 편집으로 어긋날 것입니다(추정 — `EntryGate` 구조체가 `retry.go:461`에 있고 rowid 울타리 필드가 필요합니다). | `ES:175` · `:198` · `:213` · `:229`, 정본 risk-management `:101-104` · `:133` | 남은 하나를 고치고, 정본에 들어갈 블록에는 좌표 대신 이름을 씁니다. |
| 9 | P2 (T) | 정본 「엔진 런타임 수명주기」의 주석 「a092의 header note(`:24-25`)와 어긋난다 … 그 문장의 정리는 a092가 진다」가 archive 뒤에도 남습니다. K9는 같은 부류의 서술 둘만 고쳤고, a092는 그 요구를 건드리지 않는다고 선언합니다(`ES:28`). | 정본 engine-safety `:221-224`, `ES:28` | 잔여로 적거나 Manager에게 넘깁니다. |
| 10 | P2 (T) | K8 「풀 경로는 원장 수리 뒤 성공한 재시작」은 불완전합니다. 복원 실패 래치는 모드 사유 `ReasonOperatingModeBlocked`이고 울타리 초기값은 0이라, 뒤이은 성공 투영(자동 강화 → `mode-release`)이 그 래치를 교체·해제합니다. 사람 승인을 거치므로 안전 방향이지만 서술은 틀렸습니다. | `modegate.go:35-50`, design.md:791 · `:820` | 서술을 정정합니다. |
| 11 | P2 | 커버리지 빈칸이 셋입니다. K12(durable 실패 승격 무통지, 새 SHALL)의 RED가 없습니다. 입구의 `remindAfter=0`이 재무장하지 않는다는 RED가 없습니다. K13 구현은 `runAuxiliary`의 고정 이벤트 타입을 바꿔야 하는데, 그 FLM이 22.2·23.2 어느 목록에도 없습니다. | `ES:65` · `:77`, tasks.md:20-36 · `:48-50`, `auxiliary.go:99-109` | 23.3과 23.2에 추가합니다. |
| 12 | P2 (T) | exit-policy 정본 문장에 이관 **버퍼**라는 수단이 들어갑니다. engine-safety는 「유실 허용 형태」만 요구합니다(`ES:51`). 발행 실패로 버려지는 경우는 한정어가 덮지 않습니다. | `EP:30` · `:58` | 수단 이름을 빼고 「버려지면 기록된다」로 씁니다. |
| 13 | P2 (T) | K16 근거: 완화 통지를 동기 `Notify`로 보내면 전송 실패는 호출자에게 보이지 않습니다. 그래서 「통지 실패를 함께 보여야」(`ES:322`)는 전송 실패에 대해 만족할 수 없습니다. 드러나는 것은 durable 쓰기 실패뿐입니다. | `notifier.go:126-129` · `:130-139`, design.md:822 | 재읽기 결과로 보이는 사실을 「통지 실패」의 정의로 적습니다. |
| 14 | P2 (T) nit | (가) 22판 `engine.go:636`은 `c1e34dc4`에서도 이미 `:637`이었습니다. 23판 정정 `:639`는 맞지만 이유 서술이 부분적입니다. (나) `outbox.go:142-146`「(주석)」에서 `:146`은 코드 줄입니다. | `git show c1e34dc4:cmd/tossctl/engine.go` → 637, `outbox.go:142-146` | 서술만 고칩니다. |

## 필수 판정 1~8

1. **K1**: 성립합니다.
   - AST에서 B15 `:410` · B16 `:417`은 통지(`:478-479`) 전에 반환합니다.
   - 정본 「재알림 창 한 번」(`:708-710`)의 기준은 event key라서, 전이별 key와 모순되지 않습니다.
   - 강화 → 완화 → 재강화의 반복은 매번 사람 완화가 필요하므로 폭주하지 않습니다.
   - 다만 짝 SHALL의 과잉(발견 2)이 있습니다.
2. **K2 · K4**:
   - 문장은 a124와 같은 규범입니다. (i)은 정본 `:1471`, (ii)는 `:1463`과 같습니다.
   - 그러나 수단 배치와 `:520` RED가 틀렸고, 「승격 미포함 판정은 승격하지 않는다」가 빠졌습니다(발견 3).
   - 게이트를 잘못 여는 경로는 찾지 못했습니다. 승인 뒤의 승격은 투영 배선 아래에서 진입을 막습니다.
3. **K3**:
   - 해석 자체는 HEAD 순서(`:468` → `:475` → `:476` → `:478`), 정본 risk-management `:133`과 정합합니다.
   - 핀 문언은 HEAD `:469`에서 실패하고, 투영기 내부는 덮지 않습니다(발견 5).
   - engine-safety가 risk-management의 낱말을 해석하는 교차 권위라는 점은 기록해 둘 만합니다.
4. **K5**: 비시험 `execgw.New(` 호출자는 **정확히 2개**입니다. 핀 형태(전수 + 역할)는 적절합니다. 정의의 순환(발견 7)은 고쳐야 합니다.
   - `internal/app/engine/gateway.go:296` — `Entry: entry` `:302` ← `:249`
   - `cmd/tossctl/flatten.go:233` — `Entry: gate` `:239` ← `:232`
   - 둘 다 같은 함수 안의 `NewEntryGate` 결과입니다. `Gateway` 리터럴은 `New` 안(`gateway.go:161`) 하나이고, `entry`는 `:167`에서만 대입됩니다.
5. **K6 · K7**: 비시험 호출자 수는 아래와 같습니다.
   - `EnqueueAlert(` 호출 **2개**: `replay.go:551`(구체 `*journal.Journal`), `risk_relaxation_command.go:158`(인터페이스 경유, 선언 `:40`). 그 밖에 정의 `outbox.go:131`과 인터페이스 선언 1개가 있습니다.
   - `RecordAlert(` **0개**.
   - `ClaimAlertForDelivery` **1개**(`notifier.go:262`, `n.mu` `:254` 아래) — census에서 빠져 있습니다.
   - `INSERT INTO alert_outbox`는 `outbox.go:352` 하나입니다.
   - 재무장 SHALL을 exit 기록으로 한정한 것은 enqueue-only 계약(`outbox.go:142-146`, remindAfter 0)과 양립합니다.
   - 결론: 부류 정의(발견 1)와 a066 위반(발견 4)은 P1입니다.
6. **flatten 잔여**: 닿지 않습니다(발견 6). P0/P1이 아니고 정적 증명으로 닫을 수 있습니다.
7. **archive 정합성**: 정본 모순이 셋, 낡은 사실이 셋입니다.
   - 발견 1 · 2 · 4는 archive 시 정본에 모순이거나 HEAD가 이미 위반하는 SHALL을 넣습니다.
   - 발견 8 · 9 · 12는 낡거나 낡을 HEAD 사실을 넣습니다.
   - MODIFIED 「배달 실행자의 정지가 …」의 규범 문장은 정본 `:1000-1131`과 diff상 **글자 그대로**이고, blockquote 줄만 바뀌었습니다. 「등급화된 알림」 첫 문단·Scenario도 정본 `:178-183`과 같습니다.
8. **좌표 · AST · 반영 여부**:
   - D0.3h 인용은 전수 대조에서 참이었습니다: `mode.go:59-62`, `notifier.go:59` · `:285-293` · `:378-399` · `:382-383` · `:851/855/864`, `operating_mode.go:410` · `:417` · `:468` · `:475-479` · `:593-596`, `replay.go:535-551`(B2 `:548-550`, 삽입 `:551` 최상위 — 정정이 맞음), `gateway.go:249` · `:261-263` · `:269-271` · `:296`, `flatten.go:232-233`, `engine.go:639` · `:680`, `position_policy_command.go:98`, `auxiliary.go:99-109`, `backup.go:76`, `retry.go:532-539`, 정본 risk-management `:133`. 예외는 발견 8 · 14입니다.
   - 반영 누락: K9의 두 번째 `:102-108`, K12 RED, K13 FLM.
   - 반영 과잉·오류: K12 예외가 코드보다 좁음(발견 2), K8 서술(발견 10), K2 수단(발견 3).

리포지토리 상태: `rtk proxy git -C /mnt/D/Axipient/workspace/TossOS status --short -- openspec/changes/a092-an-alert-does-not-hold-the-stop` 출력은 `?? openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/review-23/` 하나뿐입니다(예상된 항목이며 건드리지 않았습니다). 사본과 임시 파일은 세션 scratchpad `/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/`에만 두었습니다.

Recommendation: BLOCK. 24판에서 발견 1~4(P1)를 문서로 고친 뒤 재리뷰하기를 권합니다. 네 건 모두 archive하면 정본에 들어갈 결함입니다: 입구 부류가 동기 claim 경로를 규범과 census 밖으로 떨어뜨리고, 「전이마다 통지」 SHALL과 먼저-잠금 SHALL은 착지한 HEAD 코드(a124 실행자·`Notifier.escalate`·a066 `:158`)가 이미 어기며, K2 RED는 a124 SHALL NOT을 어기는 구현으로 이끕니다.
