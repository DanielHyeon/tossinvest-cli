## 0. 계약과 증거

- [x] 0.1 `base-commit.txt` 고정 — proposal-freeze 직전 (`capture_change_base.py`) → `4798d399` (2026-09-26; AST 추출 `463cc895` 와 Go diff 0)
- [x] 0.2 `openspec validate a124-a-deliverer-that-keeps-failing-blocks-entry --strict` 통과 (openspec 1.4.1, rc 0, 2026-09-26)
- [x] 0.3 **AST 산출물이 문서보다 먼저** — 함수 5개(`cycle` · `deliverOne` · `Notifier.deliver` · `Notifier.notifyCritical` ·
      `restoreAlertEntryLatch`), 분기 48, HEAD `463cc895` (2026-09-25, Manager). 편집 대상 둘은 편집 뒤 재추출(1.3)
- [ ] 0.4 proposal-freeze 리뷰(**적대적 Eng 필수 + 교차 모델**) → `review.md`. 열린 결정 D2 · D3 를 답한다
      — 1판 실행(2026-09-26, Teammate 적대 Eng + codex): **REJECT**. D2 = ㄱ(3) · D3 = 센다로 답함. F1·F2·F3·F4·F8 과
      Manager 결정 M1·M2 를 design 에 반영한 뒤 재freeze (`review.md` §0)

## 1. 증거와 Pre-Edit

- [x] 1.1 CodeGraph: `deliverOne` · `cycle` · `PendingAlerts` · `EntryGate.Block` · `EscalateOperatingMode` 의 callers/callees →
      `analysis/code-context/` (codegraph 1.6.0; CGC kuzu 잠금·GBrain busy 로 not-applicable, 불일치 R1–R5 는 HEAD·AST 로 해소, R6 미해소 = F1)
- [ ] 1.2 Pre-Edit 선언(High-risk): 대상 심볼 · 호출부 · 기존 시험 · 불변식 · 실패 시험 · rollback — 대상에 **`execgw.EntryGate`**(`Clear` ·
      새 `ClearEpoch` · `BlockUnlessClearedSince`)를 포함하고, `Clear(ReasonAlertUndelivered)` 호출자 둘과 `revision` 소비자(전략 봉인)를 다시 잰다
- [ ] 1.3 편집 대상 함수 `revision: current` 재추출, Branch Test Map 재번호(옛/새 ast diff 정렬) — 대상 셋:
      `alertDeliverer.cycle` · `alertDeliverer.deliverOne` · `Journal.settleUnderClaim` · `EntryGate.Clear`
      (뒤의 둘은 base 번들 2026-09-26; `Notifier.Acknowledge` 는 5판에서 편집 대상이 아니다).
      `Journal.PendingAlerts` 도 편집 대상이면 base 번들을 **편집 전에** 만든다
- [ ] 1.4 Branch Test Map 의 기존 시험 실측 — `go test -covermode=set` 으로 분기 도달을 재고 파일 이름을 함수 이름으로

## 2. RED

- [ ] 2.1 한도 도달 → `Gate.Block(ReasonAlertUndelivered)` + `EscalateOperatingMode(CRITICAL_ALERT_UNDELIVERED)`,
      **동기 경로가 한 번도 시도하지 않은 구성**(publisher 응답 없음 · publisher 없음 둘 다)
- [ ] 2.2 승인으로 미전달 0 → 차단 해제, 모드는 남는다
- [ ] 2.3 한도 도달 뒤 재시작 → 기동 복원이 다시 잠그고 모드는 원장에 있다
- [ ] 2.4 굶주림: 한도 행 batch 개 + 새 critical 행 → 다음 사이클이 새 행을 먼저 시도한다
- [ ] 2.5 한도 행은 잔여 자리로 밀려도 사라지지 않고 미전달 수에 세어진다
- [ ] 2.6 손절 즉시성 불변 — `a098_the_backlog_does_not_delay_protection_test.go` 그대로 초록(보존 시험) + 변형:
      publisher 없음(D3 원장 쓰기 증가, F12) · 실행자의 정산·승격을 인위 지연 → exit 체류 불변 · 실행자가 `g.mu` 를 기다리게(전략 dispatch
      모사로 `g.mu` 보유) 해도 exit `Notify` · 비상 청산 `Notify` 가 실행자 때문에 기다리지 않음(Q1). 늘면 멈추고 보고
