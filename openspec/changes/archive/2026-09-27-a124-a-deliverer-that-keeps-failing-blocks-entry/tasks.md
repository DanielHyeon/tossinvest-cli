## 0. 계약과 증거

- [x] 0.1 `base-commit.txt` 고정 — proposal-freeze 직전 (`capture_change_base.py`) → `4798d399` (2026-09-26; AST 추출 `463cc895` 와 Go diff 0)
- [x] 0.2 `openspec validate a124-a-deliverer-that-keeps-failing-blocks-entry --strict` 통과 (openspec 1.4.1, rc 0, 2026-09-26)
- [x] 0.3 **AST 산출물이 문서보다 먼저** — 함수 5개(`cycle` · `deliverOne` · `Notifier.deliver` · `Notifier.notifyCritical` ·
      `restoreAlertEntryLatch`), 분기 48, HEAD `463cc895` (2026-09-25, Manager). 편집 대상 둘은 편집 뒤 재추출(1.3)
- [x] 0.4 proposal-freeze 리뷰(**적대적 Eng 필수 + 교차 모델**) → `review.md`. 열린 결정 D2 · D3 를 답한다
      — 1판 실행(2026-09-26, Teammate 적대 Eng + codex): **REJECT**. D2 = ㄱ(3) · D3 = 센다로 답함. F1·F2·F3·F4·F8 과
      Manager 결정 M1·M2 를 design 에 반영한 뒤 재freeze (`review.md` §0)
      — 12회차(2026-09-27 codex): **REJECT**, AC1 = freeze 차단(Manager §0.15) → 7판(논거 교정 + 운영 효과 선행 조건)
      — 13회차: **REJECT**, AD1 · AD2 P1(7판 신설 문장 정밀도) → 8판(Manager 승인 §0.18)
      — **14회차: PASS**(codex gpt-6-astra, session `01a0e2e8-6156-7ea2-abe9-552bdc7809cb`, 2026-09-27 21:52 KST, `codex-r14-output.md` 「VERDICT: PASS」) +
      **Manager 독립 검증**(PASS 원문 · 커밋 경로 청정 · `openspec validate --strict` rc 0) → 체크(2026-09-27, Manager 승인). AE1~AE3 는 D10 정렬로 반영(재리뷰 불요,
      Manager 판정 §0.20). **freeze 계약으로 남는 것**: W1(b) 미해소 — B 의 알림 사유 보호는 (i) a092 의 `Acknowledge` 셈~해제 원자화(「셈과 해제 사이 유입」 궤적만)
      또는 (ii) 투영기 배선 + 모드 기동 복원(AC2 수리 · 진입 허용 전 복원 순서 동반)이 각자의 궤적 범위에서만 세운다(design D10); 「운영 효과 선행 조건」 — 모드의
      진입 집행은 `SetModeProjector` 생산 배선 + 기동 `RestoreOperatingModeProjection` 착지 뒤에만 참(소유 a092 축소판, 코드 착지 조건 아님, proposal)

## 1. 증거와 Pre-Edit

- [x] 1.1 CodeGraph: `deliverOne` · `cycle` · `PendingAlerts` · `EntryGate.Block` · `EscalateOperatingMode` 의 callers/callees →
      `analysis/code-context/` (codegraph 1.6.0; CGC kuzu 잠금·GBrain busy 로 not-applicable, 불일치 R1–R5 는 HEAD·AST 로 해소, R6 미해소 = F1)
- [x] 1.2 Pre-Edit 선언(High-risk): 대상 심볼 · 호출부 · 기존 시험 · 불변식 · 실패 시험 · rollback — 대상에 **`execgw.EntryGate`**(`Clear` ·
      새 `ClearEpoch` · `BlockUnlessClearedSince`)를 포함하고, `Clear(ReasonAlertUndelivered)` 호출자 둘과 `revision` 소비자(전략 봉인)를 다시 잰다
      — review §1 「Pre-Edit Gate」. ⚠ 선언문은 편집 **뒤**에 적었다(증거 입력은 편집 전 — 과정 이탈 기록, Manager 수용)
