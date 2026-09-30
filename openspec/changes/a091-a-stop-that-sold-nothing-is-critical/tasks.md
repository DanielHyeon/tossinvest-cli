# a091 tasks

> **6판(2026-10-01)**: 4라운드 반영(3.3a 머리 · 3.3b 합쳐진 오류 · 3.3c 종료 대기 · 5.3 전체 경과 · 밀린 행 100).
> **5판(2026-10-01)**: 3라운드 R3-1~R3-6 반영(3.0 · 3.2a · 3.2b · 3.3a · 3.3b · 3.4 · 3.8 · 5.1 · 5.3 · 5.4 · 6.8).
> **4판(2026-10-01)**: 2라운드 R2-1~R2-12 · Manager 판정 Q1~Q3 반영으로 2~6 절을 다시 썼다(`review.md` 「4판」).
>
> **High-risk.** 손절 경로의 함수를 편집한다. 다만 **제출 수량 계산은 건드리지 않는다** —
> 바꾸는 것은 보고(등급·종류·문구)뿐이다. proposal-freeze 리뷰(적대적 Eng 필수)가
> 구현 착수 전에 필요하다.
>
> **발효 조건(C1)**: a092 완주(아카이브) 뒤에만 0.x 를 시작한다. 첫 리뷰(2026-08-06)는
> FREEZE 거부였고 2026-09-30 재작성이 그 요구를 반영했다 — 0.3 은 **재리뷰**다.

## 0. 게이트 선행

- [x] 0.0 **[2026-10-01] 확인 — 아카이브 `75d138b5` 가 새 base 의 조상, review 2차 판 0.0.** **a092 아카이브 확인** — C1 발효 조건. exit 관측 goroutine의 critical이
      기록까지만 동기임이 정본이 된 뒤에만 진행
- [x] 0.1 **[2026-10-01] `ec29dc72` → `b30318d6`(`16cb1a1a`, 영수증 `39feef86` · review 2차 판 0.1). `capture_change_base.py` 는 기존 파일을 거절하므로 선례대로 단독 커밋.** `capture_change_base.py --change a091-a-stop-that-sold-nothing-is-critical`
      (base 재고정 — WORKFLOW 「사람 승인 base 재고정」 절차)
- [x] 0.2 **[2026-10-01] 정본 블록 125 줄 기계 복사 + a091 몫 셋(열거 항목 · 0주 문단 넷 · Scenario 넷). 정본 비공백 83 줄 중 delta 에 없는 것 = 열거 첫 줄 1(의도한 치환) — review 2차 판 0.2.** **spec delta 재기저화** — MODIFIED 「등급화된 알림」 블록을 a092 아카이브 뒤
      정본 위에 재작성한다(delta 머리의 재기저화 의무). a091이 더하는 것은 열거 항목
      하나와 0주 문단들뿐, a092 문단은 전부 보존
- [x] 0.3 **[2026-10-01] 번들 14(편집 대상 applyFloor · submit · 등급 SeverityOf + 알림 경로 alert · RecordOnly.Notify + M1 사슬 9) — base AST · 커버리지, calls 표 수치(Floor 읽기 최악 · 기록 경로 busy 5s · 재알림 1h). 하네스 `analysis/harness/write_bundles.py`.** FLM 재검증 — `applyFloor`·`SeverityOf`·(신규) 알림 경로의 AST를 재고정
      HEAD에서 재생성, 좌표 이동 반영. **calls 표의 timeout/retry 칸을 수치로 채운다**
      (첫 리뷰의 방법 교훈)
