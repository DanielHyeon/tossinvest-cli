# a129 Tasks

Story: `STORY-TOS-a129` · base `381e73da` · High-risk(주문 제출 경로 · exit 관측 주기) — 기록만 더하는 방향.

## 0. 계약·증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a129.yaml`, `_registry.yaml`, `FEAT-TOS-009` 역링크) + tracker 재생성
- [x] 0.2 proposal · design · spec 델타 · tasks, `openspec validate a129-an-exit-order-leaves-a-line --strict --no-interactive` rc=0
- [x] 0.3 `capture_change_base.py` 로 base 고정(`base-commit.txt` = `381e73da`)
- [x] 0.4 CodeGraph hard evidence(`analysis/code-context/`) — `submit`·`ObserveOnce`·`engineRuntime` 의 정의·호출자·피호출자,
      `affected --filter '*_test.go'`
- [x] 0.5 Function Logic Map + Branch Test Map(편집 전) — `ExitObserver.submit` · `ExitObserver.ObserveOnce` · `engineRuntime`
- [x] 0.6 proposal-freeze 리뷰(적대적 Eng 보이스 포함) + Pre-Edit 선언 → `review.md`
      — APPROVE-WITH-FIXES(P1 2 · P2 6 · P3 10), 전부 GREEN 전에 design 2판 · spec 2판 · 시험으로 처리(review.md 표)

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — `internal/app/engine/a129_exit_order_log_test.go` · `cmd/tossctl/a129_engine_runtime_exit_order_log_test.go`
      (design D7 의 1~6) 가 base 에서 실패함을 확인
- [x] 1.2 GREEN — `ExitObserverOptions.ExitOrderLog`, 새 파일 `internal/app/engine/exit_order_log.go`(화이트리스트 줄 · 토큰 거름 ·
      쓰기 하나만 감싼 recover · 거절 래치 · 정규 신원 체결 따라가기 · 무체결 종결 줄), `submit` 의 return 경로별 `noteExitOrder`(대기열),
      `ObserveOnce` 의 `defer o.endExitOrderCycle(ctx)`, `engineRuntime` 의 `ExitOrderLog: logger`
      — 리뷰 반영으로 1판 계획(`followExitFills` defer · `LookupFill`)에서 바뀜
- [x] 1.3 REFACTOR + 편집 뒤 AST 재추출 · FLM/BTM 갱신 · `check_analysis.py --change` 통과

## 2. 검증

- [x] 2.1 변이(사본, 무변이 대조군 GREEN 먼저) — 화이트리스트에 계좌 필드 · 배선 삭제 · nil 가드 삭제 · recover 삭제 · 체결 증가 비교
      반전 → 전부 CAUGHT, 원장 `analysis/mutation/`
      — 최종 round4: CONTROL GREEN, **19/19 CAUGHT**(시험 실패만 셈 — 하네스가 컴파일 실패를 BUILD-FAILED 로 따로 셈; round2 의 해제 삼킴 4건은
      컴파일 실패였음을 이 구분으로 찾아 고침). round1 · round2 는 이력
- [x] 2.2 `make lint` · `go test -count=1 -race ./internal/app/engine/ ./cmd/tossctl/` · `make test-seams` · openspec validate
      — 전부 rc 0(`analysis/verify/receipts.md`). race 는 엔진 패키지 2093s 라 기본 10m timeout 을 넘음 → `-timeout 45m`
- [x] 2.3 `make sdd-sync`(CodeGraphContext 300s 타임아웃은 advisory — rc·원인 기록) · `make sdd-check`
      — sdd-sync rc 0: codegraph sync 4 파일 · CodeGraphContext update 이번엔 타임아웃 없이 완료 · GBrain 은 다른 세션 소유(`gbrain serve` pid 23201)라 advisory busy
- [x] 2.4 구현 리뷰(분리된 리뷰 패스) → `review.md` — APPROVE-WITH-FIXES(P1 1 · P2 2 · P3 7), 처리 표 review.md

## 3. 사람 · Manager

- [ ] 3.1 Manager `make gate CHANGE=a129-an-exit-order-leaves-a-line` + 독립 검증 + 푸시
- [ ] 3.2 배포 — 사람이 KST 05:00~09:00 에 `make image CHANGE=a129-an-exit-order-leaves-a-line` → 컨테이너 교체(design 「배포」).
      교체 뒤 다음 청산 주문에서 `"origin":"exit"` 줄 확인
- [ ] 3.3 archive + Story 경로 갱신 + PM check