- [x] 1.3 편집 대상 함수 `revision: current` 재추출, Branch Test Map 재번호(옛/새 ast diff 정렬) — 대상 셋:
      `alertDeliverer.cycle` · `alertDeliverer.deliverOne` · `Journal.settleUnderClaim` · `EntryGate.Clear`
      (뒤의 둘은 base 번들 2026-09-26; `Notifier.Acknowledge` 는 5판에서 편집 대상이 아니다).
      `Journal.PendingAlerts` 도 편집 대상이면 base 번들을 **편집 전에** 만든다
      — 편집 번들 7(`alertDeliverer.release` · `EntryGate.Block`(줄만 이동) 포함) 재추출 + 편집 전→후 분기 대응(difflib, `analysis/harness/flm_a124.py`).
        `PendingAlerts` 는 편집하지 않음(새 메서드)
- [x] 1.4 Branch Test Map 의 기존 시험 실측 — `go test -covermode=set` 으로 분기 도달을 재고 파일 이름을 함수 이름으로
      — `analysis/harness/branch_coverage.py` 시험별 실행, 원장 `coverage-*.json`(엔진 493 · journal 158 · execgw 59 · obs 81)

## 2. RED

- [x] 2.1 한도 도달 → `Gate.Block(ReasonAlertUndelivered)` + `EscalateOperatingMode(CRITICAL_ALERT_UNDELIVERED)`,
      **동기 경로가 한 번도 시도하지 않은 구성**(publisher 응답 없음 · publisher 없음 둘 다)
- [x] 2.2 승인으로 미전달 0 → 차단 해제, 모드는 원장에 남는다(모드 단언은 2.14 (a) 규칙)
- [x] 2.3 한도 도달 뒤 재시작 → 기동 복원이 다시 잠그고 모드는 원장에 있다 — 이 시험의 차단은 **PENDING 복원**이 만든다(모드가 아니다, AC1); PENDING 0 변형은 2.14 (c)
- [x] 2.4 굶주림: 한도 행 batch 개 + 새 critical 행 → 다음 사이클이 새 행을 먼저 시도한다
- [x] 2.5 한도 행은 잔여 자리로 밀려도 사라지지 않고 미전달 수에 세어진다
- [x] 2.6 손절 즉시성 불변 — `a098_the_backlog_does_not_delay_protection_test.go` 그대로 초록(보존 시험) + 둘로 나눈 변형(Y4):
      (a) **격리 단언**: 실행자가 `g.mu` 를 기다리게(전략 dispatch 모사로 `g.mu` 보유) 하고 원격 발행을 멈춰도, exit `Notify` · 비상 청산
      `Notify` 가 `n.mu` 에서 실행자를 기다리지 않음(실행자는 `n.mu` 를 잡지 않는다 — 구조 + 행동). (b) **원장 연결 경합 측정**: 실행자의 정산 ·
      승격 트랜잭션 **안**에 지연을 주입(연결을 쥔 채)하고 exit 사이클 체류를 잰다 — 수락선 = 같은 일을 한(judged 수 동일) 무실행자 기준선
      + **고정 여유** `a098ExitCycleDwellMargin`(250 ms, a098 시험의 상수를 인용 — 측정값에 따라 늘어나는 예산이 아니다, AA3). 주입 지연을 그 여유보다
      크게 잡은 변형에서 **빨강**이 나는지도 확인해 계측기가 눈멀지 않았음을 보인다(Z5: 사이클 전체에 걸친 반복 인터리빙); publisher 없음(F12)도 같은 방식.
      (c) **backlog 크기**(AB1): PENDING P = 10 · 1 000 · 10 000 에서 `PendingAlertsForDelivery` 의 연결 점유(bench, `a098_cycle_cost_bench_test.go` 확장)와
      실행자가 도는 동안의 exit 사이클 체류 — 같은 고정 여유. 넘으면 멈추고 보고(대안 = 가산 색인 = 스키마 변경, Manager 결정)
