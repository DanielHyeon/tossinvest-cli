# a130 Tasks

Story: `STORY-TOS-a130` · base `233d5d9c` · High-risk(원장 — 체결 반영 트랜잭션 안의 성과 동결 · 엔진 기동 배선).

## 0. 계약·증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a130.yaml`, `_registry.yaml`, `FEAT-TOS-005` 역링크) + tracker 재생성
- [x] 0.2 proposal · design · spec 델타 · tasks, `openspec validate a130-an-adopted-trade-keeps-its-outcome --strict --no-interactive` rc=0
- [x] 0.3 `capture_change_base.py` 로 base 고정(`base-commit.txt` = `233d5d9c`)
- [x] 0.4 CodeGraph hard evidence(`analysis/code-context/`) — 정의 · 호출자 · 피호출자, `affected --filter '*_test.go'`
- [ ] 0.5 Function Logic Map + Branch Test Map(편집 전) — `freezeTradeOutcomeTx` · `computeTradeOutcome` · `adoptedRoundTripLegs` ·
      `Journal.BackfillTradeOutcome` · `ApplyExitFill` · `Journal.runApplyHooks` · `bindApplyHooks` · `NewContext`
      — 1판 8 번들 완료. **[blocked]** 리뷰 F-5 가 번들 5개 추가 요구(`ApplyPositionAdjustment` · `ConvergeQuantities` ·
      `resetExitStateForReadoptTx` · `engineSellOrders` · `roundTripLegs`) — 재설계 뒤
- [ ] 0.6 proposal-freeze 리뷰(적대적 Eng 보이스, 분리된 컨텍스트) + Pre-Edit 선언 → `review.md`
      — **[blocked]** 1차 2026-10-13 **REJECT**(P0 2 · P1 4 · P2 4 · P3 3, review.md). P0 F-1 = 사용자 결정 1항 전제 오류 →
      사람 결정 대기(issues.md I-1 · I-2). 결정 뒤 design/spec 2판 → 재리뷰

## 1. 구현 (RED → GREEN → REFACTOR)

- [ ] 1.1 RED — design D7 1~8 시험이 base 에서 실패함을 확인(`analysis/red/`)
      — 1판 RED 기록 있음(`analysis/red/red-1.txt`: 컴파일 RED + 격리 worktree 행동 RED — 사건 재현 FAIL · 결정 경로 회귀 PASS). 2판 설계에 맞춰 다시
- [ ] 1.2 GREEN — 편입 뒤 양의 조정 다리 · 사유 코드 · `ApplyHooks.OutcomeRefused` · 엔진 싱크 · backfill 계획/실행 · 도구 · 콘솔 안내문
- [ ] 1.3 REFACTOR + 편집 뒤 AST 재추출 · FLM/BTM 갱신 · `check_analysis.py --change` 통과

## 2. 검증

- [ ] 2.1 변이(사본, 무변이 대조군 GREEN 먼저) — 수량 비교 원복 · 조정분 누락 · 거절 기록 삭제 · 분모 수량 바꿔치기 외, 원장 `analysis/mutation/`
- [ ] 2.2 `make lint` · 관련 패키지 `go test -race -timeout 45m` · `make test-seams` · openspec validate
- [ ] 2.3 운영 저널 **사본**에 미리보기(읽기 전용) — 11건 재계산 · 기존 31행 재계산 일치 → `analysis/verify/`
- [ ] 2.4 `make sdd-sync` · `make sdd-check`
- [ ] 2.5 구현 리뷰(분리된 리뷰 패스) → `review.md`

## 3. 사람 · Manager

- [ ] 3.1 Manager `make gate CHANGE=a130-an-adopted-trade-keeps-its-outcome` + 독립 검증 + 푸시

### 게이트 뒤 절차 (체크박스 아님 — a108 · a129 선례)

1. archive → Story `openspec.path` 를 아카이브 경로로 → tracker 재생성 → PM `--check`.
2. 배포 — 사람이 KST 05:00~09:00(두 시장 닫힘)에 `make image CHANGE=a130-an-adopted-trade-keeps-its-outcome` → 컨테이너 교체.
   스키마 불변(35)이라 롤백은 이미지 핀 되돌림.
3. backfill — 배포 뒤 **사람이** 운영 저널에: `go run ./tools/a130-outcome-backfill` (미리보기, 쓰기 0) → 출력 확인 →
   `go run ./tools/a130-outcome-backfill --execute --confirm <토큰>`. 에이전트는 실행하지 않는다.
