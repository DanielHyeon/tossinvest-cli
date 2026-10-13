# a130 Tasks

Story: `STORY-TOS-a130` · base `233d5d9c` · High-risk(원장 — 체결 반영 트랜잭션 안의 성과 동결 · 엔진 기동 배선). 4판(묶음별 브로커 1분봉 재계산).

## 0. 계약·증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a130.yaml`, `_registry.yaml`, `FEAT-TOS-005` 역링크) + tracker 재생성
- [x] 0.2 proposal · design · spec 델타 · tasks, `openspec validate a130-an-adopted-trade-keeps-its-outcome --strict --no-interactive` rc=0
- [x] 0.3 `capture_change_base.py` 로 base 고정(`base-commit.txt` = `233d5d9c`)
- [x] 0.4 CodeGraph hard evidence(`analysis/code-context/`) — 정의 · 호출자 · 피호출자, `affected --filter '*_test.go'`
- [x] 0.5 Function Logic Map + Branch Test Map(편집 전) — 편집 대상 9(`freezeTradeOutcomeTx` · `computeTradeOutcome` ·
      `adoptedRoundTripLegs` · `Journal.BackfillTradeOutcome` · `ApplyExitFill` · `Journal.runApplyHooks` · `Journal.RecordFill` · `bindApplyHooks` ·
      `NewContext`) + 무편집 인용 근거 6(F-5: `ApplyPositionAdjustment` · `Converger.ConvergeQuantities` · `resetExitStateForReadoptTx` ·
      `engineSellOrders` · `roundTripLegs`, 그리고 `Journal.backfillConfirmedStrategyFillTx`)
- [x] 0.6 proposal-freeze 리뷰(적대적 Eng 보이스, 분리된 컨텍스트) + Pre-Edit 선언 → `review.md`
      — 1차 REJECT(F-1~F-13) · 2차 APPROVE-WITH-FIXES(R2) · 3차 REJECT(R3-1 가격 출처 원장에 없음) → 사용자 결정 3차(브로커 1분봉) → 4판 ·
      4차 **APPROVE-WITH-FIXES**(R4-1~R4-13 — GREEN 커밋에 전부 반영, review.md)

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — design D7 시험이 base 에서 실패함을 확인(`analysis/red/red-4.txt` — 컴파일 RED + 로컬 매도 · flatten 쓸어 담기 조작 행 행동 RED; 이력 red-1 · red-2)
- [x] 1.2 GREEN — 사유 코드 타입 · 묶음별 1분봉 산식 · 편입 시기 · 인스턴스 매도 거절 · `ApplyHooks.OutcomeRefused` · 커밋 뒤 보고 · 엔진 싱크 ·
      공식 1분봉 읽기 · backfill 계획/실행 · 도구 · 콘솔 안내문(실제로 렌더해 확인)
- [x] 1.3 REFACTOR + 편집 뒤 AST 재추출 · FLM/BTM 갱신 · `check_analysis.py --change` 통과
      — a130 번들 오류 0. 남은 요구는 base 233d5d9c 뒤 착지한 a131 커밋 9개의 함수(쌓인 change) — **base 재고정 필요(사람 · Manager 승인 절차,
      WORKFLOW 「사람 승인 base 재고정」)**: 옛 base 이후 이 change 의 자기 Go 커밋 = 이 GREEN 하나(재고정은 이 커밋의 부모 af0c0ce7 로 — a130 FLM 번들의
      편집 전 AST 8 파일은 233d5d9c = af0c0ce7, 리뷰 4차 확인)

## 2. 검증

- [ ] 2.1 변이(사본, 무변이 대조군 GREEN 먼저) — design D7 변이 목록, 원장 `analysis/mutation/`
- [ ] 2.2 `make lint` · 관련 패키지 `go test -race -timeout 45m` · `make test-seams` · openspec validate
- [ ] 2.3 운영 저널 **사본**(읽기 전용) — census(동결 행 회귀 · 사유 분포) + 도구 미리보기(브로커 1분봉 읽기 전용 GET) + 이관한 사본에서 실행해 항목별 쓰기
      잠금 시간 → `analysis/verify/`(수 · 사유만, F-6), GREEN 커밋 sha 기록. 항목별 결과(종목 · R · 분봉)는 완료 보고에만
- [ ] 2.4 `make sdd-sync` · `make sdd-check`
- [ ] 2.5 구현 리뷰(분리된 리뷰 패스) → `review.md`

## 3. 사람 · Manager

- [ ] 3.1 Manager `make gate CHANGE=a130-an-adopted-trade-keeps-its-outcome` + 독립 검증 + 푸시

### 게이트 뒤 절차 (체크박스 아님 — a108 · a129 선례)

1. archive → Story `openspec.path` 를 아카이브 경로로 → tracker 재생성 → PM `--check`.
2. 배포 — 사람이 KST 05:00~09:00(두 시장 닫힘)에 `make image CHANGE=a130-an-adopted-trade-keeps-its-outcome` → 컨테이너 교체.
   스키마 불변(35)이라 롤백은 이미지 핀 되돌림.
3. backfill — 배포(스키마 35) 뒤 곧 **사람이 두 시장이 닫힌 시간에** 운영 저널에:
   `go run ./tools/a130-outcome-backfill --journal ~/.config/tossctl/journal.db --config-dir ~/.config/tossctl`(미리보기, 원장 쓰기 0, 1분봉 읽기 전용 GET,
   엔진과 같은 토큰 캐시 — 출력의 `token cache:` 경로가 엔진의 것인지 확인) → 출력 확인 → `... --execute --confirm <토큰>`. 에이전트는 실행하지 않는다. 스키마가 빌드와 다르면 도구가 거절한다. 앞으로 생기는 편입 뒤 증가 종결도 같은 절차로(자동화는 issues I-4).
