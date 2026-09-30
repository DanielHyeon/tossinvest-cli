# a095 · tasks — 10판 (구현 로트)

- **Change**: `a095-a-stop-must-know-what-it-covers`
- **위험 등급**: **High-risk** — 무보호 보고의 등급과 진입 차단 도달. §0.3 적용.
- **base-commit**: `1ffe2295994de4ca4530fc1086535c1f7dc37835` (구현 로트 재고정 `821ab02a`, 이전 3판 `0b17784e`)
- 결정 (1)~(3) 원문은 `proposal.md` §0. `[비움 — Qn]`은 결정이 덮지 않아 Manager 에게 올린 질문이다
  (`proposal.md` 「열린 질문」). 비운 task는 답이 오기 전에 착수하지 않는다.

## 0. 게이트 선행

- [x] 0.1 `base-commit.txt` 고정 — **3판 재고정** `ec29dc72` → `02716357`(`0b17784e`). a095 디렉터리를 만진
      커밋 중 `.go`를 고친 것은 `a30eb35a` 하나이고 새 base 앞이다
      > **구현 로트 재고정(2026-09-30)** `02716357` → `1ffe2295`(`821ab02a`, 단독 커밋) — 자기 Go 편집 0, 옛 base required 85는 형제 착지 몫.
      > 승인: 사용자 상임 지시 + 2026-09-30 재개 지시 → Manager 배정. `review.md` §4.1 · 「착수 승인 기록」
- [x] 0.2 `openspec validate a095-a-stop-must-know-what-it-covers --strict` — 3판에서 다시 통과(`review.md` 3판 기록)
- [x] 0.3 **AST 산출물이 문서보다 먼저** — 3판 번들 21개(새로 12 · 다시 뽑음 4 · 해시 일치로 산문만 5).
      생성기 `analysis/harness/render_bundles.py`, 커버리지 `analysis/harness/coverage/`
- [x] 0.4 `check_analysis.py --change a095-…` — 3판 판정은 `review.md` 3판 기록(격리 worktree, 완료 커밋 기준)
- [x] 0.5 **proposal-freeze 리뷰**(적대적 Eng 필수) → `review.md`. 교차 보이스 · 교차 모델을 여기서 지킨다
      > **freeze 성립(2026-09-29, 9판 `5d2f1b6e`)**: 교차 모델 codex 6라운드 **PASS**(`eccb07cb`, 6판) → 7판 Claude 독립 보이스 두
      > 실행(r7a APPROVE · r7b REJECT P1 2) → 8판 반영 → 좁힌 재검증 REJECT(P1 1) → 9판 반영 → 좁은 재검증 Claude 독립 보이스
      > **APPROVE**(P0 0 · P1 0). 남은 P2 1 · P3 4(r9 R9-1~R9-5)는 기록 — `review.md` §3.24. 구현 착수는 0.7(Manager 스케줄링)을 거친다
      > **2라운드(2026-09-25) FAIL**(`review.md` 「2라운드」). **사용자 결정(2026-09-25)**: §2.11 1 **독립** ·
      > 2 **거부** · 3 **수용**(범위 이동). 3판이 그것을 반영했다. 3라운드는 Manager 가 따로 지시한다
- [x] 0.6 **열린 질문 Q1~Q7의 답** — 답이 온 뒤 비운 절을 채우고 3라운드 전에 다시 뽑을 번들을 정한다
      > **파킹(2026-09-27, Manager 분류 — `review.md` §3.8)**: 사용자행 Q5 · Q2(d) · Q7, 구현 로트행 Q1 · Q2(a)(b)(c) ·
      > Q3(정지 조건) · Q4 · Q6. 교차 모델은 Codex 401 가능성 — 실행 시점에 확인
      > **사용자 결정(2026-09-28)**: Q5 보류 + 델타 SHALL 해제 · Q2(d) → a124 정본 · Q7 → a092 21판(둘째 면은 이름
      > 붙은 잔여). **남은 열림은 구현 로트행 Q1 · Q2(a)(b)(c) · Q3 · Q4 · Q6** — 착수 시 코드 영수증으로 확정
      > **구현 로트 답(Manager 판정 2026-09-30)**: Q1 (a) 새 critical 종류 · Q2(a)(b)(c) normal · Q3 원장 조정 순증(journal 읽기 leaf) ·
      > Q4 normal + 최대 수량 래치 · Q6 후속 이월 · Q8 (b) 시점 사건 문구. 델타에 규범 추가(10판), 값 · 영수증은 design.
      > `review.md` §4.3 · §4.4

