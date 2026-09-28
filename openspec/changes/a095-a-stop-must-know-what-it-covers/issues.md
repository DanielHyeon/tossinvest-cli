# a095 · issues — 5판

> 이 파일은 a095가 **고치지 않고 남기는 것**과 **다른 change의 선행 사실이 되는 것**을 기록한다.
> 분기 인용은 `analysis/function-logic/`의 번들(base `02716357`)에서 온다.

## I1. `baseline_price`의 쓰기 자리 넷 — 래칫 상향의 선행 사실 (tasks 5.3)

**2판의 「유효 손절가를 쓰는 경로는 하나뿐(재편입)」은 거짓이었다**(2라운드 P0-3). 3판이 AST로 열거한 사실:

| 자리 | 번들 | 값을 바꾸는가 | 하향에 대한 분기 |
| --- | --- | --- | --- |
| `Journal.OpenExitState` INSERT | `internal-journal--journal.openexitstate` | 최초 값(진입 손절) | — 행을 만든다(기존 값을 낮추는 쓰기가 아니다) |
| `Journal.recordExitJudgementTx` UPDATE | `internal-journal--journal.recordexitjudgementtx` | **예** — 판정마다 | 옛 경로 B23 `:487` → B25 `:491` `notBelow("baseline", …)`가 **스칼라** 기준 하향을 거부하고, B37 `:557`이 거짓이라 스칼라 열만 쓴다(effective JSON은 안 씀). 스냅샷 경로 B28 `:506` → B29 창 `exitpolicy.SelectRecoverySnapshot`은 **저장된 effective 스냅샷**과 비교하며, 저장 스냅샷이 없으면 비교 없이 재계산값을 받는다(`internal-exitpolicy--selectrecoverysnapshot` B2 `:138`) — 스칼라와는 비교하지 않는다 |
| `Journal.RefreshExitObservation` UPDATE | `internal-journal--journal.refreshexitobservation` | **조건부** — B23 `:127`이 `sameExitOperationalLine`으로 **effective JSON**과 비교해 거짓이면 거절한다. 스칼라와 effective 스냅샷이 일치하면 값이 움직이지 않고, 갈라져 있으면 스칼라를 스냅샷 보호가로 되돌린다(낮출 수 있다). 생산 도달은 재지 않았다 | B23 |
| `resetExitStateForReadoptTx` UPDATE | `internal-journal--resetexitstateforreadopttx` | **예** — 재편입 관측의 합성 손절로 | **없음 — 낮출 수 있다.** 분기 B1~B6 전부 오류·행 수 검사. 운영자 행동(`positionpolicy.ActionReadopt`)에서만 불린다 |

- 주석 *"the only reset writer for the four guarded execution-time columns"*의 네 열은 `guardedColumns` —
  `taken_ratio_total` · `pending_action` · `pending_level` · `pending_intent_id`(apply_hook.go `:503-505`)이고
  손절 열이 아니다.
- 475150의 57,900(본전 승격, 2026-08-07 측정)은 판정 경로의 산물이다.
- 정본 exit-policy 「Baseline Ratchet」과 「복구된 기준선은 낮아질 수 없다」가 판정 · 복구 경로의 단조를
  이미 요구한다.

**래칫 상향(불타기 방향)을 도입할 후속 change가 받는 사실**: 판정 경로에는 하향 거부가 있되 경로마다 비교 대상이
다르고(스칼라 · effective 스냅샷 · 비교 없음), reset 경로에는 없다. `EvaluateLadder`에서 **rung 잠금가와 수익률 기준,
그리고 R의 분모는 `entry_price`에서 나온다.** 그러나 마지막 rung의 runner 보호는 관측 워터마크 × (1 − trail%)이고
이전 기준선도 최댓값 합성에 들어간다(`ladder.go:391-403` — `internal-exitpolicy--evaluateladder` B16 `:386` 창의
`lockPrice(entry, …)` · B18 `:392` 창의 runner 후보 · `ComputeProtectedStop`). 2판~3판의 「모든 선이 `entry_price`에서 파생된다」는 과대 서술이었다(r3 N6). 이 사실을 a095 델타에 SHALL로
다시 세울지는 **Q6**(`proposal.md`).

StockOS `position-campaign-core` spec의 「하향 거부와 기록」 SHALL은 **라이브에 배선되지 않은 검토된
계약**이다(2판 I1 인용 유지 — `:52-56`, `:123-126`).

## I2. 총위험(R3) — 후속 change 후보 · `positions.avg_price`의 출처

2판은 *"브로커가 원가를 안 주면 직전 값이 이어질 수 있다"*고 적었다(`ApplyPositionAdjustment` `:312`
`firstNonEmpty(req.NewAvgPrice, target.AvgPrice)`). 보이스 A(A-5)는 더 강하게 주장한다 — 수량 수렴
경로(`ConvergeQuantities`)가 늘 `NewAvgPrice: ""`를 보내므로 수량 증가 경로에서 평단은 **설계상** 낡는다.
반대 사례도 있다 — 272210 inst4 `avg_price = 81922.222222`(2라운드 P2-4).