- [x] 0.4 **[2026-10-01] valid.** `openspec validate a091-a-stop-that-sold-nothing-is-critical --strict --no-interactive`
- [x] 0.5 **[2026-10-01] 5 라운드 — 2라운드 REJECT(A · B · codex) → 4판 → 3라운드(A REJECT · B APPROVE · codex FAIL) → 5판 → 4라운드(A APPROVE · codex FAIL) → 6판 → 5라운드 codex PASS. `review.md` 「2차 판」. Manager 승인 대기.** **proposal-freeze 재리뷰** (적대적 Eng 필수) → `review.md`에 2차 판 추가
- [x] 0.6 **[2026-10-01] 깨끗한 격리 워크트리 @ `6d30d8ea`: a091 번들 15 stale 0 · required 7 은 전부 base 뒤 a112 착지 몫(strategy_* 4 · strategyflow seal census 3) — 게이트 전 재고정(a094 · a095 절차)에서 닫는다.** `check_analysis.py --change a091-…` — FLM 산출물 완결 확인

## 1. 산출물 (완료)

- [x] 1.1 **Function Logic Map** — `ExitObserver.applyFloor` (branches 6, returns 7)
- [x] 1.2 **Function Logic Map** — `SeverityOf` (branches 1, returns 2)
- [x] 1.3 **Branch Test Map** — 위 두 함수. 미테스트 분기 B4·B6 식별
- [x] 1.4 **[2026-10-01] base 재조사 → `issues.md` 「소비자 조사」.** 새 이벤트 종류의 **소비자 조사** — 콘솔 필터·로그 대시보드·`CriticalEvents()`
      호출자. 조사 결과를 `issues.md`에 기록

## 2. 이벤트 종류 신설 (D1) — 4판

- [x] 2.0 **[2026-10-01] `analysis/implementation/pre-edit.md`.** **Pre-Edit 선언** — `internal/obs/event.go`(등급표 한 줄 · 종류 상수 · 목록 주석)
- [x] 2.1 **[2026-10-01] `TestA091TheStopSoldNothingKindIsCritical` — RED `r1-red-obs.log`.** **RED** — `SeverityOf(EventExitStopSoldNothing)` = critical
- [x] 2.2 **[2026-10-01] `TestA091NoOtherGradeMoved`.** **RED** — 기존 **19**종(base `event.go:337-361`)의 등급 무변화 · 미등록 종류는 normal
- [x] 2.3 **[2026-10-01] 같은 시험 + class rule 셋 통과.** **RED** — 값 문자열 핀 `"exit.stop_sold_nothing"` · subject `exit` · `CriticalEvents()` 등재 — class rule 셋(measurement · a074 · a109) 통과
- [x] 2.4 **[2026-10-01] `event.go`.** **GREEN** — 종류 추가 + `criticalEvents` 등록 + 목록 주석. `EventExitProposalCapped` 등급 · 값 무변화

## 3. 0주 보고 (D1 게이트 · D2 · D3 · D7 · D8) — 4판

> 하네스 둘: (가) 기존 `newExitHarness`(가짜 알림 수집기 — 종류 · 필드 · 호출 수), (나) **실제 `obs.RecordOnly` + 실제 원장 + 로그 캡처**
> (`a092_exit_cycle_records_only_test.go` 의 `a092RecordOnlyHarness` 모양) — 행 · 등급 · 게이트 · 로그 줄. RED 는 둘 중 단언이 사는 쪽에서.

- [x] 3.0 **[2026-10-01] `pre-edit.md`.** **Pre-Edit 선언** — `ExitObserver.applyFloor` · `ExitObserver.submit` · `ExitObserverOptions`(+`NotificationsEnabled`) ·
      `Context.ExitObserver` 생산 배선(설정 값으로 덮기) · `obs.Notifier.escalate`(로그 두 줄의 계좌 필드만 — D8). High-risk — 제출 수량 ·
      시점 무변경이 경계
- [x] 3.1 **[2026-10-01] `TestA091AProtectiveZeroIsACriticalRow`.** **RED (나)** — 알림 켜짐 · 보호 · 하한 0(Bound Sellable) → 새 종류 critical · outbox 행 1 · 키 `exit.stop_sold_nothing|<pos>` ·
      본문에 원인 범주 + 관측 시각(UTC) · 한국어 · `이름(코드)`
