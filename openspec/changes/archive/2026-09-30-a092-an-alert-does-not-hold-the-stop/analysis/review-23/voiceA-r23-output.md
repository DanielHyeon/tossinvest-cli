<!-- 23라운드 보이스 A(생산·안전 렌즈) — Claude code-reviewer 서브에이전트(독립 컨텍스트, 같은 하네스). 프롬프트 prompt-common.md + 호출 렌즈.
     대상 a88e7079(리뷰 기준 HEAD 7002270f, MANIFEST 운영 파일 13개 sha 일치). 원문 그대로. -->

**판정: BLOCK** — P0 0 · P1 2 · P2 9

23판을 적힌 대로 구현해도 안전 불변식은 지켜진다. 손절 경로는 약화되지 않고, 토글 OFF 영향과 LIVE side effect는 없다. exit goroutine에 원격 대기도 새로 생기지 않는다(판정 근거는 아래 1~8).
BLOCK인 이유는 문서 두 곳이다.
- ADDED에 새로 넣은 전칭 SHALL이 같은 델타의 SHALL과 모순되고, HEAD의 무통지 승격과도 모순된다(#1).
- 델타가 금지한 「먼저 잠그지 않고 원장에 직접 쓰는 기록자」가 HEAD 생산 경로에 이미 있고(a066), archive 선행 조건이 적혀 있지 않다(#2).

판정 6(flatten CLI)의 답: **닿지 않는다.** 잔여는 오늘 도달 0이다.

좌표 기준: HEAD `7002270f`, 리뷰 대상 `a88e7079`. MANIFEST에 든 운영 파일 13개의 sha는 HEAD와 전부 일치한다(직접 대조). `risk_relaxation_command.go`는 살아 있는 a066 로트가 작업 트리를 고치는 중이라 `git show HEAD:`로 읽었다.

## 발견 표

| # | 등급 | 주장 | 증거(파일:줄) | 권고 |
|---|---|---|---|---|
| 1 | **P1 (T)** | ADDED의 「운영 모드 전이의 통지는 **상태를 바꾼 전이마다** 한 번 나가야 한다(SHALL)」는 전칭이다. 이 문장이 모순되는 곳은 셋이다. ① 같은 델타 :77 「durable 기록 실패 경로의 승격은 **통지하지 않는다**(SHALL)」 — 그 승격도 NORMAL→ENTRY_BLOCKED로 상태를 바꾼다. ② HEAD의 전달 실패 승격 둘은 설계상 announcer가 nil이다. ③ 새 Scenario 「재알림 창 안의 재강화도 통지된다」(자동 트리거 전반)는 트리거가 `CRITICAL_ALERT_UNDELIVERED`이면 거짓이다. design D0.3h K16이 말하는 「같은 호출 안의 `n.escalate`가 방금 푼 모드를 다시 조인다」가 정확히 그 **무통지** 재강화다. 이대로 archive하면 정본에 서로 다른 SHALL 둘이 부딪힌다(문언상 P0 요건). 행동상 안전 손실은 없어 P1(T)로 잡았다. | spec.md(engine-safety 델타):324 · :77 · :362-364; `internal/obs/notifier.go:360-368`(No announcer 주석) · `:382-383`(`EscalateOperatingMode(…, nil)`); `internal/app/engine/alertdelivery.go:451-452`(nil) | :324를 「announcer를 받는 전이」 또는 「통지가 요구되는 전이(자동 강화 중 전달 실패 트리거와 durable 실패 경로 제외 · 사람 완화)」로 한정한다. Scenario :362는 트리거 이름을 적는다(예: 관측 두절). |
| 2 | **P1** | a066 완화 통지는 `repo.EnqueueAlert`로 critical 행을 원장에 직접 넣는다. release seq마다 새 키이고, 먼저 잠그지 않는다. 델타 :67 「입구를 거치지 않는 기록자는 행을 넣기 **전에** 자기 사유를 세워야 한다(SHALL)」를 HEAD 생산 경로(엔진 제어 소켓)가 이미 어긴다. tasks 23.4는 「확인」이지 「준수」가 아니고, archive 선행 조건이 없다. 권하는 두 선택지 가운데 「먼저 잠금」은 완화 통지 때문에 새 진입 차단을 세우는 것이다. 그러면 사람이 다시 풀어야 하는 마찰이 생기므로 사실상 불가하다. | `git show HEAD:internal/app/engine/risk_relaxation_command.go` :40(인터페이스) · :151-173(`notifyRelaxation`, :158 적재, 잠금 없음); 살아 있는 작업 트리 로트도 같은 모양(:168); design D0.3h 4 표 | archive 게이트를 명시한다: 「a066 기록자가 입구(`RecordAlert` via 알림기)로 옮겨진 커밋을 인용하기 전에는 a092를 archive하지 않는다」. 아니면 델타에 이름 붙인 예외로 적는다. a066 교차 통지에서 「먼저 잠금」 선택지는 뺀다. |
| 3 | P2 (T) | K2 (ii)의 수단 위치가 틀렸다. HEAD에서 승격은 `deliver`의 래치 자리가 아니라 `claimAndDeliver`가 돌아온 뒤 `notifyCritical` :228에서 일어난다. `:520`(시도 기록 NotFound)은 `lost=true`로 돌아가(:523 → :310-314에서 owed=false) **승격 자체가 없다.** 그래서 tasks 23.3 K2 「세 래치 자리 각각」의 RED는 :520에서 성립하지 않는다. 그 RED를 맞추려고 :520에 승격을 더하면 a124의 「승격을 포함하지 않는 판정은 승격을 만들어서는 안 된다」를 어긴다. 델타 :69는 그 조항과 「원장 승인 시각으로 순서 추정 금지」를 이식하지 않았다. | notifier.go:223-229 · :309-316 · :519-523 · :571-573; a124 archive `specs/engine-safety/spec.md:31-39` | (ii)는 승격 호출이 돌아온 자리에서 :484·:571 판정에만 건다. K2 RED에서 :520을 뺀다. 델타 :69에 a124의 두 조항을 옮긴다. |
| 4 | P2 (T) | 판정 6: flatten CLI는 `ReplayInDoubt`에도 `parkAlert`에도 **닿지 않는다**(아래 6). 엔진에서도 `Attested: nil`이라 `parkAlert`는 오늘 도달 0이다. D0.3h 4의 잔여 서술(「그 프로세스의 parkAlert가 먼저 잠그는 것은…」)은 측정 전 가정이다. 또 flatten `Saga.Notifier` 필드가 존재한다. 누가 배선하면 둘째 프로세스의 알림기가 `ClaimAlertForDelivery`로 행을 넣는데, 이것은 K6 census(이름 `EnqueueAlert`·`RecordAlert`)에 걸리지 않는다. | `internal/flatten/flatten.go:80` · :689-693; `cmd/tossctl/flatten.go:247-263`(Notifier 없음) · :233-240(Replay 없음) | D0.3h 4와 tasks 23.2를 측정값(도달 0)으로 바꾼다. 핀을 둘 둔다: 「flatten 조립에 Recovery·Replayer·Replay·Notifier 없음」, 「비시험 `obs.Notifier` 생성은 `newNotifier` 하나」. 델타 :65-67의 입구·셈~해제는 「엔진 프로세스」로 한정한다. |
| 5 | P2 | K5 핀은 「`Entry` 값이 같은 함수의 `NewEntryGate` 결과」만 센다. 역할(그 게이트가 `Acknowledge`가 푸는 알림기 게이트와 **같은 것**)은 세지 않는다. 엔진 조립을 `Entry: execgw.NewEntryGate(in.clock, nil)` 인라인으로 바꾼 변이가 살아남는다. flatten의 게이트는 원래 다른 프로세스 것이라 역할을 증명하지 못한다. | `internal/app/engine/gateway.go:249` · :302 · :323(`newNotifier(…, entry, …)`); `cmd/tossctl/flatten.go:232-239` | 핀을 「`execgw.New`의 Entry 식별자 == `newNotifier` 게이트 인자」 동일성으로 바꾼다. flatten은 「알림 기록자 도달 0」으로 센다. |
| 6 | P2 (T) | K3 핀 문언 「Commit과 Project 사이에 `go`·반환 없음」은 HEAD 양성 대조군에서 실패한다. Commit 오류 갈래의 `return`(:469, B25 안)이 Commit 호출(:468)과 `ProjectOperatingMode`(:476) 사이에 있다. projector nil로 건너뛰는 조건(:475)을 어떻게 다루는지도 정하지 않았다. | `analysis/head-ast-21/…transitionoperatingmode.json` B25 `:468` · return `:469` · B26 `:475`; operating_mode.go:468-477 | 구간을 「Commit을 담은 if 문 다음 문장 ~ Project 호출」로 정의한다. nil 갈래는 생산 배선 핀이 진다고 적는다. |
| 7 | P2 (T) | 델타 :65 「임차 없는 **재무장** 기록은 입구를 통해서만(SHALL)」은 문언상 거짓이다. `EnqueueAlert`(재알림 창 0)도 모르는 상태의 행을 PENDING으로 재무장한다. 그 행은 `UndeliveredCount`(state=PENDING만)에 안 세이므로 셈~해제 사이 재무장이 셈을 바꾼다. 실제로는 키가 attempt·seq마다 유일해 도달 0이다. | outbox.go:142-146 · :411-416(default → rearm) · :555-561 | 문장을 「재알림 창에 의한 재무장」으로 한정하거나 census에 넣는다. |
| 8 | P2 (T) | K1 신원으로 rowid를 쓰면 `VACUUM INTO` 복원 뒤 재번호에 기댄다(design 스스로 「추정」). 전이 행에는 이미 안정된 유일 신원이 있다(`operating_modes.id` TEXT PRIMARY KEY, 레코드 `ID`). | core_domain.go:185-186; operating_mode.go:440-454; backup.go:76 | 통지 키는 `rec.ID`로 하고, rowid는 울타리 순서에만 쓴다. |
| 9 | P2 (T) | 좌표 누락·오류가 넷 있다. ① D0.3g 8 C14 `engine_risk_relaxation.go:136,194` → 실제 `:130,184`(HEAD·a88e7079 동일)인데 D0.3h 6 정정표에 없다. ② MODIFIED ✅ 블록에 낡은 `risk-management :102-108`이 남았다(델타 :229 — 머리 주석은 「넷」을 고쳤다고 적었고, :198의 같은 좌표만 `:133`으로 고쳤다). ③ K14 `OperatingModeHistory :593-596` — ORDER BY는 :597이다. ④ a066 `:158`은 작업 트리 로트에서 :168로 밀린다. census는 이름으로 센다. | `git show HEAD:cmd/tossctl/engine_risk_relaxation.go` Use: :130 · :184; spec.md(델타):229; operating_mode.go:595-597 | 정정표에 더하고, 이름 기반 census로 둔다. |
| 10 | P2 (T) | ADDED 「통지만 실패하면 완화 성공 + 통지 실패를 함께 보인다」: `Notify`는 **전송** 실패에 nil을 돌려준다. 그래서 `ErrModeAnnouncementFailed`는 기록 실패에서만 난다. 전송 수단이 죽은 상태의 완화는 단서 없이 「완화됨」으로 보고되고, 같은 호출의 `escalate`가 무통지로 다시 조인다. K16의 재읽기가 부분적으로 보완한다. | notifier.go:124-130 · :200-230; mode.go:57; operating_mode.go:478-482 | 「통지 실패 = 기록 실패」로 정의를 명시한다. 명령 출력에 통지 행의 PENDING 여부를 보인다. |
| 11 | P2 | K8 「풀 경로 = 원장 수리 뒤 재시작」은 실제보다 좁다. 복원이 일시 오류로 실패했고 원장이 ENTRY_BLOCKED로 읽히면, `mode-release`가 NORMAL 행(rowid > 울타리 0)을 커밋해 복원 실패 래치를 푼다(OPERATOR+audit이라 구멍은 아님). tasks 머리는 「22.3 C15 풀 경로 좁힘」이라 적었지만 22.3 C15 본문과 RED는 그대로이고, K12(무통지) RED도 없다. | operating_mode.go:423-434 · :475-476; modegate.go:35-51; tasks.md:15 · :64 | 문장을 정정하고 K8·K12 RED를 더한다. |

## 판정 1~8

1. **K1 — 부분 성립.**
   - 「변화 없으면 통지 전 반환」은 참이다. AST 기준 B15 `:410` → return `:415`, B16 `:417` → return `:421`이고, 통지(B27 `:478`)보다 앞이다. 완화는 사람만 할 수 있으므로 통지 수는 사람의 행동 수에 묶이고, 폭주는 없다.
   - 정본 「같은 조건의 critical 알림은 재알림 창 안에서 한 번만」(정본 engine-safety :708-710)과는 모순이 아니다. 키 정의는 발행자 몫이고, 재강화는 새 에피소드다.
   - 다만 새 전칭 SHALL과 Scenario가 무통지 승격 경로와 모순된다 → #1(P1 T). 신원은 rowid보다 `id`가 낫다 → #8.
2. **K2·K4 — 규범 핵심은 a124와 같다.** 해제 뒤 차단 생략 + 승격 적용, 허용 창, 승격 실패 시 무조건 차단 — 델타 :69 ↔ a124 :31-44.
   - 빠진 조항: 「근거 확정 앞의 해제는 판정 불변」은 Scenario와 task에만 있다. 「승인 시각으로 추정 금지」와 「승격 없는 판정은 승격 금지」는 없다 → #3.
   - **게이트를 잘못 여는 경로는 찾지 못했다.** 해제는 `Acknowledge`가 `n.mu` 아래에서 미전달 0을 확인한 뒤에만 한다(notifier.go:851-876). 발송 중의 해제는 그 행의 승인을 뜻하므로, 정산이 `SettleAlreadySettled`로 끝나 lost 갈래로 간다(:440-451 · :548-561). 승인 뒤 재잠금은 허용 창과 (ii)에서만 일어나고, 둘 다 보수 방향이다.
3. **K3 — 해석은 정본 및 HEAD와 정합한다.**
   - HEAD의 순서는 Commit `:468`(B25) → 투영 `:475-476` → 통지 `:478-479`이고, 정본 risk-management :133 「동시에」와 맞는다.
   - 비시험 `EscalateOperatingMode`·`TransitionOperatingMode` 호출자는 전부 엔진 프로세스 안이다: notifier.go:382, runtime.go:462, alertdelivery.go:451, exitloop.go:846, retry.go:413, riskguardian.go:644. 프로세스 밖에서 전이하는 곳은 0이다.
   - 핀 문언은 HEAD에서 스스로 실패한다 → #6.
4. **K5 — 사실은 참이다.** `execgw.New(` 비시험 호출자는 정확히 둘이다(gateway.go:296 · flatten.go:233). 둘 다 같은 함수의 `NewEntryGate`에서 `Entry`를 받는다(:249→:302 · :232→:239). 형태는 「전수 + 존재」다. 호출자를 더한 변이와 Entry를 뺀 변이는 잡힌다. 「같은 게이트」라는 역할은 안 센다 → #5.
5. **K6·K7 — 입구 정의는 타당하다.** 재무장 SHALL을 exit 기록으로 한정한 것은 a094·a090의 enqueue-only(재알림 창 0)와 양립한다.
   - HEAD census는 맞다: 정의 outbox.go:131, 호출자 replay.go:551, risk_relaxation_command.go:158(인터페이스 :40). 셋 다 rg로 확인했다.
   - 빈틈: 프로세스 밖 알림기(#4), 모르는 상태의 행 재무장(#7), a066 위반(#2, P1).
6. **프로세스 밖 기록자(flatten) — 닿지 않는다. 등급 P2 (T).** 근거 사슬:
   - `ReplayInDoubt`의 비시험 호출은 `reconcile/recovery.go:351`(`r.opts.Replayer`) 하나다.
   - Replayer 주입과 `reconcile.New`는 엔진 `runtime_wiring.go:184` · `:188`뿐이다.
   - flatten은 `reconcile.Collector`·`Stabiliser`만 쓴다(liquidate.go:578 · :587). `buildFlattenWiring`에는 Recovery가 없다(flatten.go:199-272).
   - 호출된다 해도 flatten 게이트웨이에는 `Replay`가 없어 `replay.go:249-251`에서 반환한다. 엔진은 `Attested: nil`(gateway.go:312-315)이라 `:257-259`에서 거절한다. 그래서 `:358`의 `parkAlert` 호출은 생산 어디서도 도달 0이다.
   - flatten의 다른 경로도 알림 행을 쓰지 않는다: `Resolver.park`는 Block만 한다(indoubt.go:372-383). `Saga.event`는 Notifier가 nil이라 로그만 남긴다(flatten.go:689-693).
   - 만약 닿게 되면, 엔진의 셈~해제 창에 들어온 행 때문에 엔진 게이트가 열린 채 PENDING critical 행이 생긴다. 그 기간의 상한은 a124 래치 시간이고, 재시작 복원이 다시 잠근다. flatten 프로세스가 자기 게이트를 먼저 잠가도 엔진은 보호되지 않는다. 그 시점에는 P1이 된다. 그러니 지금은 핀으로 도달 0을 고정한다(#4).
7. **archive 정합성 — 결함 셋.**
   - SHALL 대 SHALL 모순(#1), HEAD에서 거짓인 SHALL(#2), 문언상 거짓(#7).
   - MODIFIED 「배달 실행자의 정지가 …」의 규범 문장은 바뀌지 않았다. 정본 :1000-1131과 델타 :159-299를 diff하면 차이는 인용 블록 추가와 좌표 병기뿐이다. 다만 낡은 좌표 하나(:229)가 남았다 → #9.
8. **좌표·AST — 대체로 정확하다.** 직접 확인한 것: mode.go:59-62, notifier.go:59 · :285-293 · :378-399 · :875, operating_mode.go:410 · :417 · :468-479, replay.go:535-551, engine.go:639 · :680, position_policy_command.go:98, outbox.go:140 · :142-146, auxiliary.go:99-109, backup.go:76, gateway.go:261-271, retry.go:532-539, risk-management :133.
   - 오류와 누락은 #9와 같다.
   - 22판 C2 이후 exit goroutine이 기다리는 것은 `n.mu` 아래 로컬 작업뿐이다. 23판이 더한 것은 입구(로컬 `RecordAlert`)와 동기 발송자 쪽 (ii)뿐이라 **원격 대기 0은 유지된다.**
   - 22라운드를 「적은 대로」 반영했는지는 review.md 열람 금지로 대조하지 못했다. 내부 정합만 봤다: tasks 머리의 「22.3 C15 좁힘」은 본문에 반영되지 않았고, K8·K12 RED가 없고, K2 RED는 범위가 과하다(#3·#11).

git 상태(`rtk proxy git -C … status --short -- openspec/changes/a092-an-alert-does-not-hold-the-stop`):
```
?? openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/review-23/
```
기대한 그대로다. 저장소 파일은 만들거나 고치지 않았다. 스크래치는 `/tmp/claude-1000/a092-r23/voiceA-scratch/{canon.md,delta.md}`(정본/델타 diff용)뿐이다.

Recommendation: BLOCK — 24판에서 #1(전칭 통지 SHALL을 announcer를 받는 전이로 한정)과 #2(a066 기록자가 입구로 옮겨진 커밋을 archive 선행 조건으로 명시)를 고치고, #3~#6 핀·수단 위치를 정정한 뒤 재리뷰 because 안전 불변식과 exit goroutine 원격 대기 0은 서 있지만, 이대로 archive하면 정본에 SHALL끼리의 모순과 HEAD 생산 경로가 어기는 SHALL이 들어간다.