- [x] 0.7 **구현 착수 순서 — Manager 스케줄링 사항**(9판 r8 N1 · N2, 10판 r9 R9-1 · R9-2 강화). 구현 착수 순서는 Manager 스케줄링 사항이다. **기본 스케줄은 a092 「모든 보유자」 착지 이후다.** 그 **전에** 착수하려면 Manager 승인만으로는 부족하고 **사용자 확인**이 필요하다 — 추가되는 대기가 손절 루프의 `o.alert` 네 자리(`exitloop.go:831` 관측 두절 · `:1633` 판정 거부 · `:1657` 제안 거부 · `:1687` 청산 지연)를 지나 안전 불변식 4(손절 즉시성)에 닿기 때문이다. **깨질 때의 행동**: 착지 전인데 사용자 확인 기록이 없으면 착수 금지. 사용자 확인과 Manager 승인은 이 change의 `review.md` 「착수 승인 기록」 절에 적고, tasks 7.3 공시에 증폭 수용을 명기한다. 「a092 『모든 보유자』 착지」의 판정 기준은 a092 tasks **21.4**(GREEN — `claimAndDeliver` 잠금 범위)가 체크된 커밋이 이 브랜치 역사에 있고 그 커밋에서 a092 tasks **21.3 (f)**(다른 호출자의 동기 발송이 원격 전송 중일 때 exit 기록이 그 전송을 기다리지 않는다) 시험이 통과하는 것이다 — 그 커밋 해시를 기록에 적는다.
      결정 (1)(범위)은 무접촉이다
      > **충족(2026-09-30)**: a092 21.4 첫 체크 커밋 `06b39a78`, 그 커밋에서 21.3 (f) `TestA092ARecordDoesNotWaitForAnotherSendersTransport`
      > PASS · GREEN 코드 `fbc6df5f`. 사용자 확인 불요. `review.md` 「착수 승인 기록」

## 1. 산출물 (3판 — 문서보다 먼저)

- [x] 1.1 다시 뽑음(소스 줄 이동): `obs.SeverityOf` · `obs.Notifier.Notify` · `obs.Notifier.publishBestEffort` ·
      `journal.resetExitStateForReadoptTx`
- [x] 1.2 호출자(2라운드 §2.16 지시): `engine.ReconcileDriver.judgeHoldings` · `engine.ExitObserver.workingSet` ·
      `engine.notifierAlerter.ExternalPositionFound`
- [x] 1.3 3판 주장의 근거로 새로 뽑음: `engine.ExitObserver.ObserveOnce` · `engine.ExitObserver.alertUnmanaged` ·
      `engine.ReconcileDriver.adopt` · `obs.Notifier.notifyCritical` · `obs.Notifier.claimAndDeliver` ·
      `obs.Notifier.deliver` · `reconcile.Ingestor.IngestExternalPositions` · `journal.Journal.recordExitJudgementTx` ·
      `journal.Journal.RefreshExitObservation`
- [x] 1.4 해시 일치 확인(산문만 3판으로): `engine.ReconcileDriver.checkExternalIncrease` ·
      `engine.ReconcileDriver.alertUnmanaged` · `journal.Journal.OpenExitState` ·
      `journal.Journal.ApplyPositionAdjustment` · `exitpolicy.EvaluateLadder`
      (앞의 둘은 생성기로 다시 씀 — 2판 FLM의 「B2 — 미편입 보유가 여기로 온다」 거짓 정정)
