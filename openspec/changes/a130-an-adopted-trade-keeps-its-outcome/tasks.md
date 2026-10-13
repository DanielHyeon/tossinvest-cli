# a130 Tasks

Story: `STORY-TOS-a130` · base `233d5d9c` · High-risk(원장 — 체결 반영 트랜잭션 안의 성과 동결 · 엔진 기동 배선). 3판(묶음별 편입가 재계산).

## 0. 계약·증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a130.yaml`, `_registry.yaml`, `FEAT-TOS-005` 역링크) + tracker 재생성
- [x] 0.2 proposal · design · spec 델타 · tasks, `openspec validate a130-an-adopted-trade-keeps-its-outcome --strict --no-interactive` rc=0
- [x] 0.3 `capture_change_base.py` 로 base 고정(`base-commit.txt` = `233d5d9c`)
- [x] 0.4 CodeGraph hard evidence(`analysis/code-context/`) — 정의 · 호출자 · 피호출자, `affected --filter '*_test.go'`
- [x] 0.5 Function Logic Map + Branch Test Map(편집 전) — 편집 대상 9(`freezeTradeOutcomeTx` · `computeTradeOutcome` ·
      `adoptedRoundTripLegs` · `Journal.BackfillTradeOutcome` · `ApplyExitFill` · `Journal.runApplyHooks` · `Journal.RecordFill` · `bindApplyHooks` ·
      `NewContext`) + 무편집 인용 근거 6(F-5: `ApplyPositionAdjustment` · `Converger.ConvergeQuantities` · `resetExitStateForReadoptTx` ·
      `engineSellOrders` · `roundTripLegs`, 그리고 `Journal.backfillConfirmedStrategyFillTx`)
- [ ] 0.6 proposal-freeze 리뷰(적대적 Eng 보이스, 분리된 컨텍스트) + Pre-Edit 선언 → `review.md`
      — 1차 REJECT(F-1~F-13) · 2차 APPROVE-WITH-FIXES(R2-1~R2-11, 2판) → 사용자 결정 변경(재계산) → 3판 · 3차 **REJECT**(R3-1 P0:
      결정한 가격 출처가 원장에 없음) → **[blocked]** issues.md I-3 사람 결정 대기

## 1. 구현 (RED → GREEN → REFACTOR)

- [ ] 1.1 RED — design D7 시험이 base 에서 실패함을 확인(`analysis/red/red-3.txt`; 이력 red-1 · red-2)
- [ ] 1.2 GREEN — 사유 코드 타입 · 묶음별 편입가 · 편입 시기 거절 · `ApplyHooks.OutcomeRefused` · 커밋 뒤 보고 · 엔진 싱크 · backfill 계획/실행 · 도구 ·
      콘솔 안내문(실제로 띄워 확인)
- [ ] 1.3 REFACTOR + 편집 뒤 AST 재추출 · FLM/BTM 갱신 · `check_analysis.py --change` 통과

## 2. 검증

- [ ] 2.1 변이(사본, 무변이 대조군 GREEN 먼저) — design D7 변이 목록, 원장 `analysis/mutation/`
- [ ] 2.2 `make lint` · 관련 패키지 `go test -race -timeout 45m` · `make test-seams` · openspec validate
- [ ] 2.3 운영 저널 **사본**(읽기 전용) — census(동결 행 회귀 · 사유 분포 · backfill 모집단) + 도구 미리보기 → `analysis/verify/`(수 · 사유만, F-6),
      GREEN 커밋 sha 기록. 항목별 미리보기(종목 · R · 관측 지연 · 사유)는 저장소 밖 · 완료 보고에만
- [ ] 2.4 `make sdd-sync` · `make sdd-check`
- [ ] 2.5 구현 리뷰(분리된 리뷰 패스) → `review.md`

## 3. 사람 · Manager

- [ ] 3.1 Manager `make gate CHANGE=a130-an-adopted-trade-keeps-its-outcome` + 독립 검증 + 푸시

### 게이트 뒤 절차 (체크박스 아님 — a108 · a129 선례)

1. archive → Story `openspec.path` 를 아카이브 경로로 → tracker 재생성 → PM `--check`.
2. 배포 — 사람이 KST 05:00~09:00(두 시장 닫힘)에 `make image CHANGE=a130-an-adopted-trade-keeps-its-outcome` → 컨테이너 교체.
   스키마 불변(35)이라 롤백은 이미지 핀 되돌림.
3. backfill — 배포(스키마 35) 뒤 **사람이** 운영 저널에: `go run ./tools/a130-outcome-backfill --journal ~/.config/tossctl/journal.db`
   (미리보기, 쓰기 0) → 출력 확인 → `... --execute --confirm <토큰>`. 에이전트는 실행하지 않는다. 스키마가 빌드와 다르면 도구가 거절한다.
