# a209 review

## Pre-Edit Gate (2026-11-28)

- change id / task id: a209-console-exit-evidence-counts-arming-rows-only / 2.2
- 대상 심볼: `journal.ReadOnly.BrokerOrderExitLinks`(internal/journal/account_views.go)
- CodeGraph 1.6.0: callers — 생산 `Console.ledgerOrderEvidence`(internal/console/orders.go:716) 하나, 나머지는 journal 시험
  (`account_exit_orders_test.go` 16 · `a207_compact_evidence_test.go` 1). `affected --filter '*_test.go'` 는 파일 단위라 1,267(journal 패키지 전체)로
  퍼져 판단에 쓰지 않음. a166 census `shapes.json` 에 account_views.go 없음.
- CodeGraphContext: not-applicable — 대상이 한 함수의 SQL 문자열 두 줄이고 호출자가 하나라 문맥 확장 후보가 없음. 증거 조정은 design §1 · §4.
- 기존 동작 근거: 편집 전 AST `analysis/ast-evidence/journal--ReadOnly.BrokerOrderExitLinks.base.ast.json`(28 갈래), RED 출력 `analysis/tdd/red-journal.txt`.
- FLM / BTM: `analysis/function-logic/internal-journal--readonly.brokerorderexitlinks/`
- upstream 상속 시험 영향: no(TossOS 고유 리더)
- 실패 시험 선행: yes(RED 3 실패 — FILLED · REFUSED · CANCELLED)
- 설정 · DB · journal 변경: 없음(읽기 SQL 만, 스키마 0)
- §0 안전 불변식: 통과 — 표시 전용 리더, 주문 · 손절 · 원장 쓰기 · 대사 경로 무변경

## 비례 원칙

- 변이 원장: not-applicable — High-risk 경로 · 게이트 판정을 바꾸지 않는 표시 리더(docs/WORKFLOW.md 「비례 원칙」).
- 다중 적대 리뷰: not-applicable — 같은 사유. 경량 gstack `review` 1회(상한 2회).

## 경량 gstack 리뷰 r1 (2026-11-28 — 상한 2회 중 1회)

- 도구: gstack `review` 스킬(1.98.0.0). Review Army — Testing · Maintainability · Performance 세 specialist + Claude 적대 패스 1(모두 별도 컨텍스트
  subagent, 읽기 전용 · 시험 실행 금지 · 저널 접근 금지 · 새 가드 · 범위 제안 금지). Security · Simplification 은 비례 원칙으로 생략(Go 변경이 SQL
  두 줄 + 시험 파일, 이어 붙이는 것은 컴파일 시점 상수뿐). Codex 외부 패스 생략 — 작성자와 다른 계열 보이스가 없으므로 **외부 커버리지 누락**으로 적음.
- Gate scan: `GATE_SUMMARY: listed=1 eligible=0` — 새 시험 파일 하나, 완화된 게이트 0.

| # | 출처 | 지적 | 처분 |
|---|---|---|---|
| 1 | testing (INFO, 6/10) | design D3(훅 행만 → 연결 없음)과 D1(NULL action 은 무장 행으로 셈 — `coalesce` 가 없으면 `NULL NOT IN` 이 거짓이라 조용히 빠짐)에 시험이 없음 | 수리. `TestA209HookOnlyIsUnlinkedAndANullActionRowStillCounts` 추가(새 함수 — 기존 시험 함수 무편집). 변이: 두 자리의 `coalesce` 를 빼면 `null action beside an arm` 갈래가 실패함(Ambiguous:false), 복원 후 sha 8351a432b348 = FLM ast.json 과 같음 |
| 2 | maintainability · performance | NO FINDINGS(성능: `idx_exit_events_proposed_intent` 로 intent 당 ≈2 행 action 조회만 늚) | — |
| 3 | 적대 #1 | 커밋된 `verify-journal.txt` 가 실행 중 빈 파일이고 `verify-journal-timeout10m.txt` 는 기본 10 분 시한 FAIL(`TestClaimStrategyDispatchLeasePairedKRUS…` 실행 중 — 무관) → 회귀 통과 증거 0 | 수리. Makefile 과 같은 `-timeout 30m` 으로 재실행해 결과를 아래 「시험 기록」에 적음. 10 분 기록은 시한 초과의 증거로 남김 |
| 4 | 적대 #2 | a207 의 exit-policy 델타(「orders 연결의 연결 행 수 판정은 바뀌어서는 안 된다」 · 「압축 행 + 체결 훅 → 모호」 Scenario)와 정면 충돌 | 이 change 에서 고치지 않음(a207 은 다른 writer). design §5-2 열린 질문 · tasks 3.3 — a207 archive 전 결정 필요 |
| 5 | 적대 #3 | a207 배포 뒤에는 체결된 청산 주문이 압축 무장 행이면 「반복 거절(요약)」 문구가 뜸 — 앞으로 흔해질 모양 | design §5-1 에 보강(사람 결정). 기본은 그대로 |
| 6 | 적대 #4 | 훅 목록 사본이 `a166_replay_origin.go` Go switch 에 따로 있음(기존) | 보고만. a209 가 만든 사본이 아니고, 새 훅이 생겨 갈리면 SQL 쪽은 무장 행으로 세어 모호(fail-closed)로 떨어짐 |
| 7 | 적대 #5 | 훅 행만 intent 의 표시가 legacy_event → `exit_evidence_unlinked` 로 바뀜 | 의도(design D3), 운영 0 건. 1 의 시험이 못 박음 |