- [x] 1.5 **발신 자리 넷의 대조**(2판 1.11) — 결과: **같은 사실이 아니다.** exit 관측 · reconcile 무관리 ·
      수량 증가(보호 중) · fold(생산 도달 불가)로 갈린다. `design.md` D1 「발신 자리 넷의 3판 처분」
- [x] 1.7 **4판 근거 번들 5개**(r3 N3 · N5 처분의 분기 근거): `journal.Journal.recordAlertTx` · `journal.claimOwed` ·
      `engine.alertDeliverer.cycle` · `engine.alertDeliverer.deliverOne` · `exitpolicy.SelectRecoverySnapshot`.
      HEAD `04a0dd25` 깨끗한 연결 worktree에서 AST · 커버리지(`analysis/harness/coverage/r4-*.out`)
- [x] 1.8 **5판 근거 번들 2개**(r4 R4-1 · R4-2 · R4-3): `engine.ReconcileDriver.adoptOne` ·
      `engine.resolveNotificationPublisher` — HEAD `f69a3dab` 깨끗한 연결 worktree의 AST · 커버리지
      (`analysis/harness/coverage/r5-app-engine.out`). `alertUnmanaged` · `adopt` 번들 결론을 편집 경계로 갱신
- [x] 1.9 **6판 근거 번들 1개**(r5 R5-1): `engine.ReconcileDriver.alert`(B2 — `Notify` 오류를 로그로만) — 커버리지
      `analysis/harness/coverage/r6-app-engine.out`(HEAD `f69a3dab` 실행, `reconcileloop.go` base 이래 무변화)
- [x] 1.11 **9판 근거 번들 1개**(r8 N3): `config.Adoption.validate` — 같은 `r8-config.out`
- [x] 1.10 **8판 근거 번들 2개**(r7 F1 · F3): `config.mergeAdoption` · `config.mergeNotifications` — 커버리지
      `analysis/harness/coverage/r8-config.out`(작업트리 `internal/config` 무수정 · base 이래 무변화 확인 뒤 실행). 생성기의 오류
      계약을 번들별 명시로 바꿔 21개 맵을 정정(아래 review §3.21 영수증)
- [x] 1.6 **a091과의 병합 확인** — 둘 다 등급 표를 건드릴 수 있다. 3판은 표 크기를 시험에 적지 않는다
      > **2026-09-30**: a091은 활성 · 미구현(tasks 2.4 GREEN 미체크). 두 change 모두 `criticalEvents`에 **줄을 더할 뿐**이라 충돌은 map 리터럴의
      > 인접 줄뿐이다. a095 시험은 표 크기를 단언하지 않고 자기 종류의 등급만 단언한다(`SeverityOf(EventExitPositionAdoptionFailed)`)

## 2. R1′ — 등급을 사실로 (결정 (1)·(2), design D1)

- [x] 2.0 **Pre-Edit 선언** (`review.md` §4.5) — 대상은 Q1의 답이 정한다(10판: `adopt` 결과 형태 · `judgeHoldings` · `alertUnmanaged` · `checkExternalIncrease` · `NewReconcileDriver` · `Context.ReconcileDriver` + 새 종류 등재 · journal 새 leaf)
- [x] 2.1 **RED** — exit 관측 자리(`workingSet` B6 → `ExitObserver.alertUnmanaged`)의 사실은 **normal**이고
      `Notify` B1 창(`publishBestEffort`)으로 간다: outbox 행 0 · `n.mu` 미획득. 기존
      `TestAPositionWithNoEntryDecisionIsSkippedAndAlertedOnce`의 normal 단언을 유지한다