- [x] 3.2 **[2026-10-01] `TestA091AFloorThatCannotBeComputedIsTheSameReport`.** **RED (나)** — 알림 켜짐 · 보호 · 하한 계산 실패(B2) → **같은 종류 · 등급 · 키**, 원문 오류 문자열은 제목 · 본문 · payload 에 없음
- [x] 3.2a **[2026-10-01] `TestA091WithAlertsOffNothingChangesButTheWording`.** **RED (나)** — 알림 **꺼짐** · 보호: ① B2 → **알림 0**(Notify 호출 0) · 로그 한 줄(옛 종류 · 계좌 없음), ② 끝 → 옛 종류 normal 알림 하나.
      둘 다 outbox 행 0 · 게이트 사유 없음 · 모드 무변화(불변식 3 — base 와 같은 호출 수)
- [x] 3.2b **[2026-10-01] `TestA091TheProductionAssemblyReadsTheLoadedSwitch`.** **RED (배선 하네스 — `a092_exit_record_only_wiring_test.go` · `a092_export_test.go` 의 `OptionsForTest` 모양)** — 생산 배선이
      로드된 설정의 `notifications.enabled` 로 옵션을 **덮는다**(호출자 값 무시) — 참 · 거짓 두 팔
- [x] 3.3 **[2026-10-01] `TestA091APartialCapIsUnchanged`.** **RED (가)** — 보호 · **부분** 캡 → 옛 종류 · 등급 · 문구 **무변화**(기존 `TestTheConfirmedFloorCapsTheLiquidation` 단언 유지)
- [x] 3.3a **[2026-10-01] `TestA091AHoldingsBoundZeroMeansTheAccountHoldsNone` · `TestA091AZeroHoldingIsNotAFailedStop`.** **RED** — 보유 0 판정: `Bound == FloorBoundHoldings ∧ Quantity == "0"` ⇒ 신선한 보유 0(반대 방향은 조건부) — `riskcalc.ConfirmedFloorQuantity` 표
      시험으로 **한 방향(⇒)을 정확히, 반대 방향(⇐)은 조건부로** 핀(design D3 ③ 5판): 신선 보유 0 × 매도가능 {신선 0, 신선 양수, 없음, 낡음}
      × 로컬 {0, 양수, 비정상}, 보유 양수 × 같은 칸 — Holdings 한정 0 은 「신선 보유 0 ∧ 신선 매도가능 ∧ 유효 로컬」에서만. 알림 켜짐 · 보호 ·
      보유 0(Holdings 한정) → 옛 종류 normal · 본문 「계좌에 보유가 없다」. **새는 칸 셋**(매도가능 조회 실패 · 보유 스냅숏 낡음 · 로컬 오류)은
      critical 로 남음을 단언(과보고 방향 — 이름 붙인 잔여)
- [x] 3.3b **[2026-10-01] `TestA091OnlyACancellationOnlyFailureIsSuppressed` 팔 8 · `TestA091TheCancellationPredicate` — (vii) 은 생산 클라이언트의 오류 모양(`client.go` doRequest 의 `%w: %s`)으로, httptest 공식 클라이언트 왕복은 아님(이름 붙인 차이).** **RED (나)** — 종료 취소는 **출처**로(design D5 6판): (i) 하한 오류가 **취소뿐**(잎이 전부 `context.Canceled` — `%w` 감쌈 ·
      `*url.Error` 감쌈 포함)이고 ctx 끝남 → 알림 0 · 행 0 · 래치 0 ·
      승격 0, (ii) 진짜 하한 오류가 돌아온 **뒤** ctx 취소 → ① 보고 · 행 1 · 가짜 래치 0(`WithoutCancel` 기록), (iii) 원인 판정 **뒤** · 기록 **전** ctx
      취소(주입 지점) → 행 1 · 가짜 래치 0, (iv) ctx 살아 있고 HTTP 시한(`DeadlineExceeded`) → ①, (v) **합쳐진 오류** — 인증 거절 +
      승격 원장 연산의 취소(`errors.Join`, 생산 `Retrier.Query` 모양) → ① 보고(억제 안 함), (vi) 재시도 대기 중 취소 → 직전 일시 오류 → ① 보고,
      (vii) **생산 공식 클라이언트**(httptest 서버가 응답을 멈춘 채 ctx 취소)로 HTTP 중 취소 → 잎 `ErrTransport` → ① 보고 · 가짜 래치 0(과보고 잔여 핀)