- [x] 2.7 판정 입력(D1, F1) — 가산 읽기 원자성(Z4: 트랜잭션 안 SELECT · 커밋 실패 주입 → 저장 상태 불변 · 결과 사용 불가, 세 호출자): 나열 뒤 다른 발송자가 attempts 를 올린 행 → 판정은 커밋된 값으로; 재무장된 행은 0 부터;
      기록 오류의 `SettleResult{}` 가 「적용됨」으로 새지 않는다(F10)
- [x] 2.8 발송 중 승인(D7): 전송 중 승인 → 기록 `AlreadySettled` → 잠금·승격 없음. `LeaseLost` 도 없음(재무장 인터리빙 포함).
      실패 기록의 `NotFound`·모르는 값은 잠금만(승격 없음, N6) — 원칙 E 적용 단계를 거쳐도 승격 없음(W4)
- [x] 2.9 원칙 E(D7, spec 「늦은 적용은 제때 적용과 같아야 한다」) — 결정적 인터리빙(세대 읽기·울타리 직전 훅). 모든 단언은 **게이트 사유와
      durable 운영 모드 둘 다**(모드 완화 승인 없음; 모드 단언은 원장 행으로 — 2.14 (a)). 세대 읽기 창(AB2): 「세대 읽기 → 해제 → 적용」 = 차단 없음, 「정산 → 해제 → 세대 읽기 → 적용」 =
      보수적 재잠금(허용 예외)을 각각 단언:
      (a) 해제가 근거 확정 **앞**(옛 승인의 `Clear` → 새 행 B 한도, ㉬) → 차단 + 승격;
      (b) 해제가 **뒤**(한도 커밋 → 남이 전달 → 운영자 빈 목록 승인·해제, ㉫) → 차단 없음 + **승격 있음**, 그리고 같은 시험에서 대조 경로
      「한도 커밋 순간 `Block` + 승격 → 같은 해제」의 게이트·모드 상태와 **같음**을 단언(등가성 고정, Y1);
      **RED ㉣**: 래치 전 전체 승인 → 차단 없음(빈 backlog 재잠금 없음) + 승격 있음(제때와 같음) — 「실제 삭제 때만 세대 증가」 변형에서 빈
      backlog 재잠금이 남는 것을 관측해 RED 증거로 기록.
      나머지: ㉠ 진행 중 승인 · ㉢ 무관·실패한 승인(차단 유지) · ㉤ 전달·재무장 뒤(차단) · ㉥ · ㉦ · ㉨/㉧ 오류 판정은 원장 승인 상태로 버리지
      않음(Y2 — 부분 승인의 시각이 오류보다 앞서도 차단) · ㉩/㉪ (차단 없음 + 승격) · 실패 기록 `NotFound` 는 「뒤」 해제에서도 승격 없음 ·
      승격 쓰기 실패 → 조건부 차단이 섰든 거절됐든 무조건 차단(Z2 · AA1 — 「차단 성공 → 해제 → 승격 실패」 인터리빙 포함). 실행자가 `n.mu` 를 잡지 않음을 구조로 확인(Q1)
- [x] 2.10 기록 실패 지속(D8, 행별 계수): 한 행의 기록·임차 오류 3 연속 → 잠금(+승격 시도), 최소 ≈ 2 사이클; 그 행의 정산 `Applied` 만
      지우고 `LeaseLost`·`AlreadySettled`·반납 `Applied`·다른 행의 성공은 지우지 않음(A 영구 실패 + B 성공 교차 → 잠금, R3 · X3); 해제 세대가
      바뀌면 연속이 끊김(승인·해제 → 새 행 B → A 기록 오류 한 번 → 잠금 없음); 해제로 이어지지 않은 승인·실패한 승인은 계수를 지우지 않음(Q3);
      한도째 오류 뒤 해제 → 차단 없음 + 승격(원칙 E); 계수 (2, e0) → 한도째 오류 → 해제 → 세대 읽기
      → **리셋 없이** 승격(Z1, 행·나열 둘 다); 기대값(AA2): 해제 없음 → 차단 + 승격, 직전 증가~한도째 사이 해제 → 차단 없음 + 승격, 한도째 뒤 해제 →
      차단 없음 + 승격; 임차가 「이미 정산됨」이면 계수 지움(Y3 직접 시험); 남이 전달한 행의 항목은 완전 나열에서만 지워짐(Q5); 같은 id 재무장(해제 없이 전달 →
      재무장이면 옛 기록 오류를 물려받아 3 회째에 잠금 — 보수 방향 기대값, V5); 나열 오류 3 사이클 → 잠금, 한도째 나열 오류 뒤 해제 → 차단 없음·계수 0
      (W2), 나열 성공이 나열 계수를 지움; Run ctx 취소는 안 셈, transport timeout 은 발행 실패