- [x] 2.2 **RED** — **알림 켜짐** · 편입 켜짐 · 편입 시도 실패(`adopt` B8 거짓 — `adoptOne` 범주 ① 편입 전 거절 ·
      ② 영속 실패, 두 분기 모두 오늘 미진입)의 reconcile `alertUnmanaged` B5
      사실은 **critical**이고 outbox 행을 만든다. 연기분(`adopt` B2 · B6 · B7)은 이 단언에 넣지 않는다(r3 N2,
      Q2(c) 열림). 싣는 방식은 `[비움 — Q1]`. **생산 배선**(실 `Notifier` · outbox · 배달 실행자)으로 잰다 —
      발신 가짜로만 재면 안 된다(r3 N1 codex 제안)
- [x] 2.3 **RED** — exclude(`judgeHoldings` B11 · `alertUnmanaged` B4)는 **normal** — outbox 행 0 · 진입 게이트
      래치 0 · 운영 모드 승격 0
- [x] 2.4 **RED** — `adoption.enabled=false` ∧ 미지정(`judgeHoldings` B12 · 기본 사유)은 **normal** — 2.3과 같은
      단언. 정본 exit-policy 「false에서의 동작은 무관리 보유 알림을 포함한 기존 동작과 동일」의 동등성 시험
- [x] 2.5 **RED** — **알림 off**(`notifications.enabled=false`) 엔진에서 B5 사실은 **critical로 기록되지 않는다**
      — outbox critical 행 0, 따라서 `deliver` B3 · `notifyCritical` B4 · 배달 실행자 `deliverOne` B8 사슬에 닿지
      않고 진입 게이트 래치 0 · 승격 0(r3 N1). 이미 원장에 남은 critical 행의 재시작 차단은 정본대로 유지됨을 함께
      단언한다. 방식은 `[비움 — Q1]`. **꺼짐 + topic 유지**(파일에 topic이 남음) 경우를 포함한다(r4 R4-2)
- [x] 2.5a **RED** — **켜짐 + topic 없음**(전송기 nil): 판정 근거는 설정 `enabled`이므로 B5 사실은 **critical**이고,
      그 행은 배달 실행자 `deliverOne` B8 · a124 정본대로 처리된다. `Publisher == nil`을 「꺼짐」으로 읽는 구현은
      이 시험에서 실패해야 한다(r4 R4-2). 생산 배선으로
- [x] 2.5b **RED** — **거부된 알림 블록**(파일에는 `enabled: true`, 검증 실패): 로드된 값이 거짓이므로 꺼짐 — a095 사실은
      critical로 기록되지 않는다(8판, `config.mergenotifications` B3). 로드된 실효 설정으로
- [x] 2.6a **RED** — **거부된 편입 블록**의 무관리 보고는 사실 칸 · event key가 「설정 거부」이고 「편입 꺼짐∧미지정」이
      아니다(관측 가능한 것만 단언). **등급은 단언하지 않는다**(Q2(a) 열림). 픽스처에 두 거부 모양 — 편입 켜짐의 범위 밖 pct ·
      꺼짐 · include 없음이지만 범위 밖 pct가 남은 블록 — 을 둘 다 넣는다(9판 r8 N3 · N4, `config.mergeadoption` B3 ·
      `config.adoption.validate` B1)
- [x] 2.6 **RED (Q2 답)** 설정 거부(B3) · `adopt` B2 · B6 · B7 연기분은 normal — outbox 행 0, 연기와 시도 실패는 다른 key. include 지정 시도 실패(B6,
      편입 꺼짐)는 **critical**(10판 정정 — 정본 exit-policy include 동일성, codex P1 → Manager 판정 (가), `review.md` §4.7) (알림 on에 transport 죽음은 Q2(d)로 a124 정본에 이관됨 — 여기 없음, r3 N7)