- [ ] 2.7 판정 입력(D1, F1): 나열 뒤 다른 발송자가 attempts 를 올린 행 → 판정은 커밋된 값으로; 재무장된 행은 0 부터;
      기록 오류의 `SettleResult{}` 가 「적용됨」으로 새지 않는다(F10)
- [ ] 2.8 발송 중 승인(D7): 전송 중 승인 → 기록 `AlreadySettled` → 잠금·승격 없음. `LeaseLost` 도 없음(재무장 인터리빙 포함).
      실패 기록의 `NotFound`·모르는 값은 잠금만(승격 없음, N6) — 울타리 재시도·미룬 판정 경로를 거쳐도 승격 없음(W4)
- [ ] 2.9 해제 세대 울타리(D7): 결정적 인터리빙 — 판별 케이스 ㉠~㉧ 각 하나(진행 중 승인과 겹침 · 해제 뒤 늦은 실패/기록 오류 · 무관한 승인 ·
      실패한 승인 · 래치 전 전체 승인 · 전달·재무장 뒤 늦은 잠금 · 정산 전 승인 · 정산 전 재무장 · 한 행만 승인 + 전달 정산 오류)
      + 임차와 해제 겹침(㉨ — `e1 != e2` 면 발행 없음) + 셈~해제 사이 독립 `EnqueueAlert`(㉩ — 원장 확인 후 잠금) + 원장 확인 읽기 오류면
      잠금 + 원장 DELIVERED 면 잠금(㉪, W1) + 세 번 연속 해제 경합이면 다음 사이클로 미루고 그 사이 승인됐으면 버림(㉫, W3) + 「해제 → 같은 id 재무장 → 임차」에서 새 에피소드 판정 유지. 실행자가
      `n.mu` 를 잡지 않음을 구조로 확인(Q1)
- [ ] 2.10 기록 실패 지속(D8, 행별 계수): 한 행의 기록·임차 오류 3 연속 → 잠금(+승격 시도), 최소 ≈ 2 사이클; 그 행의 `Applied` 만
      지우고 `LeaseLost`·`AlreadySettled`·다른 행의 성공은 지우지 않음(A 영구 실패 + B 성공 교차 → 잠금, R3); 승인이 맵을 비움
      (승인 → 새 행 B → A 기록 오류 → 재잠금 없음, R4); 해제로 이어지지 않은 승인·실패한 승인은 계수를 지우지 않음(Q3); 남이 전달한 행의
      항목은 완전 나열에서만 지워짐(Q5); 같은 id 재무장(해제 없이 전달 → 재무장이면 옛 기록 오류를 물려받아 3 회째에 잠금 — 보수 방향 기대값, V5); 나열 오류 3 사이클 → 잠금, 나열 계수 울타리 실패 → 재계수(해제 경합 · 나열 회복 · 새 행 · 계속 실패, W2); Run ctx 취소는 안 셈, transport timeout 은 발행 실패
- [ ] 2.11 정제(D9, R2): 잠금 detail 과 새 로그 줄에 행 제목·본문·payload·`last_error`·토큰·계좌 참조·원문 오류가 없다 — 계좌·토큰·원장 오류에
      sentinel 문자열을 심고 줄 전체를 grep
- [ ] 2.12 발행 성공 + 전달 기록 실패(D1 전달 정산 표, N3): 오류·`NotFound`·모르는 값 → 즉시 잠금 + 승격, 임차 유지, 다음 사이클 재발행 없음;
      `AlreadySettled`·`LeaseLost` → 잠금 없음; 기록 오류 직전 승인 → 잠금 없음(세대 울타리); 배치 서비스 중 임차 만료 → 재발행은
      허용되고 기록된다(R6)