- [x] 2.11 정제(D9, R2): 잠금 detail 과 새 로그 줄에 행 제목·본문·payload·`last_error`·토큰·계좌 참조·원문 오류가 없다 — 계좌·토큰·원장 오류에
      sentinel 문자열을 심고 줄 전체를 grep
- [x] 2.12 발행 성공 + 전달 기록 실패(D1 전달 정산 표, N3): 오류·`NotFound`·모르는 값 → 즉시 잠금 + 승격, 임차 유지, 다음 사이클 재발행 없음;
      `AlreadySettled`·`LeaseLost` → 잠금 없음; 기록 오류가 승인을 가려도 차단 + 승격(원장 상태로 버리지 않음, Y2); 배치 서비스 중 · 세대 읽기 대기 중
      임차 만료 → 재발행은
      허용되고 기록된다(R6)

- [x] 2.13 해제 세대 계약(execgw, spec 「진입 게이트는 사유별 해제 세대를 가진다」): 해제 요청마다 +1(래치 없어도) · 단조 · 다른 사유
      `Clear` · 모든 `Block` · `BlockSymbol`/`ClearSymbol` · `ProjectOperatingMode` · `RebuildReconcileProjection` 이 이 세대를 안 바꿈 · `revision`
      은 기존대로 실제 변화 때만 · `BlockUnlessClearedSince` 세대 불일치면 무변화·false, 일치면 `Block` 규칙 · `-race` 동시 해제/조건부 잠금 ·
      구조 시험: `Clear(ReasonAlertUndelivered)` 비테스트 호출자 = `Notifier.Acknowledge` 둘뿐(새 호출자면 빨강)
- [x] 2.14 집행 경계 핀(AC1, design D10): (a) 이 change 의 시험은 픽스처에서 `SetModeProjector` · `RestoreOperatingModeProjection` 을 부르지 않는다 —
      부르면 생산에 없는 집행을 시험이 만든다; 모드 단언은 `CurrentOperatingMode`(원장 행)로만. (b) 구조 시험: 두 함수의 비시험 호출자 수 = 0 을
      세어 고정 — 배선(a092 축소판)이 착지하면 빨강이 되고, 그때 D10 · spec 경계 문장 · 이 핀을 함께 고친다(정상 경로). (c) 생산 빌더로 조립한 엔진에서
      「한도 → 승격 → 승인(미전달 0)」 뒤와 「PENDING 0 재시작」 뒤의 경계를 문서와 같게 고정한다 — 안전 주장이 아니라 문서의 참/거짓 핀이다. 단언은 **셋으로
      나눈다**(8판, codex 13회차): ① 게이트 래치 집합에 알림 사유 · 모드 사유가 **둘 다 없음**(사유 단언), ② 모드 행이 원장에 있음(`CurrentOperatingMode`),
      ③ `CheckEntryFor == nil` 은 **별도의 통제된 허용 경우에서만** — 생산 임계값을 유지한 채 필수 조회에 정당한 관측을 준 뒤(생산 빌더 게이트는 필수 조회 미관측으로
      처음부터 막혀 있으므로 「nil 아님」을 알림·모드 사유의 존재로 읽지 않는다). 재잠금 두 변형도 같은 방식으로(AD2): 허용 재잠금 「정산 → 해제 → 세대 읽기 →
      적용」 = 알림 사유 **있음**, 강제 재잠금 「차단 → 해제 → 승격 쓰기 실패」 = 알림 사유 **있음** — 행복 경로 핀을 보편 주장으로 읽지 않게 한다

## 3. GREEN (최소)