- [x] 2.7 **RED** — 키 분리(결정 (3)(iii)): exit 관측 자리와 reconcile 자리의 event key가 다르다
- [x] 2.8 **RED** — 전이 상태 무알림 유지: `judgeHoldings` B9(RECONCILE) · B10(묵은 스냅샷)에서 알림 0
- [x] 2.9 **RED** — `notifierAlerter.ExternalPositionFound`의 등급은 normal로 남고, 생산 배선에서
      `IngestExternalPositions`의 알림 어댑터가 nil이다(B12). 2판 6.2a는 이것으로 대체된다
- [x] 2.10 **GREEN** — Q1의 방식대로. `publishBestEffort` · `notifyCritical` · `claimAndDeliver` · `deliver` 본문은
      바꾸지 않는다. **`SeverityOf` · `Notify`의 경계는 Q1 답에 조건부**(r3 N4): (a) 새 종류 등재면 본문 불변,
      (b) `SeverityOf` 계약 변경이면 경계를 다시 선언하고 두 번들을 다시 뽑아 재리뷰한다
- [x] 2.12 **RED** — **연기(normal) → 같은 프로세스의 다음 사이클에서 시도 실패(critical)**: 재시작 없이 critical로
      기록된다. Q2(c)의 두 답(연기 = normal · critical) 모두에서 통과하는 형태로 쓴다. 픽스처에 **같은 사이클에서 앞선
      `adoptOne` 실패 뒤 `adopt` B7(관측 묵음)이 남은 후보를 반환하는 묶음**을 넣는다(8판 r7b F8). 생산 배선(실 `Notifier` · outbox ·
      배달 실행자)으로(r4 R4-1). **6판**: critical은 메모리 래치를 거치지 않는다 · normal은 같은 사실의 반복만 억제한다(r5 R5-1)
- [x] 2.13 **RED** — `adoptOne` 범주 ③(편입 커밋 뒤 exit state 미개설, B3 창 → true)은 critical 요구 밖임을 명명된
      경계로 고정한다 — 후속 후보 `issues.md` I7(r4 R4-3)
- [x] 2.14 **전이 행렬**(r4 codex 권고, Manager 수용) — 「새 critical 0」은 「기존 행 · 래치 0」이 아니다. 알림
      켜짐→PENDING 행 생성→꺼짐, 꺼짐→켜짐, 각 경우 재시작 전후에서: 새 a095 사실의 등급 · 기존 PENDING 행의 존속 ·
      정본 「재시작이 진입 차단을 푸는 우회로가 되어서는 안 된다」의 재차단을 표로 단언한다. **6판(r5)**: 토글은 **로드된
      실효 설정**으로 바꾼다 — 설정 파일만 고친 것은 실행 중 전환의 증거가 아니다. 지연 배달 성공을 운영자 승인으로 세지
      않는다
- [x] 2.15 **RED** — **배달됨 → 재알림 창 경과 → 여전히 시도 실패**(같은 프로세스): 정본 재알림 규칙대로 다시 전송된다.
      메모리 래치가 막지 않음을 단언한다(r5 R5-1). 생산 배선으로
- [x] 2.16 **RED** — **outbox 기록 실패 → 저장소 회복 → 같은 시도 실패**(같은 프로세스): 그 관측에서 기록된다 —
      `ReconcileDriver.alert` B2가 오류를 로그로만 남겨도 다음 관측이 다시 시도한다. 사람 소유 진입 게이트 래치는 이 과정에서
      풀리지 않는다(r5 R5-1). 생산 배선으로
- [x] 2.17 **RED** — **사실 식별자**(r5 R5-2): 같은 등급의 연기 → 시도 실패가 서로 다른 event key로 기록된다(Q2(c)=critical
      가정) · 창 안에서 A → B → A는 **정착한 A**면 A의 재전송을 창 규칙대로 흡수하고 **PENDING인 A**면 재시도한다(둘을 나눠
      시험) · 다른 발송자가 A의 임차를 쥐고 있으면(`ClaimHeldElsewhere`) 그 관측은 배달하지 않는다 · 오류 문구만 바뀐 반복은
      같은 key다(7판 r6 R6-1)
