# a095 · tasks — 3판

- **Change**: `a095-a-stop-must-know-what-it-covers`
- **위험 등급**: **High-risk** — 무보호 보고의 등급과 진입 차단 도달. §0.3 적용.
- **base-commit**: `027163575e78482cdd4d8129f027e928b769ae87` (3판 재고정 `0b17784e`)
- 결정 (1)~(3) 원문은 `proposal.md` §0. `[비움 — Qn]`은 결정이 덮지 않아 Manager 에게 올린 질문이다
  (`proposal.md` 「열린 질문」). 비운 task는 답이 오기 전에 착수하지 않는다.

## 0. 게이트 선행

- [x] 0.1 `base-commit.txt` 고정 — **3판 재고정** `ec29dc72` → `02716357`(`0b17784e`). a095 디렉터리를 만진
      커밋 중 `.go`를 고친 것은 `a30eb35a` 하나이고 새 base 앞이다
- [x] 0.2 `openspec validate a095-a-stop-must-know-what-it-covers --strict` — 3판에서 다시 통과(`review.md` 3판 기록)
- [x] 0.3 **AST 산출물이 문서보다 먼저** — 3판 번들 21개(새로 12 · 다시 뽑음 4 · 해시 일치로 산문만 5).
      생성기 `analysis/harness/render_bundles.py`, 커버리지 `analysis/harness/coverage/`
- [x] 0.4 `check_analysis.py --change a095-…` — 3판 판정은 `review.md` 3판 기록(격리 worktree, 완료 커밋 기준)
- [ ] 0.5 **proposal-freeze 리뷰**(적대적 Eng 필수) → `review.md`. 교차 보이스 · 교차 모델을 여기서 지킨다
      > **2라운드(2026-09-25) FAIL**(`review.md` 「2라운드」). **사용자 결정(2026-09-25)**: §2.11 1 **독립** ·
      > 2 **거부** · 3 **수용**(범위 이동). 3판이 그것을 반영했다. 3라운드는 Manager 가 따로 지시한다
- [ ] 0.6 **열린 질문 Q1~Q7의 답** — 답이 온 뒤 비운 절을 채우고 3라운드 전에 다시 뽑을 번들을 정한다
      > **파킹(2026-09-27, Manager 분류 — `review.md` §3.8)**: 사용자행 Q5 · Q2(d) · Q7, 구현 로트행 Q1 · Q2(a)(b)(c) ·
      > Q3(정지 조건) · Q4 · Q6. 교차 모델은 Codex 401 가능성 — 실행 시점에 확인
      > **사용자 결정(2026-09-28)**: Q5 보류 + 델타 SHALL 해제 · Q2(d) → a124 정본 · Q7 → a092 21판(둘째 면은 이름
      > 붙은 잔여). **남은 열림은 구현 로트행 Q1 · Q2(a)(b)(c) · Q3 · Q4 · Q6** — 착수 시 코드 영수증으로 확정

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
- [ ] 1.6 **a091과의 병합 확인** — 둘 다 등급 표를 건드릴 수 있다. 3판은 표 크기를 시험에 적지 않는다

## 2. R1′ — 등급을 사실로 (결정 (1)·(2), design D1)

- [ ] 2.0 **Pre-Edit 선언** — 대상은 Q1의 답이 정한다
- [ ] 2.1 **RED** — exit 관측 자리(`workingSet` B6 → `ExitObserver.alertUnmanaged`)의 사실은 **normal**이고
      `Notify` B1 창(`publishBestEffort`)으로 간다: outbox 행 0 · `n.mu` 미획득. 기존
      `TestAPositionWithNoEntryDecisionIsSkippedAndAlertedOnce`의 normal 단언을 유지한다
- [ ] 2.2 **RED** — reconcile `alertUnmanaged` B5(편입 켜짐 · 편입 실패)의 사실은 **critical**이고 outbox 행을
      만든다. 싣는 방식은 `[비움 — Q1]`
- [ ] 2.3 **RED** — exclude(`judgeHoldings` B11 · `alertUnmanaged` B4)는 **normal** — outbox 행 0 · 진입 게이트
      래치 0 · 운영 모드 승격 0
- [ ] 2.4 **RED** — `adoption.enabled=false` ∧ 미지정(`judgeHoldings` B12 · 기본 사유)은 **normal** — 2.3과 같은
      단언. 정본 exit-policy 「false에서의 동작은 무관리 보유 알림을 포함한 기존 동작과 동일」의 동등성 시험
- [ ] 2.5 **RED** — **알림 off**(`notifications.enabled=false`) 엔진에서 B5 사실이 발생해도 `deliver` B3 ·
      `notifyCritical` B4 사슬에 닿지 않는다 — 진입 게이트 래치 0 · 승격 0. 방식은 `[비움 — Q1]`