- [x] 3.1 `alertDeliverer` 에 `Gate` · `AccountRef` 배선 (`auxiliary.go`)
- [x] 3.2 `deliverOne` 한도 판정(B8 · B9 끝), 상수 하나
- [x] 3.3 `PendingAlertsForDelivery`(새 메서드, 한도 아래 먼저) — `PendingAlerts` 서명·정렬 불변(AA4)
- [x] 3.4 `SettleResult.Attempts` (additive) — `settleUnderClaim` 적용 경로의 같은 트랜잭션 읽기(D1)
- [x] 3.5 `EntryGate` 해제 세대(`clearEpochs` · `ClearEpoch` · `BlockUnlessClearedSince` · `Clear` 한 줄) + 실행자의 세대 읽기(정산 직후 — 실패 기록은 반납 뒤)·조건부 잠금·무조건 승격(D7),
      승격·로그는 잠금 밖
- [x] 3.6 행별 연속 기록 실패 계수 + 나열 계수(D8), 전달 정산 판정(D1), 고정 detail·허용 목록 로그(D9)

## 4. VERIFY

- [x] 4.1 뮤테이션(사본 · 무변이 대조군 선행): 판정 제거 · 승격 제거 · **「뒤」 해제에서 승격도 버림**(2.9 (b) 등가성이 잡음) · 순서 제거 · 행 버림 ·
      나열 값으로 판정 · Outcome 검사 제거 · 세대 비교 제거 · `Clear` 세대 증가 제거 · 세대를 「실제로 지웠을 때만」으로(㉣ 이 잡음) · 다른 사유가
      세대를 올리게 · **세대 읽기를 정산 전으로** — 시험은 「변이 자리의 읽기 → 해제 → 정산 → 정상 자리의 읽기」 순서를 강제해야 잡힌다 ·
      오류 판정에 원장 승인 상태 확인을 넣음(Y2 시험이 잡음) · 기록 실패 계수 제거 · ctx 취소 제외 제거 · `LeaseLost`/반납 `Applied` 로 계수 지움 ·
      「이미 정산됨」 거름 제거 · 전달 정산 실패 판정 제거 · 승격을 게이트 잠금 안으로(2.6 (a) 가 잡음) → 각각 빨강
      — 원장 `analysis/harness/mutation-ledger.tsv` 39/39 CAUGHT(무변이 대조군 GREEN). 「승격을 게이트 잠금 안으로」는 B′ 에서 engine 이 g.mu 아래
        코드를 돌릴 API 가 없어 그대로는 표현 불가 — 구조 핀 `TestTheEpochMethodsCallNothingUnderTheLock` + M28 로 대신(review §1)
- [x] 4.2 `make test` · `make test-seams` · `make test-race` · `make vet` · `make validate` · `make sdd-sync` · `make sdd-check`
      — 2026-09-28 격리 워크트리 `TossOS-worktrees/a124-gate` @ `dfdd6d9a`: test 99 · test-seams 100 · test-race · vet · validate · sdd-check rc 0,
        sdd-sync advisory 미완만(review §1 「4.2 재실행」)
- [x] 4.3 gstack 리뷰 + 독립 적대 diff/test 리뷰 + Manager 검증 패스
      — gstack `/review` · 독립 Eng APPROVE · codex 구현 1~4회차(4회차 PASS). Manager 검증 배터리는 아카이브 전 — 실패하면 이 체크를 되돌림(review §1)
- [x] 4.4 최악 래치 시간 실측(큐 대기 포함) → `review.md` (R3)
      — 가짜 시계, 전제 H: 4 s · 34 s · 214 s(D6 일치), Q = I 판 36 s · 216 s(review §1 4.4). 원장 · 게이트 비용과 Q ≤ C 상한은 흉내 내지 않음

## 5. 종결