- [x] 2.11 **(Q8 답 — (b) 시점 사건 문구)** **사실이 해소된 뒤의 critical 행**(r3 N3) — 편입 실패 → 다음 사이클 편입 성공 → 지연 배달 ·
      재시작의 시험. 기존 outbox 기계로는 PENDING 행을 갱신 · 정산할 수 없다(design D1). ~~답 전에는 구현 착수
      금지 — 정지 조건~~ → 10판 답: 행 문구가 그 시각의 실패 사건을 말하고(해소 뒤 배달도 참), 편입 성공이 진입 게이트 래치를 풀지 않음을 시험한다

## 3. R2′ — 수량 증가 (결정 (3), design D2)

- [x] 3.1 **RED** — `checkExternalIncrease` B2 **무변화**(R2-B2 삭제): 조회 오류에서 조용히 반환하고
      무관리 알림으로 보내지 않는다
- [x] 3.2 **RED (Q3 답)** 엔진 개설 포지션(`judgeHoldings` B7, 편입 기록 없음)의 수량 증가 검사 — 원장 조정 순증 > 0이면 normal 보고,
      0 · 음수 · 조정 없음이면 보고 없음. 비교 술어 변이(부등호 방향 · off-by-one · 열 바꿔치기) — 10판 구현: 기존 `Journal.PositionAdjustments`
      위의 engine leaf `netAdjustedQuantity`(journal 새 함수 없음, design D2(i) 정정)
- [x] 3.3 **(Q4 답 — normal · 종류 · key 유지 · 최대 수량 래치)** 수량 증가 사실의 종류 · 등급 · 키. critical이면 정본 engine-safety 재알림 창 SHALL NOT과
      대조한 키로, normal이면 `d.grown`(B1) 래치의 기준만
- [x] 3.4 **(Q4 답)** 475150 원장 순서(편입 2 → 3 · 4 · 5 · 8 · 26 · 32) 재생 시험 — 「32가 운영자에게
      전해진다」를 Q4의 답에 맞는 형태로 못 박는다
- [x] 3.5 **GREEN**

## 4. R3 — 총위험 — 보류 (Q5 결정)

- [x] 4.1 **Q5 결정(2026-09-28)**: R3 보류 + exit-policy 델타에서 SHALL 해제(요구사항 삭제) · 후속 change 후보를
      `issues.md` I2에 기록. 2판 3.1~3.6은 이력(`2fbdcd78`의 tasks.md)

## 5. 하지 않는 것을 고정한다 (design D4 · D6)

- [x] 5.1 **RED (§6)** — 평단이 내려간 포지션에서 **자동 경로**(판정 · 관측 갱신 · 복구)의 유효 손절가가 내려가지
      않는다. 운영자 재편입 reset은 이 요구 밖임을 명명한 시나리오로 함께 적는다(델타 4판, 3라운드 V-N5 — 그 하향의
      승인 · audit 여부는 `issues.md` I6)
- [x] 5.2 **RED** — `EvaluateLadder`의 산출 무변화 — rung 잠금가 · 수익률 기준 · R 분모는 계속 `entry_price`에서,
      runner 보호는 관측 워터마크에서, 이전 기준선은 최댓값 합성에 그대로(`ladder.go:391-403`, r3 N6)
- [x] 5.3 `issues.md` I1에 **`baseline_price` 쓰기 자리 넷의 사실**을 번들 분기로 기록 — 판정(B25 `notBelow` ·
      B29 창 선택) · 관측 갱신(B23 — effective 스냅샷과 비교, 스칼라와 일치할 때만 값 유지 · 갈라지면 되돌림) · 재편입 reset(비교 분기 없음) · 최초 INSERT