**3판은 이 주장을 재검증하지 않았다**(`ConvergeQuantities`의 번들이 없다).

**처분 — 사용자 결정(2026-09-28)**: *"Q5 = R3(총위험) 보류 + delta 의 SHALL 지위 해제(archive 차단 해소, 후속
change 후보로 기록)"*. a095 exit-policy 델타에서 총위험 요구를 지웠다. **후속 change 후보**:
「보호는 자기가 덮는 수량과 총위험을 말한다」 — 착수 조건은 (1) `ConvergeQuantities` · `ApplyPositionAdjustment`의
평단 경로를 AST 번들로 먼저 재검증, (2) 평단 출처를 정할 것(브로커 `Holding.CostBasisRaw` 등), (3) 2판 D3 ·
tasks 3.1~3.6(`2fbdcd78`)을 이력으로 참조.

## I3. 발신 자리 넷은 같은 사실이 아니다 (2판 판단 정정)

2판은 넷을 「가진 것에 손절이 안 걸려 있다」는 같은 사실로 묶었다. **틀렸다.**

| 자리 | 사실 | 3판 |
| --- | --- | --- |
| `ExitObserver.alertUnmanaged` (`workingSet` B6) | exit 관측의 적격하지 않은 보유 | normal 유지(결정 (1)) |
| `ReconcileDriver.alertUnmanaged` (`judgeHoldings` B15) | 사유가 B3~B6 · 기본으로 갈린다 | 사유별(`design.md` D1) |
| `checkExternalIncrease` (`judgeHoldings` B8) | 편입 후 수량 증가 — **증가분은 원래 손절의 보호를 받는다**(본문 adoption.go `:461`) | Q4 |
| `notifierAlerter.ExternalPositionFound` | fold 알림 | **생산 도달 불가** — `IngestExternalPositions` B12 `in.Alert == nil`, `reconcileloop.go:338` |

앞의 둘은 같은 키 철자 `exit.position_unmanaged|<posID>`를 쓴다 — 결정 (3)(iii)이 가른다.

## I4. 배포 재생 결과 — 미실행

`[미실행]` — 구현 뒤 tasks 3.4(Q4)의 재생과 7.2의 배포 직전 측정 결과를 여기에 적는다.

## I5. 현재 열린 포지션

2판은 2026-08-07 원장으로 *"010170 30주가 무보호"* 등을 적었다. **2026-09-25 재조회(2라운드 P1-6)에서
그 동기 포지션은 TSLA 먼지 1건을 빼고 전부 CLOSED였다.** 3판은 소급 보호를 하지 않고, 배포 직전의
원장 측정을 tasks 7.2에 둔다. `alert_outbox`의 `exit.position_unmanaged` 0행(2026-09-25, 16행 전부
critical)은 결함 계열이 남아 있다는 증거로 유지한다.

## I6. 운영자 재편입 reset의 비교 없는 하향 — 해제 원칙 가족의 잔여 질문 (4판, 3라운드 V-N5)

`resetExitStateForReadoptTx`(`internal-journal--resetexitstateforreadopttx`)는 분기 B1~B6 전부가 오류 · 행 수
검사이고 이전 `baseline_price`와 비교하는 분기가 없다 — 운영자 재편입은 유효 손절을 **낮출 수 있다**(I1 표).
4판 exit-policy 델타는 평단 비하향 요구를 **자동 경로**로 한정하고 이 경로를 명명해 뺐다(Manager 처분 2026-09-28).

**남는 질문**(요구로 굳히지 않는다): 사람 소유의 하향 — 운영자 재편입으로 손절이 내려가는 것 — 에 승인 참조 ·
audit를 물릴 것인가. 같은 가족의 선례는 a092 21판의 운영 모드 완화(OPERATOR + 승인 참조 + commit 전 audit, 사용자
결정 2026-09-28)와 a066의 완화 · overage 해제다. a095는 이 질문을 여기 등록만 하고 답하지 않는다.

## I7. 편입이 커밋됐는데 보호가 열리지 않은 보유 — 후속 후보 (5판, r4 R4-3)

`adoptOne`(`internal-app-engine--reconciledriver.adoptone`) B3 `:344` 창: `OpenAdoptedExitState`가 실패해도 로그를 남기고
`:379`에서 **true**를 돌려준다. 주석은 편입 기록이 커밋됐으므로 「편입됨」이 참이고 다음 사이클이 개설을 마친다고 적는다.
그 사이 보유는 exit state 없이 편입된 상태다.

a095는 이것을 critical 요구에 **넣지 않고**(델타가 「이 요구사항이 정하지 않는다」로 명명) 이름 붙은 경계로 남긴다
(Manager 처분 2026-09-29 — 조용히 범위에 넣지 말 것). **후속 후보**: 커밋 뒤 보호 미개설이 몇 사이클 지속되면 무엇으로
알리는가, 그리고 exit 관측 루프는 그 보유를 어떻게 보는가.