- [x] 5.1 (2026-09-28: 돌리고 → 적고 → 체크. 1차 gate(격리 워크트리 TossOS-worktrees/a124-gate · d2b07522)는 2/11 에서 이 줄 하나를 미완료로 세어
      멈췄다(`/tmp/claude-1000/a124-lot/gate1.log`). 2차가 실패하면 이 체크를 되돌린다. 아카이브 · Story 경로 · PM 은 Manager 검증 배터리 뒤.)
      `make gate CHANGE=a124-…` · archive · Story 경로 · PM `--check`
      — **종결 조건(Manager 판정 2026-09-28)**: review.md 종결 절과 아카이브 커밋 메시지에 「§6 배포 · 운영 재측정은 미실행, 사람 몫」 한 줄을
      반드시 남긴다(§6 은 게이트가 안 세므로 아카이브 영수증에서 보이게 — a112 「배포 미완」 선례)
- [x] 5.2 a092 21판에 착수 조건으로 인용 · a092 델타 「굶주림」 문단 삭제 확인 · a092 델타 「운영자의 승인은 … 되살리지 않는다」 문단을
      이 change 의 요구로 가리키게(정본 사본 둘 방지) · **운영 효과 선행 조건 전달(AC1)**: `SetModeProjector` 생산 배선 + 기동 `RestoreOperatingModeProjection`
      을 a092 축소판의 명명된 후속으로(현재 「미배정 후속」, a092 proposal :499 · :582) — AC2(`modegate.go:35-50` 지움→재삽입 창 · 커밋 뒤 투영 순서) 동봉 ·
      a092 델타의 래치 지연 **식을 교체**(`Q + (L−1)·C + I_list + (T + S + M)` + 전제 H —
      「사이클 주기」 재정의는 발행 시간을 이중 계산) · `Acknowledge` 셈~해제 구간의 독립 기록자 경합(Follow-ups)을 a092 소유로 전달
      — 전달 완료(2026-09-28, Manager 경유) — review §1 「5.2 a092 전달 목록」. 수행은 a092 21판 몫

## 6. 배포 (사람 승인 — 완료 게이트 밖, 체크박스 아님)

배포는 사람이 승인하고 실행한다. 아래는 그 절차에 넣을 항목이며, 이 change 의 완료 게이트가 세는 작업이 아니다.

- **운영 환경에서 2.6 전제 재측정** — 실행자 원장 트랜잭션 길이(실패 기록 · 반납 · 승격)와 그 옆 exit 사이클 체류를 운영 디스크에서 잰다. 전제는
  「트랜잭션 하나(실제 비용 포함) ≈25 ms 이내 → 고정 여유 250 ms 안」(design D7 「측정된 전제」; **개발 ext4 NVMe 측정, 운영 fsync 미측정** — 2026-09-28: 실패 기록 평균 11.1 / p99
  19.3 ms, 반납 11.1 / 14.8 ms, 10 ms 주입 수락 중앙값 96~199 ms, 기준선 15~18 ms). 넘으면 배포를 멈추고 Manager 에게 올린다.
  1. 트랜잭션 길이(평균 · p99): `TMPDIR=<운영 원장과 같은 파일시스템의 디렉터리> go test ./internal/app/engine/ -run '^$' -bench
     A124JudgementTransactions -benchtime 200x` — 판정 경로의 세 트랜잭션을 하나씩 잰다(codex 3회차 T3; `A098ClaimSettleRoundTrip` 은 임차 +
     전달의 합이라 이 전제를 직접 재지 못한다).
  2. 체류: 같은 `TMPDIR` 로 `go test ./internal/app/engine/ -run 'TestJudgingTransactionsDelayTheExitCycle|TestTheSelectionScales' -count=1 -v`.
  시험 원장은 `t.TempDir()`(= `$TMPDIR` 아래)에 생기므로 작업 디렉터리를 옮기는 것으로는 디스크가 바뀌지 않는다(codex 2회차 R4); 실행 전
  `df "$TMPDIR"` 로 그 파일시스템이 운영 원장과 같은지 확인한다. 운영 원장 **사본**의 크기를 흉내 내려면 시험 원장이 아니라 사본 위에서 따로 잰다 —
  이 도구들은 빈 새 원장을 쓴다.
- **운영 원장의 `attempts ≥ 3` PENDING 행 수 조회** — 배포 직후 첫 실패 사이클의 승격 여부를 예측한다(design Risks 「배포 직후」, 사람 몫).