- [x] 5.4 **(Q6 답 — Manager 판정 2026-09-30)** 래칫 선행 조건을 SHALL로 다시 세우지 않는다 — `issues.md` I1에 두고 후속 change로(design D4)

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0
- [x] 6.2 **§0.3** — exit goroutine에 **새** critical Notify가 없음을 구조로 보인다: `workingSet` B6 경로의
      사실이 normal(2.1). 기존 exit 발신의 `n.mu` 대기는 **a092 21판 소유**(Q7 이관, 2026-09-28) — 대사 goroutine 자신의 대기는 a092 21판 밖의 이름 붙은 잔여
- [x] 6.3 **§0.4** — 새 브로커 조회 0건
- [x] 6.4 **토글 OFF 동등성** (10판 보강: 엔진이 연 포지션의 수량 증가 보고는 `adoption.enabled`와 무관한 새 normal 보고 — 편입 동작이 아니므로 정본 false 동등성의 대상이 아님, 델타 새 요구에 명시. 진입 · 청산 무영향) — 기존 토글 둘(`notifications.enabled` · `adoption.enabled`)의 OFF에서 **등급과 진입 차단
      결과**가 이 change 전과 같다(2.4 · 2.5). normal 래치 키가 사실 식별자로 바뀌어 normal 보고 횟수 · 로그 줄은 달라질 수
      있다. `adoption.enabled=false`에 include 지정이 있는 경로는 Q2(b)로 열려 있다(8판 r7a F5). 새 토글은 도입하지 않는다
- [x] 6.5 `openspec validate --strict`의 한계 — 델타가 ADDED만 쓰는지 확인하고 적는다
- [x] 6.6 FLM · AST **재생성**(구현 후) + `check_analysis.py` 통과
      > 10판: stale 20 재추출 + 새 번들 2, `check_analysis` rc 0 evidence complete(required 6). RED/GREEN/변이 영수증 `review.md` §4.6.
      > 6.2 = exit 루프 무편집 + `TestA095TheExitObserverReportStaysNormalAndKeyedApart` · 6.3 = 새 호출은 원장 읽기(`PositionAdjustments`)뿐 · 6.4 = 2.4 · 2.5 · 6.5 = 델타 두 파일 모두 ADDED 만
- [ ] 6.7 `make sdd-sync` → `make sdd-check`
- [ ] 6.8 **격리 worktree에서** `make gate CHANGE=a095-a-stop-must-know-what-it-covers`
- [ ] 6.9 **독립 리뷰**(구현과 분리된 컨텍스트) · 교차 모델
- [ ] 6.10 PM 동기화 → `openspec archive`

## 7. 배포와 운영 — 사람이 승인한다

- [ ] 7.1 배포 전 `main`과 SchemaVersion 대조
- [ ] 7.2 **배포 직전에 원장을 다시 잰다.** 2026-09-25 재조회(2라운드 P1-6)의 OPEN은 TSLA 먼지 1건이었다.
      배포 직후 critical로 울 것의 예측은 **그 시점의 측정**으로 쓴다 — 2판의 「6건 중 최소 2건」은 거짓이
      되었으므로 지웠다
- [ ] 7.3 **공시** — 편입 켜짐(`adoption.enabled=true`)이고 알림 on인 엔진에서 편입 실패(B5)가 critical이 되고,
      전달 실패 시 진입이 막힌다. 알림 off(거부된 알림 블록 포함) · 편입 off · exclude에서는 막히지 않는다(결정 (2)).
      critical이 메모리 래치를 지나므로 PENDING 동안 대사 쪽 배달이 늘어 exit goroutine의 기존 critical 발신 대기가 늘 수
      있음을 함께 적는다. a092 「모든 보유자」 착지 전에 구현한 경우 **그 창의 증폭을 수용했다는 사용자 확인과 Manager 승인 기록**
      (`review.md` 「착수 승인 기록」)을 인용한다(0.7). transport가 죽은 경우의 교환은 정본 「배달 실행자는 지속 실패를 진입 차단과 운영 모드 승격으로 잇는다」(a124)를 따른다고 함께 적는다(Q2(d) 이관)
