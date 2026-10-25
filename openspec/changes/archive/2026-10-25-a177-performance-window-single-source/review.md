# a177 review

## proposal-freeze (2026-10-25, 경량: validate + Manager 셀프리뷰)

- 보이스: Manager(Fable 5.1) 단독(표시 문자열 출처 변경, High-risk 미접촉).
- 정찰(Sonnet, 읽기 전용) 결과를 proposal 「가설」의 사전 근거로 기록한다 — 구현자가 1.1 에서 코드로 재확인한다:
  - `templates_optimization.go:217` 리터럴 "최근 30일 · … · complete lineage · 읽기 전용" 은 데이터를 읽지 않는다.
  - `/performance-history` 칩 `templates_optimization.go:306,337,352` 는 `.View.Query.PeriodDays` 를 찍는다. `handlePerformanceHistory`(`performance_history.go:35-56`) 는 리더가 nil 이면 `Unwired`, 오류면 `LoadErr` 로 두고 `View` 는 **영값**이다. 칩 `<section>`(`:300-310`) 이 `{{if .Unwired}}…` 분기(`:312`) **앞**에 렌더되어 0일이 찍힌다. 성공 경로는 `performance.DefaultQuery`(`performance/query.go:36`, `PeriodDays: 30`) 로 30일.
  - 판정 방향: 의미는 같은 30일 기본값(H1), 그리고 **미배선·오류 상태에서 0일을 찍는 둘째 결함**(spec 706 "0으로 꾸미지 않는다" 위반). 표기 문구도 다르다("complete lineage" vs "complete lineage only").
  - 기존 시험 `performance_history_test.go:53,63` 은 wired 경로의 "최근 30일"·`PeriodDays == 30` 을 고정한다 — 유지된다.
- 결정: 리터럴 제거 + 두 화면이 같은 값 객체에서 문구를 만들고, 미배선·오류 상태는 두 화면 모두 "미측정 · 사유 코드" 로. 기본값 일수(30)는 바꾸지 않는다(사람 결정 불필요, H2 아님).
- 결론: proposal-freeze 승인. 구현 착수 가능(worktree/a177, Sonnet 5.5).

## validate 기록
- 2026-10-25: 델타 없음으로 FAIL → `specs/lane-performance/spec.md` ADDED 1건 추가 후 재실행(구현자가 1.1 전에 재확인해 여기 적는다).

## 구현 리뷰 (2026-10-25, gstack review 경량 — HEAD~1 diff 1건, 구현자 Sonnet 5.5)

범위: `performance_history.go`·`optimization.go`·`optimization_view.go`·`templates_optimization.go`·`performance_window_test.go`. 파일 수정 없이 diff 만 읽음. 스킬의 Greptile·전문가 군단·Codex 적대 단계는 인자(경량)에 따라 생략 — `not-applicable`(Low risk, 표시 문자열 출처 변경, 비례 원칙).
Scope Check: CLEAN. 의도=창 표기 한 출처 + 미측정 표기 / 실제=동일. SQL·동시성·LLM·셸·enum 확장 해당 없음. 테스트 stub 제안 없음.

| # | 발견 | 신뢰 | 판단 |
|---|------|------|------|
| 1 | P3 `optimization.go` 최적화 화면은 Dashboard 를 부르지 않으므로 **LoadErr 상태에서도 30일 창을 그대로 표기**(성과 이력 화면도 LoadErr 시 창은 시도한 기본 질의로 표기, 오류는 아래 `danger` 줄). review.md 사전 결정("Unwired/LoadErr → 미측정")과 다름 | 9/10 | **수용(편차 기록)**. 창은 서버 상수(DefaultQuery)이지 측정값이 아님; LoadErr 로 0일이 찍히던 결함은 해소(Window 는 View 영값이 아니라 DefaultQuery 에서 만듦). 최적화 화면이 창 표기만을 위해 DB 를 한 번 더 읽는 비용·결합이 더 나쁨. 사람이 LoadErr 도 미측정을 원하면 별도 결정 |
| 2 | P3 `newPerformanceWindow` 의 `all lineage`·비기본 Market/Lane 분기는 현재 `DefaultQuery` 로는 도달 불가 | 8/10 | **수용**. 표시 헬퍼가 질의 필드를 정직히 옮기는 최소 형태이며 0일/침묵 거짓을 막음. 시험은 기본값 경로만 고정 — 분기 시험 없음을 알림(미측정 아님: 도달 불가 코드) |
| 3 | P3 wired 하위시험은 리터럴을 되돌려도 통과(리터럴 문구를 새 정본 문구와 같게 맞췄으므로) | 9/10 | **수용**. 변이는 unwired 하위시험이 잡음(2.3 기록). wired 하위시험은 두 화면의 문구 동일성(30일 포함)을 고정 |
| 4 | 정보 최적화 화면의 문구가 "complete lineage" → "complete lineage only" 로 바뀜 | 10/10 | 의도(두 화면 문구 통일, `performance_history_test.go:53` 고정 문구 유지) |

반영할 수정 없음. 기존 시험 약화·삭제 0.