- [ ] 2.13 해제 세대 계약(execgw, spec 「진입 게이트는 사유별 해제 세대를 가진다」): 해제 요청마다 +1(래치 없어도) · 단조 · 다른 사유
      `Clear` · 모든 `Block` · `BlockSymbol`/`ClearSymbol` · `ProjectOperatingMode` · `RebuildReconcileProjection` 이 이 세대를 안 바꿈 · `revision`
      은 기존대로 실제 변화 때만 · `BlockUnlessClearedSince` 세대 불일치면 무변화·false, 일치면 `Block` 규칙 · `-race` 동시 해제/조건부 잠금 ·
      구조 시험: `Clear(ReasonAlertUndelivered)` 비테스트 호출자 = `Notifier.Acknowledge` 둘뿐(새 호출자면 빨강)
## 3. GREEN (최소)

- [ ] 3.1 `alertDeliverer` 에 `Gate` · `AccountRef` 배선 (`auxiliary.go`)
- [ ] 3.2 `deliverOne` 한도 판정(B8 · B9 끝), 상수 하나
- [ ] 3.3 `PendingAlerts` 선택 순서(한도 아래 먼저)
- [ ] 3.4 `SettleResult.Attempts` (additive) — `settleUnderClaim` 적용 경로의 같은 트랜잭션 읽기(D1)
- [ ] 3.5 `EntryGate` 해제 세대(`clearEpochs` · `ClearEpoch` · `BlockUnlessClearedSince` · `Clear` 한 줄) + 실행자의 세대 읽기(임차 전후 두 번)·조건부 잠금·원장 확인·미룬 판정(D7),
      승격·로그는 잠금 밖
- [ ] 3.6 행별 연속 기록 실패 계수 + 나열 계수(D8), 전달 정산 판정(D1), 고정 detail·허용 목록 로그(D9)

## 4. VERIFY

- [ ] 4.1 뮤테이션(사본 · 무변이 대조군 선행): 판정 제거 · 승격 제거 · 순서 제거 · 행 버림 · 나열 값으로 판정 ·
      Outcome 검사 제거 · 세대 비교 제거 · `Clear` 세대 증가 제거 · 세대를 「실제로 지웠을 때만」으로 · 다른 사유가 세대를 올리게 ·
      세대 읽기를 한 번만(임차 전 또는 후) · 원장 확인 제거(울타리 실패면 바로 버림) · DELIVERED 로 버림 · 미룬 판정 제거 · 기록 실패 계수 제거 · ctx 취소 제외 제거 ·
      `LeaseLost` 로 계수 지움 · 승인에 맵 비움 제거 · 전달 정산 실패 판정 제거 · 정산을 배제 안으로 → 각각 빨강(마지막은 2.6 이 잡아야 한다)
- [ ] 4.2 `make test` · `make test-seams` · `make test-race` · `make vet` · `make validate` · `make sdd-sync` · `make sdd-check`
- [ ] 4.3 gstack 리뷰 + 독립 적대 diff/test 리뷰 + Manager 검증 패스
- [ ] 4.4 최악 래치 시간 실측(큐 대기 포함) → `review.md` (R3)

## 5. 종결

- [ ] 5.1 `make gate CHANGE=a124-…` · archive · Story 경로 · PM `--check`
- [ ] 5.2 a092 21판에 착수 조건으로 인용 · a092 델타 「굶주림」 문단 삭제 확인 · a092 델타 「운영자의 승인은 … 되살리지 않는다」 문단을
      이 change 의 요구로 가리키게(정본 사본 둘 방지) · a092 델타의 래치 지연 **식을 교체**(`Q + (L−1)·C + I_list + (T + S + M)` + 전제 H —
      「사이클 주기」 재정의는 발행 시간을 이중 계산) · `Acknowledge` 셈~해제 구간의 독립 기록자 경합(Follow-ups)을 a092 소유로 전달
