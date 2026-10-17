# a146 Tasks

Story: `STORY-TOS-a146` · **High-risk**(Guardian · 전략 진입 · 레인 잠금 분류) — Pre-Edit 선언 · 편집 함수 FLM/BTM · RED 선행 · 변이 원장 · 독립 적대 리뷰 의무.

> 병행: Codex 세션이 같은 트리에서 활동 중(미커밋 `.claude/CLAUDE.md` · `docs/architecture*` · `_registry.yaml`(a145) · `FEAT-TOS-006` · `docs/pm/generated/*`,
> a128/a140 shadow 원장 append). 이 change 는 자기 경로만 커밋한다 — `_registry.yaml` 은 HEAD + a146 줄만 index 에 올림.

## 0. 계약 · 증거

- [x] 0.1 Story 등록(`docs/pm/portfolio/stories/STORY-TOS-a146.yaml`, `_registry.yaml`, `FEAT-TOS-003` 역링크) + proposal
- [x] 0.2 design(F1~F9 편집 전 사실 · D1 분류 · D2 전달 · D3 흔적 · D4 불변 · D5 풀리는 것 · D6 기각 · D7 범위 거절 별도 발견 · D8 변이 계획) · spec 델타 · tasks —
      AST 근거 18건(`analysis/ast-evidence/`, 편집 전 HEAD `befea6ee`)
- [x] 0.3 CodeGraph hard evidence(`analysis/code-context/`, codegraph 1.6.0) · CodeGraphContext(not-applicable 사유) · evidence reconciliation
- [x] 0.4 특성 시험(편집 전 HEAD 에서 통과 — 상한 0: 주기 멈춤 · effective 잠금 · refreshOnly 삼킴, 범위 거절: effective 잠금) + `base-commit.txt`
      = `7ce7e411`(첫 Go 커밋의 부모) — 영수증 `analysis/red/characterization-head.txt`(4/4 PASS, `-count=5` · `-race -count=2` 도 통과)
- [x] 0.5 proposal-freeze 적대 리뷰(분리 컨텍스트) → review.md → 수리 — APPROVE-WITH-FIXES(P0 0 · P1 1 · P2 3 · P3 7), 전부 문서에 반영(P1-1 은 명시 한계 + 시험,
      결속 판정 이동 안은 사람 결정 잔여)
- [x] 0.6 편집할 기존 함수의 FLM/BTM 번들(8 — 하네스 `analysis/harness/a146_flm.py base`, ast-base = base 7ce7e411) + Pre-Edit 선언(review.md)

## 1. 구현 (RED → GREEN → REFACTOR)

- [x] 1.1 RED — 분류 표(상한 0 · 해석 실패 · nil 원인 · FX 오류 · 다른 코드 + 같은 sentinel) · 전달(스킵 뒤 다음 범위 · 범위 거절과 섞임 · 스킵 뒤 결함 ·
      스킵 뒤 일일 손실) · 감독자 두 종류(잠금 0 · 삼킴 0 · 계수) · 알림(일 키 · 종목 dedup) · 생산 배선 · 실제 경로(fixture: precheck 2 회 · 오류 없음 · 브로커 0) ·
      census(만드는 자리 하나) + freeze 리뷰: 상한 0 이 가린 체인 결함(양쪽) · 실제 Guardian 거절 표본 · 곁 Join 음성 행 — 영수증 `analysis/red/red-compile.txt`
- [x] 1.2 GREEN — `a146_entry_skip.go` · `admit` · `deliverEachStrategyHandoff` · `runMarket` · `Snapshot` · `NewStrategyEntrySupervisor` ·
      `NewRefreshingPairedStrategyEntrySupervisor` · obs 이벤트; 특성 시험 F1/F4/F5 를 새 동작으로 뒤집음, `TestAnExistingGuardianCapRefusalIsNotAScopeRefusal` 고침,
      `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` 칸 목록 + 근거(a137 선례)
- [ ] 1.3 REFACTOR · 편집 뒤 AST 재추출 · `check_analysis.py --change a146-guardian-cap-zero-skips-the-entry`

## 2. 검증

- [x] 2.1 `gofmt -l` · `go vet ./...` · `rtk proxy go test -count=1`(app/engine · execgw · risk · strategyworker · obs) · `-tags tossos_testseams` · `-race` 대상 ·
      `openspec validate --strict` → `analysis/verify/receipts.md`
- [ ] 2.2 변이 원장(`analysis/mutation/`, 저장소 밖 사본 · pid 접미 · 무변이 대조군 GREEN)
- [ ] 2.3 독립 적대 리뷰(분리 컨텍스트) → review.md → 수리

## 3. 사람 · Manager

- [ ] 3.1 Manager `make sdd-sync` · `make sdd-check` · `make gate CHANGE=a146-guardian-cap-zero-skips-the-entry` + 독립 검증 · PM tracker 재생성 · archive
- 배포(체크박스 아님): 엔진 이미지 교체는 사람이 KST 05:00~09:00 창에서.