- [x] 3.3c **[2026-10-01] `TestA091ShutdownWaitsForTheReportWithoutADeadline`.** **RED (나) — 종료 중 보고 대기(D5 6판)** — 기록을 막아 둔(원장 쓰기 대기 주입) 상태에서 루프 ctx 취소: 루프는 기록이 풀릴 때까지
      돌아오지 않고, 풀린 뒤 행 1 · 가짜 래치 0 으로 돌아온다 — 교착 0, 대기가 기한 없음이라는 사실을 시험 이름 · 주석에 적는다
- [x] 3.4 **[2026-10-01] `TestA091TheProtectiveSplitCoversEveryOrderableAction`(AST 열거 5) · `TestA091ATakeProfitThatSoldNothingKeepsItsGrade`(익절 끝-끝 둘) — 5 액션 전부를 `submit` 끝-끝으로 돌리지는 않음(보호 1 · 익절 1 끝-끝 + 술어 표 5, 이름 붙인 차이).** **RED (가)** — 표 시험: 주문 액션 5종(보호 2 · 익절 3 — `ratchet.go:94-118`)을 `submit` 을 거쳐 두 원인 × 알림 켜짐으로 —
      보호만 새 종류, 익절은 어떤 원인이든 옛 종류 normal(익절 B2 는 알림 0 · 로그 옛 종류). 새 액션이 생기면 이 표가 깨지게 `exitpolicy`
      패키지의 `Action` 상수를 **AST 로 열거**해 `Orderable()` 인 것의 집합과 표의 행 집합이 같음을 단언(`Orderable()` 은 술어라 열거가 아니다 — 3라운드 보이스 B)
- [x] 3.5 **[2026-10-01] `TestA091TheOutcomeIsUnchanged`.** **RED (§0.3 · §0.9 회귀 — 이 change 의 안전 경계)** — 위 전부에서 관측 결과 무변화: 제출 수량 · 제출 수 · B2/0주 뒤 레벨 해제
      (`Pending()==false`) · 하한이 풀린 다음 관측의 같은 레벨 재발의(기존 두 시험을 확장)
- [x] 3.6 **[2026-10-01] `exitloop.go` · `exit_stop_sold_nothing.go`.** **GREEN** — `submit` 이 `isProtective(proposal)` 를 `applyFloor` 에 넘기고, `applyFloor` 는 반환값을 바꾸지 않은 채 원인 분류(D3) ·
      게이트(D1)로 종류를 고른다. `exitpolicy` · `riskcalc` 무편집
- [x] 3.7 **[2026-10-01] `TestA091TheLogAndTheAlertAreOneKind`.** **RED (나) — H2** — 보호 · 알림 켜짐 · 두 원인: 그 사건의 로그 줄(B2 오류 줄 · 기록 줄)과 알림이 같은 종류. 게이트 밖이면 둘 다 옛 종류
- [x] 3.8 **[2026-10-01] `TestA091NoLineOrRowCarriesTheAccount` · `TestA091TheEscalationLinesCarryNoAccount`.** **RED (나) — 계좌 카나리(D8 5판)** — 계좌 sentinel 로: B2 오류 줄(보호 · 익절 · 게이트 밖) · 기록 줄 · **기록 실패 줄** · **`escalate` 성공 ·
      실패 줄**(`internal/obs` 시험) · 행(제목 · 본문 · payload) 어디에도 sentinel 없음. 기록 실패 팔은 원장 쓰기 실패 주입으로