- 적대 패스 권고: 「병합은 막지 않되 선택 시험을 먼저 돌려 기록할 것」. 코드 결함으로 막을 것 0.
- 2회차는 열지 않음 — r1 의 수리는 시험 함수 하나 추가와 문서뿐이고 생산 코드는 그대로임.

## 시험 기록 (2026-11-28, worktree `~/tossos-work/a209`, ext4)

| 명령 | 결과 | 원문 |
|---|---|---|
| `rtk proxy go test ./internal/journal/ -run TestA209 -count=1`(RED, 수정 전) | FAIL — FILLED · REFUSED · CANCELLED 셋이 `Ambiguous:true UnknownReason:ambiguous_exit_evidence`, 무장 둘 시험은 통과 | `analysis/tdd/red-journal.txt` |
| 같은 명령(GREEN, 수정 뒤) | ok 1.311s | `analysis/tdd/green-journal.txt` |
| 같은 명령(리뷰 r1 시험 추가 뒤) | ok 4.405s — 4 시험 · 하위 4 PASS | `analysis/tdd/green-journal-r1.txt` |
| `coalesce` 두 자리 제거 변이 → `-run TestA209HookOnly` | FAIL(`null action beside an arm` — Ambiguous:false) → 잡힘. `git checkout` 복원, sha 8351a432b348 확인 | 이 표 |
| `rtk proxy go test ./internal/journal/ -count=1`(기본 10 분 시한) | **FAIL — `panic: test timed out after 10m0s`**, 실행 중이던 시험 `TestClaimStrategyDispatchLeasePairedKRUSRefusesExpiredChangedCrossMarketAndStaleOwner`(무관). 시한 초과이며 단언 실패 아님 | `analysis/tdd/verify-journal-timeout10m.txt` |
| `rtk proxy go test -timeout 30m ./internal/journal/ -count=1`(Makefile 과 같은 시한) | ok 764.265s | `analysis/tdd/verify-journal.txt` |
| `rtk proxy go test -timeout 30m ./internal/console/ -count=1` | ok 126.418s | `analysis/tdd/verify-console.txt` |
| `rtk proxy make lint`(gofmt + `go vet ./...` + `go vet -tags tossos_testseams ./...`) | rc 0 | `analysis/tdd/verify-lint.txt` |
| `python3 tools/logic-map/check_analysis.py --change a209-…` | rc 0 — required 1(`ReadOnly.BrokerOrderExitLinks`), evidence complete | — |
| `openspec validate a209-… --strict` · PM `--check` | valid · current | — |

- 30 분 journal 실행은 리뷰 r1 시험 함수 추가 **전** 트리에서 시작함(생산 코드는 같음). 추가한 시험은 위 `-run TestA209` 로 따로 돌았고 lint 는 추가 뒤 트리에서 돎.
- 선택 이유: 바뀐 생산 함수는 journal 하나, 생산 호출자는 console 하나(CodeGraph callers). a166 census 에 account_views.go 없음 → `internal/accountbudgetruntime` 불필요.
- 실행하지 않은 범위: `make test`(전체) · `make test-seams` · `make test-race` · `make gate`(배포 뒤 · 사용자 요청 시). 화면 확인(규칙 13)은 배포 뒤 tasks 3.2.
