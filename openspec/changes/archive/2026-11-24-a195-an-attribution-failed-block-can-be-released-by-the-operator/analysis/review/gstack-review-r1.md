# gstack /review — 1회차 (2026-11-23, Manager 지시로 1회차에서 끝냄)

- 대상: 브랜치 `fix/a195-attribution-failed-operator-release`, merge-base `064f59fd9`(= base), 리뷰 시작 HEAD `72927ba00`.
- PR type: mixed(application Go + 시험 + testdata + openspec 증거 문서). DIFF_LINES 9796(코드 cmd/internal 411, 나머지 openspec 증거).
- 리뷰어들의 코드 범위는 `git diff 064f59fd9 -- cmd internal` 로 좁혔음 — openspec analysis 는 생성 증거라 의도(design.md)만 읽게 함(범위 축소를 여기 적음).
- 계획: OpenSpec change a195(proposal/design/tasks)를 의도 원천으로 씀. 별도 docs/designs 계획 파일 없음.

## Scope Check: CLEAN
- Intent: 운영자 해제 명령이 종목 범위 ATTRIBUTION_FAILED 차단을 기존 조건으로 받음, 원인 표 한 곳, 원인 중립 문구, census 키 하나.
- Delivered: 위 그대로(reconcile_resolve_causes.go 신규, engine_reconcile.go 클로저 1줄·도움말·성공 문장, shapes.json 1키, 시험 파일 신규).
- Diff read: full(코드 4 파일), openspec 증거는 요약.
- Undeclared behavior changes: none.

## Gate edits: 2 listed, 1 read, 1 finding — 0 kept, 0 restored, 1 open
- `[54d40e57b757] RH-3? internal/accountbudgetruntime/testdata/a166-wiring/shapes.json` — 고정 모양(census) 값 갱신, 소유 시험 `TestA166WiringCensusBaseline` 은 무변경. 근거: design D6(Manager 결정 2026-11-23), `analysis/census-repair/shapes-key-diff.txt`(base 재현 True, 90 키 중 1 변경, sites.json 무변경), 변이 M6 에서 census 가 실제로 죽임. 처분: **open** — 비대화형 서브에이전트라 사람 답이 없음; 권고 「B) Keep — justified: Manager 결정 D6(verified: design.md D6 · census-repair 영수증)」. Manager 가 확정.

## Specialists (Dispatching 6 specialists: testing, maintainability, security, performance, simplification, state-machines. Skipped: data-migration, api-contract, design (scope not detected).) + red-team(DIFF_LINES>200)
- testing: NO FINDINGS · security: NO FINDINGS(fixtures summary mode) · performance: NO FINDINGS · simplification: NO FINDINGS(lean already) · red-team: NO FINDINGS
- maintainability: [INFORMATIONAL 6] reconcile_resolve_causes.go:24 — 계좌 전체 AF 거절 문구의 종목 자리가 빈칸(`on  requires …`).
- state-machines (`STATE_MATRIX_CHECK: ok machines=1 kinds=9 actions=1 cells=9 tested=6 untested=3`): 시험 안 된 칸 3 — 계좌 전체 QM 운영자 해제(production 런타임), SNAPSHOT_UNAVAILABLE·SNAPSHOT_STALE 거절 행.
- PR Quality Score: 8.0/10 (specialist: critical 0 실결함 — state-machines 의 "CRITICAL" 3은 시험 공백이라 리드 판단으로 INFORMATIONAL; informational 4 × 0.5).

## ADVERSARIAL REVIEW (Claude subagent) — 판정 「merge as-is」, 차단 0
1. INVESTIGATE Low — 블록은 build 때 Restore 로 한 번 읽고 약 4초 뒤 validate/Resolve 에 씀. 잠금 밖 writer 가 그 사이 행을 더하면 그 행은 안 풀리고(안전) 원인 중립 성공 문장은 「active RECONCILE blocks released」라 다 풀렸다고 읽힐 수 있음(개수 없음).
2. FIXABLE Low — maintainability 와 같은 발견(빈 종목 자리). MULTI-SOURCE CONFIRMED.
3. INVESTIGATE Low — 실제 투영 writer(`enterReconcileScopeInTx`)가 AF 행의 scope_market 을 NULL 로 쓰는 것을 고정하는 시험 없음(시험은 EnterReconcile 로 심음). 바뀌면 Resolve 배치가 「not active」로 롤백.
4. INVESTIGATE Low — 사고 실장부 상태에서 해제가 성공함은 미증명: 거부된 매도 주문이 여전히 추적(TrackedFillOrders)되면 MissingOrders 로 영구 거절 가능.
- Codex outside pass: 이번 1회차에서는 안 돌림 — codex 0.165.0 이 같은 diff 를 구현 적대 리뷰 r1(Manager 진행)에서 이미 봄. 이 /review 기록으로는 outside coverage 없음.

## Exploratory QA and Verification Results
- 표면: functional(CLI). 실 CLI(`tossctl engine reconcile-resolve`) 실행: **blocked** — Manager 지시(tossctl·운영 장부 금지).
- smoke(저장소 시험 명령, 임시 장부·로컬 TLS 가짜 서버): capture 001 `go test ./cmd/tossctl -run ^TestA195OperatorReleasesASymbolAttributionFailure$` exit 0, capture 002(checkpoint 002) `-run ^TestA195IdentifierConflictKeepsEveryRowBlocked$` exit 0. 증거 `.gstack/tmp/qa-reports/run-20261123T135514Z/functional/`(git 무시 경로).
- 남은 공백: 실 CLI·실 장부 미실행(위 blocked).

## Fix-First
- [AUTO-FIXED] a195 시험 원인 표에 SNAPSHOT_UNAVAILABLE·SNAPSHOT_STALE 거절 행, `TestA195AccountWideQuantityMismatchStillReleases`(production 런타임, 계좌 전체 QM → OPERATOR 해제·재복원 0) 추가 — 시험 전용, 선택 시험 ok.
- Manager 결정 대기(production 문구·범위): 빈 종목 자리 문구(D5 「종전 문구」 결정과 상충 가능), 성공 문장의 개수/범위 표시, AF writer scope_market 고정 시험(journal 패키지 시험 추가 = 범위 확장), 사고 실장부 상태 검증(운영 장부 접근 필요).
- 재리뷰: 안 함(Manager 지시 1회차). 편집이 있었으므로 converged=false.

Pre-Landing Review: 0 issues unresolved as defects (0 critical, 0 informational) — 결정 대기 4건은 결함이 아닌 Manager 판단 항목, gate 1건 open. COMPLETED: false(실 CLI probe blocked · Codex outside 미실행 · 편집 뒤 재리뷰 없음).
