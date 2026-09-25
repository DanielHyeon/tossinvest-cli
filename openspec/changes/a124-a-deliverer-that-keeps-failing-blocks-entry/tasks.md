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
- [ ] 1.2 Pre-Edit 선언(High-risk): 대상 심볼 · 호출부 · 기존 시험 · 불변식 · 실패 시험 · rollback
- [ ] 1.3 편집 대상 함수 `revision: current` 재추출, Branch Test Map 재번호(옛/새 ast diff 정렬)
- [ ] 1.4 Branch Test Map 의 기존 시험 실측 — `go test -covermode=set` 으로 분기 도달을 재고 파일 이름을 함수 이름으로

## 2. RED

- [ ] 2.1 한도 도달 → `Gate.Block(ReasonAlertUndelivered)` + `EscalateOperatingMode(CRITICAL_ALERT_UNDELIVERED)`,
      **동기 경로가 한 번도 시도하지 않은 구성**(publisher 응답 없음 · publisher 없음 둘 다)
- [ ] 2.2 승인으로 미전달 0 → 차단 해제, 모드는 남는다
- [ ] 2.3 한도 도달 뒤 재시작 → 기동 복원이 다시 잠그고 모드는 원장에 있다
- [ ] 2.4 굶주림: 한도 행 batch 개 + 새 critical 행 → 다음 사이클이 새 행을 먼저 시도한다
- [ ] 2.5 한도 행은 잔여 자리로 밀려도 사라지지 않고 미전달 수에 세어진다
- [ ] 2.6 손절 즉시성 불변 — `a098_the_backlog_does_not_delay_protection_test.go` 그대로 초록(보존 시험)

## 3. GREEN (최소)

- [ ] 3.1 `alertDeliverer` 에 `Gate` · `AccountRef` 배선 (`auxiliary.go`)
- [ ] 3.2 `deliverOne` 한도 판정(B8 · B9 끝), 상수 하나
- [ ] 3.3 `PendingAlerts` 선택 순서(한도 아래 먼저)

## 4. VERIFY

- [ ] 4.1 뮤테이션(사본 · 무변이 대조군 선행): 판정 제거 · 승격 제거 · 순서 제거 · 행 버림 → 각각 빨강
- [ ] 4.2 `make test` · `make test-seams` · `make test-race` · `make vet` · `make validate` · `make sdd-sync` · `make sdd-check`
- [ ] 4.3 gstack 리뷰 + 독립 적대 diff/test 리뷰 + Manager 검증 패스
- [ ] 4.4 최악 래치 시간 실측(큐 대기 포함) → `review.md` (R3)

## 5. 종결

- [ ] 5.1 `make gate CHANGE=a124-…` · archive · Story 경로 · PM `--check`
- [ ] 5.2 a092 21판에 착수 조건으로 인용 · a092 델타 「굶주림」 문단 삭제 확인