- [ ] 7.4 배포 후 `alert_outbox`에 B5 사실의 행이 생기는지 확인

## 선후 관계

```text
a095 (이 change) ── 보호 범위의 보고
   │
   ├─ a092 알림이 손절을 잡지 않는다   **독립**(결정 (1)). exit 관측 자리는 normal로 남는다
   ├─ a124 배달 실행자의 지속 실패     아카이브됨(c1e34dc4). Q2(d) 교환의 정본 소유
   ├─ a091 한 주도 못 판 손절          경계가 같다(「엔진이 보호하기로 했는가」). 등급 표 병합 순서만 확인(1.6)
   ├─ a094 손절이 길을 치운다          독립
   └─ a089 나가지 못한 손절을 센다     독립
```

**a095는 a092와 독립이다.** 2판의 이 자리는 *"a095는 a092 뒤에 간다(1라운드 §1.7 정정)"*였고,
사용자 결정 (1)과 모순된다. 결정 원문(`review.md` §2.16):

> | 1 | **묶지 않는다** — a095 는 a092 와 독립 | 2.10 항목 4 의 ②: exit goroutine 에 critical Notify 를 **새로 두지 않는다**. 발신은 reconcile 쪽(`adoption.go` 경로)에서만 critical, exit 루프 자리(`exitloop.go:518` `alertUnmanaged`)는 normal 로 둔다. 옛 약속 사본(tasks 2.8 · §4 행 · D7) 제거 |

2판이 a092를 선행 조건으로 둔 이유(exit 루프의 관측 전 동기 critical 체류)는 결정 (1)이 그 자리를 normal로
두는 것으로 사라진다(design D1 「exit 관측 자리가 normal이어야 하는 이유」). 남는 것은 기존 exit 발신의 `n.mu`
대기이고, 그것은 순서 의존이 아니라 a092 21판이 소유하는 잠금 규율이다(Q7 이관).
**구현 착수 순서는 이 절이 정하지 않는다** — 그것은 범위가 아니라 스케줄링이고 Manager 사항이다(0.7). 기본 스케줄은 a092
「모든 보유자」 착지 이후이고, 그 **전에** a095를 구현하려면 6판 원칙이 키운 Q7 첫째 면 증폭(손절 루프 `o.alert` 네 자리 — 안전
불변식 4)의 수용에 **사용자 확인**이 착수 전에 있어야 한다(10판 r9 R9-1).
결정 (1)(범위를 a092에 묶지 않음)은 그대로다(9판 r8 N1).

**후속(별도 change)**: 불타기 방향의 손절 상향 래칫 — 선행 사실은 `issues.md` I1, 선행 조건의 SHALL 여부는 Q6.

## 안전 불변식 확인

| 불변식 | 이 change에서 |
| --- | --- |
| §1 사람 승인 없는 LIVE 주문 side effect 금지 | 주문을 내지 않는다. 알림 등급과 검사뿐. 배포는 7절에서 사람이 승인 |
| §2 `mutating: true` 자동 실행 금지 | 준수 |
| §3 토글 OFF는 upstream과 동일 | 알림 off · 편입 off에서 동작 무변화(결정 (2), 6.4) |
| §4 손절 즉시성을 약화·지연하지 않는다 | exit goroutine에 새 critical Notify 없음(결정 (1), 6.2). 기존 발신의 `n.mu` 대기 증가는 Q7 |
| §5 High-risk 경로 | 무보호 보고 · 진입 차단 도달. Pre-Edit 선언은 2.0 |
| §6 보수 방향만 | 손절가를 바꾸지 않는다(5.1). 등급은 결정이 정한 사실에서만 올린다 |
| §7 운영 토글 flip과 live 검증은 사람이 | 7절 |
| §8 시크릿·계좌 개인정보 저장 금지 | 원장 인용은 종목코드·수량·가격·시각까지 |