### Manager 검증 (Fable, 2026-10-25)
수용. LoadErr 편차 수용 사유: 창은 요청 상수(performance.DefaultQuery)이지 측정값이 아니며 실패 사유는 성과 이력 화면의 기존 danger 줄이 1회 말한다(spec 706 의 '0으로 꾸미지 않는다' 는 unwired 경로에서 충족, LoadErr 경로는 더 이상 0일을 찍지 않음 — loaderr 하위시험으로 고정). 재실행: gofmt 0 · vet OK · go test ./internal/console/ -count=1 ok.

### FLM 면제 (비례 원칙)
기존 함수 본문 편집: `(*Console).handlePerformanceHistory`(page.Window 대입 1줄 추가), `(*Console).handleOptimization`(page.PerformanceWindow 대입 1줄 추가). 콘솔 읽기 전용 렌더이며 High-risk 경로·게이트 판정을 바꾸지 않으므로 FLM/BTM not-applicable(docs/WORKFLOW.md 「비례 원칙」). 분기 추가 0.

## 완료 게이트 (2026-10-25, 종결 에이전트 Opus 5.5)

### gate ⑤ — FLM 면제 문구가 받아들여지지 않음 → 번들 3 작성
- 사전 실행 `python3 tools/logic-map/check_analysis.py --change a177-performance-window-single-source` → rc 1:
  `missing Function Logic Map for 3 function(s) modified between base b27a1ea87d35 and working tree …: internal/console/optimization.go:Console.handleOptimization, internal/console/performance_history.go:Console.handlePerformanceHistory, internal/console/performance_history_test.go:fakePerformanceReader.Dashboard`.
- 원인: 면제 표지(`Function Logic Map: not-applicable`)는 required 가 0 일 때만 쓰인다(`tools/logic-map/check_analysis.py` `_verdict`). 기존 함수 본문을 한 줄이라도 고치면 비례 원칙과 무관하게 번들이 요구된다. 위 「FLM 면제」 절은 표지 철자도 없었고, 있었더라도 통과하지 못했다. 시험 가짜 `fakePerformanceReader.Dashboard`(오류 주입 분기 추가)도 required 에 든다.
- 처분: 면제 문장을 고치는 대신(뜻을 바꾸지 않고는 통과 불가) `analysis/function-logic/` 에 번들 3(ast.json · function-logic-map.md · branch-test-map.md · risk-pattern-report.md)을 만들었다. 분기 행의 시험 인용은 시험별 `go test -run '^<Test>$' -coverprofile` 로 분기 본문 블록 count>0 을 측정한 것이다. 측정상 시험이 없는 기존 분기 3(`handleOptimization` B1 405 · B5 legacy Load 오류 · B13 protection List 오류)은 a177 무관 기존 공백으로 적었다. 재실행 → `evidence complete or diff-proven exempt` rc 0.
- 사후 작성임을 기록한다(FLM 이 구현 판단에 쓰이지 않음). a177 편집은 두 핸들러에 분기 밖 직선 대입 한 문장씩이라 AST 분기 수는 편집 전후 같다(14 · 5).

### 완료 게이트 실행 영수증
- 1 차(main 체크아웃, HEAD `d4dca0062`, 19:54) FAIL 6 단계: `[index-freshness] FAIL: CodeGraph hard-evidence index is unusable: CodeGraph status probe failed: … timed out after 15 seconds`. 같은 시각 다른 세션의 `codegraph sync`(27 분 실행, WAL 1.39 GB)와 load average 100+ 가 겹침. `make sdd-check` 단독 재실행도 같은 타임아웃(2 회째).
- Manager 진단: `codegraph status` 가 중단된 실행의 미해소 참조와 `Pending Changes: Removed` 를 보고. `codegraph sync .`(603.9 s) → `Removed: 8949`, 색인 파일 11,960 → 3,011, status 6.6 s `✓ Index is up to date`. 임계값 · 스크립트 수정 0. 살아 있는 `serve --mcp` 프로세스는 하나도 안 죽임.
- 2 차(main, 21:10) `make sdd-check` OK, 7 단계 `make test` FAIL — a177 무관 경로 사본 문제: 저장소를 훑는 구조 시험 9 개 꾸러미(`internal/app/engine` · `candidate` · `execgw` · `official` · `strategyengine` · `strategymarket` · `strategyrouter` 등)가 저장소 **안의** 다른 체크아웃(`worktree/a173` · `worktree/a177` · 다른 세션의 `.worktrees/agents-review` · `.worktrees/agents-writer`)의 사본까지 셌다(예: `non-test callers of RecordAlert = [.worktrees/agents-review/… worktree/a173/… internal/obs/record_only.go:… …]`).
- 3 차(격리 worktree `/tmp/tossos-a177-close-20261025`, detached `d4dca0062`): `make sdd-infra` rc 0 → `python3 tools/sdd/sdd_sync.py --no-gbrain` rc 0(`Indexed 3,010 files`) → `make sdd-check` rc 0 → `rtk proxy make gate CHANGE=a177-performance-window-single-source` → **GATE PASS 11/11, rc 0**(22:10). 원문 로그는 세션 스크래치패드 `a177/gate.log`.
- archive: `openspec archive a177-performance-window-single-source --yes` → `lane-performance: + 1 added`, `2026-10-25-a177-performance-window-single-source`.