- [ ] 2.6 `[비움 — Q2]` 설정 거부(B3) · include 지정 시도 실패(B6) · `adopt` B2 · B6 · B7 연기분 · 알림 on에
      transport 죽음 — 답에 따라 RED를 쓴다
- [ ] 2.7 **RED** — 키 분리(결정 (3)(iii)): exit 관측 자리와 reconcile 자리의 event key가 다르다
- [ ] 2.8 **RED** — 전이 상태 무알림 유지: `judgeHoldings` B9(RECONCILE) · B10(묵은 스냅샷)에서 알림 0
- [ ] 2.9 **RED** — `notifierAlerter.ExternalPositionFound`의 등급은 normal로 남고, 생산 배선에서
      `IngestExternalPositions`의 알림 어댑터가 nil이다(B12). 2판 6.2a는 이것으로 대체된다
- [ ] 2.10 **GREEN** — Q1의 방식대로. `SeverityOf` · `Notify` · `publishBestEffort` · `notifyCritical` ·
      `claimAndDeliver` · `deliver` 본문은 바꾸지 않는다

## 3. R2′ — 수량 증가 (결정 (3), design D2)

- [ ] 3.1 **RED** — `checkExternalIncrease` B2 **무변화**(R2-B2 삭제): 조회 오류에서 조용히 반환하고
      무관리 알림으로 보내지 않는다
- [ ] 3.2 `[비움 — Q3]` 엔진 개설 포지션(`judgeHoldings` B7, 편입 기록 없음)의 수량 증가 검사 — 비교 기준이
      정해지면 RED를 쓴다
- [ ] 3.3 `[비움 — Q4]` 수량 증가 사실의 종류 · 등급 · 키. critical이면 정본 engine-safety 재알림 창 SHALL NOT과
      대조한 키로, normal이면 `d.grown`(B1) 래치의 기준만
- [ ] 3.4 `[비움 — Q4]` 475150 원장 순서(편입 2 → 3 · 4 · 5 · 8 · 26 · 32) 재생 시험 — 「32가 운영자에게
      전해진다」를 Q4의 답에 맞는 형태로 못 박는다
- [ ] 3.5 **GREEN**

## 4. R3 — 총위험 — 보류 (Q5 결정)

- [x] 4.1 **Q5 결정(2026-09-28)**: R3 보류 + exit-policy 델타에서 SHALL 해제(요구사항 삭제) · 후속 change 후보를
      `issues.md` I2에 기록. 2판 3.1~3.6은 이력(`2fbdcd78`의 tasks.md)

## 5. 하지 않는 것을 고정한다 (design D4 · D6)

- [ ] 5.1 **RED (§6)** — 평단이 내려간 포지션에서 유효 손절가가 내려가지 않는다
- [ ] 5.2 **RED** — `EvaluateLadder`의 산출 무변화 — 모든 선은 계속 `entry_price`에서 나온다
- [x] 5.3 `issues.md` I1에 **`baseline_price` 쓰기 자리 넷의 사실**을 번들 분기로 기록 — 판정(B25 `notBelow` ·
      B29 창 선택) · 관측 갱신(B23 거절, 값 무변화) · 재편입 reset(비교 분기 없음) · 최초 INSERT
- [ ] 5.4 `[비움 — Q6]` 래칫 선행 조건을 SHALL로 다시 세울지

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0
- [ ] 6.2 **§0.3** — exit goroutine에 **새** critical Notify가 없음을 구조로 보인다: `workingSet` B6 경로의
      사실이 normal(2.1). 기존 exit 발신의 `n.mu` 대기는 **a092 21판 소유**(Q7 이관, 2026-09-28) — 대사 goroutine 자신의 대기는 a092 21판 밖의 이름 붙은 잔여
- [ ] 6.3 **§0.4** — 새 브로커 조회 0건
- [ ] 6.4 **토글 OFF 동등성** — 기존 토글 둘(`notifications.enabled` · `adoption.enabled`)의 OFF에서 동작이
      이 change 전과 같다(2.4 · 2.5). 새 토글은 도입하지 않는다
- [ ] 6.5 `openspec validate --strict`의 한계 — 델타가 ADDED만 쓰는지 확인하고 적는다
- [ ] 6.6 FLM · AST **재생성**(구현 후) + `check_analysis.py` 통과
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
      전달 실패 시 진입이 막힌다. 알림 off · 편입 off · exclude에서는 막히지 않는다(결정 (2)). transport가 죽은 경우의 교환은 정본 「배달 실행자는 지속 실패를 진입 차단과 운영 모드 승격으로 잇는다」(a124)를 따른다고 함께 적는다(Q2(d) 이관)
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