- [x] 3.9 **[2026-10-01] `TestA091AnAuthRefusalOnTheFloorReadRecordsTwo`.** **RED (나) — 겹침(D5 범위 표)** — 하한 조회가 401 이면 한 `applyFloor` 호출에서 모드 통지 기록 1 + 새 종류 기록 1(둘 다 critical 행),
      그 밖의 기록 0
- [x] 3.10 **[2026-10-01] `TestA091TheRowKeepsTheFirstCauseAndTheLogKeepsEach`.** **RED (나) — 원인 계약(D7)** — 한 에피소드에서 B2 → 끝(Sellable), 그리고 끝 → B2 두 순서: 행 1 · 본문 = 첫 원인, 로그 줄 = 관측마다 그 관측의 원인

## 4. 문구 (D4) — 4판

- [x] 4.1 **[2026-10-01] 3.1 · 3.2a 의 고정 제목 단언 + 「일부」 부재.** **RED** — 0주 문구(보호 새 종류 · 보호 게이트 밖 옛 종류 · 익절 옛 종류)가 **참인 결과를 말한다**: 제출 수량 0 을 명시하고
      「일부」 계열 문구가 없다 — 금지어 하나가 아니라 문장 단언(제목 · 본문 고정 문자열)
- [x] 4.2 **[2026-10-01] `TestA091APartialCapIsUnchanged` 고정 본문.** **RED** — 부분 캡 문구 **무변화**(고정 문자열)
- [x] 4.3 **[2026-10-01] `exit_stop_sold_nothing.go` `zeroFloorEvent`.** **GREEN**

## 5. 실측 재생 · 비용 — 4판

- [x] 5.1 **[2026-10-01] `TestA091TheAugustSecondReplay` — 결과 `issues.md` 「8/2 재생 결과」.** **2026-08-02 재생**(원장 재독의 모양 — 보유 5 · 매도가능 0 · 13 관측 / 3분, `design.md` 「8/2 원장 재독」) — 하네스 (나) + **배달 실행자**
      (생산 조립 그대로 `(&engine.Context{Journal, Entry, Notifier, AccountRef}).AlertDeliverer(clk)` — 선례
      `a124_the_production_executor_latches_test.go:54` · 같은 `EntryGate` 공유, 내보내기 훅 불요), 팔 넷:
      (i) 알림 켜짐 · 정상 전송 → 행 1 · 발송 1 · 정착 · 13 관측 중 재발송 0,
      (ii) 알림 켜짐 · 전송 실패 → 행 1 PENDING · 시도 계수 · 한도에서 래치 · ENTRY_BLOCKED 승격(의도된 a092 의미론),
      (iii) 알림 켜짐 · publisher 없음 → (ii)와 같은 판정 + 배달 실행자의 `alert_undelivered` 「no publisher」 줄이 사이클마다,
      (iv) 알림 꺼짐 → 행 0 · 래치 0 · 승격 0. 각 팔의 행 수 · 발송 수 · 래치 · 모드 · 로그 줄 수를 기록
- [x] 5.1a **[2026-10-01] `TestA091TheReminderWindowDecidesTheNextEpisode`.** **재알림 창 경계** — 정착(전달 · 승인) 행 뒤 같은 키: 1h 안 재무장 0, 1h 지나 재무장 1(본문 교체 — a097)
- [x] 5.2 **[2026-10-01] `issues.md` 「8/2 재생 결과」.** 결과를 `issues.md` 에 — 첫 리뷰 H3(`MarkAlertDelivered` PENDING ERROR 12줄)은 옛 동기 발송의 모양이었고 base 에는 그 발생원이 없으며,
      대신 (iii)의 배달 실행자 줄이 난다는 것을 **잰 수**로
- [x] 5.3 **[2026-10-01] `TestA091TheReportFitsItsShare` — 소유 칸 11~26ms(실패 승격 <0.12ms), 승인 칸 1.06~1.28s < 5s(측정 · 기록).** **몫 실측 · 수락(D5 7판 — Manager 판정 2026-10-01)** — 보고 호출 하나의 전체 경과를 네 칸에서 잰다. **a091 소유 칸**(경합 없는 원장 ·
      기록 실패 승격 · 연결 풀 경합)은 관측 최악 ≤ 750ms(넘으면 멈춤 · 보고). **승인 경합 칸**(밀린 행 100)은 측정 · 기록이고 판정은 관측 주기
      5s 미만 하나(근거 design D5). 결과를 design D5 에 적는다
- [x] 5.4 **[2026-10-01] `TestA091ALaterStopStillGoesOutInTheSameCycle/750ms` · `/1.3s`.** **뒤쪽 보호 포지션(D5 수락 (ii))** — 한 사이클에 보호 0주 포지션 → 보호 포지션 순서, 앞 포지션의 기록에 **750ms(배정)와 1.3s(실측 최악)**
      지연을 각각 주입:
      뒤 포지션이 그 사이클에 판정 · 제출되고 시세가 쓸 수 있는 상태(15s 수명 안)임을 단언

## 6. 게이트

- [ ] 6.1 `go test ./... -count=1 -race` 회귀 0(격리 worktree, 게이트와 같은 커밋)
- [x] 6.2 **[2026-10-01] 3.5 + D5 편성 + 5.3 실측.** §0.3 확인 — 제출 수량 · 제출 수 · 레벨 해제/재발의 무변화는 3.5 가, 루프 몫은 D5 편성 + 5.3 실측이 보인다(diff 한 줄로 갈음하지 않는다)
- [x] 6.3 **[2026-10-01] `applyFloor` · `submit` 번들 calls 표 편집 전후 대조 — 브로커 호출은 `ConfirmedFloor` 하나 그대로.** §0.4 확인 — **a091 이 더하는 브로커 요청 0**: `applyFloor` · `submit` 번들 calls 표를 편집 전후로 대조(RECONCILE 의 하한 읽기
      `exitwiring.go:207` · `:231` 은 그대로)
- [ ] 6.4 `make sdd-sync` 재실행 → `make sdd-check`
- [ ] 6.5 **격리 worktree에서** `make gate CHANGE=a091-a-stop-that-sold-nothing-is-critical`
- [ ] 6.6 독립 검증 (구현과 분리된 컨텍스트) · 교차 모델
- [ ] 6.7 PM 동기화 → `openspec archive`
- [x] 6.8 **[2026-10-01] `docs/operations.md` 「손절이 한 주도 나가지 않았을 때」.** **운영 문서** — `docs/operations.md` 에 `exit.stop_sold_nothing` 절(무엇이 났나 · 확인 · 사람 조치 · 승인 `tossctl engine alerts ack` ·
      모드 해제 `tossctl engine mode-release` — 둘 다 `mutating: true`, 에이전트 자동 실행 금지) + **ntfy 구독 필터 주의**: ntfy 는
      `event_type` 을 `Tags` 머리에 싣는다(`internal/obs/ntfy.go`) — `exit.proposal_capped` 로 거르던 운영자 필터는 보호 0주를 더 이상 못 본다

## 선후 관계

| change | 관계 |
| --- | --- |
| **a092 (알림은 손절을 붙잡지 않는다)** | **선행 — C1 발효 조건.** a091의 새 critical은 a092가 세운 기록 계약(기록까지 동기·임차 없음·재알림 창) 위에 선다(design D5). a092 완주 전 착수 금지 |
| a089 (outbox 재발 장부) | **불구현 아카이브(2026-09-28 사용자 결정).** 종전 판이 여기 두었던 재발·접힘 질문은 a092 재알림 창·episode가 소유한다 — 5.1·5.2가 그 의미론으로 확인한다 |
| a090 (관측 누락 계측) | 독립 |
| a087 (보호 청산의 가격) | 독립. a087은 "무엇을 보내는가", a091은 "0주였을 때 누가 아는가" |
