# a175 review

## proposal-freeze (2026-10-26, 경량: validate + Manager 셀프리뷰, docs/WORKFLOW.md 「위험 등급 가중」 UI change)

- 보이스: Manager(Opus 5.5) 단독. 적대적 Eng 보이스는 해당 없음. 주문·위험관리·원장·reconciliation 은 건드리지 않고, 프리셋 POST 는 preview 단계로 가는 기존 폼을 바이트 그대로 옮긴다.
- base: `b7d34acad`(a174 archive, origin/main 과 같음).
- 입력: Fable 5.1 정의(proposal·design·tasks.draft·spec 델타), HEAD `b7d34acad` 원문 grep(`templates_optimization.go`·`templates_signals.go`·`templates.go`·`optimization_view.go`·`internal/candidate/threshold_descriptors.go`·`internal/exitpolicy/common_policy.go`·`display_primitives_test.go`).
- 사람 결정 승계: worktree 형제 디렉터리, 구현자 GPT-5.6 Terra, WORKFLOW 비례 원칙 문구(기존 함수 편집 = 번들).
- 발견과 결정(정의본 대비 정정, design §1 ★):
  1. 익절 프리셋 컨테이너는 `.tier-card` 가 아니라 `.choice-tile` 이고, 폼은 preview 로 가는 1단계다. hover 금지 시험은 `display_primitives_test.go:565-581` 이다(`TestAmbientMotionContract` 아님). → 시험 확장 대상과 함수 이름을 정정했다.
  2. runner 의 T4 `TargetPct="999.0"` sentinel 이 지금 표에 `999.0%` 로 찍혀, 같은 타일의 "sentinel 을 노출하지 않는다" 문장과 모순된다(기존 결함). → 행렬은 "고정 목표 없음" 으로 표시하고 시험으로 고정한다(2.3).
  3. 중계 카테고리의 링크 수는 1·2·1 이다(전략·실행은 2). 그리고 성과·이력은 설정 변경 감사 이력 표를 가진다. → "링크 1" 단언을 "링크 수 불변" 으로 바꾸고, 감사 표는 유지한다(정보 손실 방지). 700px 기준에서 성과·이력은 뺐다.
  4. 후보 필터의 "공통 행" 은 값이 같다는 가정이다. → KR·US 값이 바이트 단위로 같을 때만 합치고, 다르면 시장별로 남긴다. 다른 값 fixture 로 시험한다(2.4).
  5. "기록만 하고 판정하지 않는다 6회" 는 렌더 기준 값이다(소스는 Go 문자열 1개) — 1.1 에서 렌더로 다시 센다.
  6. FLM 면제 문장을 거뒀다: 기존 함수(시험 함수 포함)를 고치면 함수마다 번들이 필요하다.
- 남은 위험: 신호 화면 정찰 값(3006px·공통 문장 수)은 Manager 가 다시 재지 않았다(구현자 1.2 기준선이 대신 잰다). `strategy-runtime` 두 열 배치가 한 시장 실패를 다른 시장 표시와 독립으로 유지하는지는 2.5 에 KR 실패 fixture 단언을 넣어 확인한다.
- 결론: proposal-freeze 승인. 착수 가능. 위치 `/mnt/D/Axipient/workspace/TossOS.worktrees/a175-optimization-and-runtime-matrix-views`. 정정본을 커밋한 **뒤에** worktree 를 만든다.

## gstack review 경량 (2026-10-26, 구현 후)

- 적용성: UI/UX 변경이므로 `impeccable`·`fixing-accessibility` 정적/시각 검토와 gstack 체크리스트의 SQL·동시성·신뢰 경계·enum·frontend 항목을 적용함. 전역 gstack은 `1.91.223.0` 업데이트 가능 상태였으나, 전역 도구 업데이트는 작업 범위를 벗어나 수행하지 않음. worktree의 `.claude/skills/review/` 경로는 없어서 같은 설치본의 `/home/daniel/.claude/skills/gstack/review/{checklist,greptile-triage}.md`를 읽었음.
- 범위 점검: handler·레지스트리·주문 경로·form action/hidden/button/disabled 계약 변경 없음. 새 view model은 읽기 전용 registry projection과 표시 문자열만 만듦. Greptile PR 댓글·Codex CLI 적대 검토는 경량 범위에서 실행하지 않음.
- 독립 검토 결과와 처리:
  1. 비개요 lifecycle 미확인을 `OFF`로 표시하던 경로를 `확인 불가 (lifecycle 미확인)`로 바꾸고 `TestOptimizationCategorySummaryFailsClosedForUnavailableLifecycle`로 고정함.
  2. 후보 필터 preview 계약이 `<details>` 안에 있던 것을 본문 `preview / CAS` 항목으로 옮기고, Help 설명만 접도록 `TestCandidateFilterMatrixFoldsDescriptionsOnly`를 강화함.
  3. 같은 KR·US descriptor의 공통 병합·`colspan`과 registry 유도 행·열 수를 `TestCandidateFilterMatrixMergesEqualMarketDescriptors`와 기존 DOM 시험으로 고정함.
  4. native table의 `aria-readonly` 존재 검사를 제거하고 실제 input/form/contenteditable 부재와 읽기 전용 문구·scope/region을 검사함.
  5. runtime의 720px 1열 계약을 `max-width: 45rem` CSS와 `TestRuntimeGridUsesSingleColumnAt720px`로 고정함.
  6. 기존 preset 시험이 비개요에 사라진 overview-only `data-*` 상태를 요구하던 불일치를 overview 렌더 검사로 옮김. 가짜 속성은 추가하지 않음.
- 검증: `analysis/review-red-output.txt`의 네 회귀 실패 후 `analysis/review-green-output.txt`·`analysis/review-contract-green-output.txt` 통과. 남은 정보성 위험은 현재 registry가 대칭인 KR·US key만 제공한다는 점임. 장래에 한 시장 전용 key를 도입하면 행렬 key 합집합과 명시적 unavailable cell을 별도 change에서 설계·시험해야 함.
- 결론: 미해결 critical 0. 이 change의 경량 리뷰 범위에서는 ship-blocking 발견 없음.

## P2 Manager 수정 라운드 (2026-10-26)

- 후보 필터 합집합: 행 key를 KR 전용이 아니라 registry 시장 순서의 first-seen 합집합으로 만들었음. 한 시장에 없는 key는 zero-value descriptor를 렌더하지 않고 `이 시장에 등록 없음`으로 표시하며 공통 행 병합에서 제외함. `TestCandidateFilterMatrixPreservesMarketOnlyKeys`가 KR 전용·US 전용 행 순서, 등록 상태, 실제 template unavailable cell을 검증함.
- 공통 행 비교: `Market`만 중립화한 descriptor 사본을 `reflect.DeepEqual`로 비교해 새 field 누락이 공통 병합으로 이어지지 않게 했음. `Category`만 다른 fixture는 분리를 검증함. 이전 수동 비교도 이미 `Category`를 포함했으므로 이 단독 fixture는 RED가 아니었고, P2 RED는 미등록 cell 상태가 없어서 실패했음.
- 신호 측정: 900px 기준선은 빈 격리 config에서 `discovery_unreadable`가 난 상태였음. 서로 다른 상태의 3243px/3640px 기록은 전후 비교에서 제외하고, base·HEAD 동일 상태를 900px/941px로 다시 측정했음. 시장 행 없는 오류 상태에만 불필요했던 공통 그림자 설명을 렌더하지 않도록 해 모바일 증가를 없앴으며 cause·안전 상태는 그대로 표시함. 상세값은 `analysis/p2-isolated-measurement.md`에 있음.
- 중계 높이: 데스크톱 전체 1218→1261px 증가는 `.optimization-nav` 513→710px가 정하며 본문은 세 카테고리 모두 감소했음. 링크·설명·여백을 줄이지 않았고, 완료 기준 조정은 Manager 판단으로 남김.
- P3 기록만: 후보 필터의 `미승인 — 숫자 값 없음`·`미승인 — verdict 비활성` 문구와 전략·실행의 `lane desired OFF · engine autostart OFF · scheduler desired OFF` 문구는 base부터 존재한 상태 비연동 고정 문구임. registry/runtime이 실제 승인·desired/effective를 제공하는 change에서 상태값으로 교체하지 않으면 표시와 값이 어긋날 위험이 남음.
- skill·research: UI 영향이 있어 `impeccable`을 적용했고 `analysis/p2-impeccable-detect.json`은 빈 배열이었음. 기존 Go `html/template`와 표 DOM helper를 확장하고 Go 표준 라이브러리 `reflect.DeepEqual`만 사용했으며, 새 의존성·UI 라이브러리·라이선스 검토 대상은 추가하지 않았음. 이 선택은 행렬의 실제 렌더·접근성 역할을 기존 시험 방식으로 계속 검증할 수 있는 최소 변경임.

## P3 프리셋 표·관련 화면 보정 (2026-10-26)

- Manager 4.5 실측에서 1280px preset region이 `scrollWidth/clientWidth = 956/879`로 overflow를 보였음. `.preset-compare`를 컨테이너 폭의 fixed layout으로 바꾸고 header `small`·policy ID는 줄바꿈을 허용했음. 설명·폼 문구·폼 action/hidden/disabled 분기는 바꾸지 않았고 390px의 `min-width: 36rem` 표 스크롤은 유지함.
- 관련 화면 세 링크는 category card와 다른 제목 있는 `<ul>` 평문 목록으로 바꿨음. 링크 수와 href는 그대로이며 nav 하단의 "관련 화면" 요구를 충족함.
- `TestPresetCompareDesktopStylesFillContainer`와 `TestOptimizationRelatedLinksAreLabelledPlainList` RED→GREEN을 남겼고, 격리 1280px 재측정에서 preset region `879/879`, 페이지 `1280/1280`을 확인함. 상세는 `analysis/p3-isolated-measurement.md`에 있음.
- research: 기존 stylesheet·Go HTML template·DOM test helper로 충분했고, 새 utility·외부 UI library·의존성·라이선스 검토 대상은 없었음. Impeccable은 기존 Operate 콘솔 refinement로 적용하며 새 시각 세계·PRODUCT.md를 만들지 않았음.
- UI 정적 검사: `analysis/p3-impeccable-detect.json`은 빈 배열임. a175 CSS 블록 안에서만 수정했고, 기존 preset form의 action·hidden field·버튼 문구·disabled 분기는 그대로임.

### 완료 게이트 실행 영수증
- 사전: main HEAD `526650e50`(origin/main `b7d34acad`, 로컬 1 앞) · W HEAD `efcd925f9` 확인 → Manager 편집 커밋 `a2eaf3458` → 5.1 체크·tracker 재생성 커밋 `d8ba4a1e3`(tracker `--check` rc 0) → `check_analysis.py --change a175-…` rc 0, required 5(번들 7) — `evidence complete or diff-proven exempt`. rebase 불필요(main 미이동).
- 격리 worktree `/mnt/D/Axipient/workspace/TossOS.worktrees/a175-close`(detached `d8ba4a1e3`, 2026-10-27). **`/tmp` 밖**에 둔 것은 사람 결정이다. 빈 `/tmp/.git` 은 Codex `workspace-write` 샌드박스가 명령마다 만든다(2026-10-26 strace 확정). 그래서 `/tmp` 아래 worktree 의 go build 가 VCS status 128 로 무작위 실패한다(a174 1 차 실패). 이번 네 회차 내내 `/tmp/.git` 이 있었지만 rc 128 은 0 건이다.
- 준비: `make sdd-infra` rc 0 → `python3 tools/sdd/sdd_sync.py --no-gbrain` 3 회 모두 codegraphcontext 단계만 300 초 시한 초과(advisory, CodeGraphContext 는 판정 밖) → `make sdd-check` 1 차 FAIL(sdd-test `test_codex_session_save` 5 건이 `save_session.py` 5 초 시한 초과, 당시 load average 195~460; 부하가 내려간 뒤 단독 7/7 PASS) → 재시도 rc 0 → `codegraph status .` `✓ Index is up to date`.
- 1 차 gate(02:29~02:54 KST) FAIL ⑦ `make test`: `tools/a112-mb-us-source` `TestTrimpathBuiltCollectorCompletesOfflineIdentityPreflightBeforeReader`(248.71s) — `prescribed build identity: running executable does not match the prescribed rebuild`. 같은 worktree·HEAD 에서 그 시험만 단독으로 돌리면 PASS(89.97s).
- 2 차 gate(03:22~04:14) ⑦ PASS, FAIL ⑧ `make test-seams`: 같은 패키지의 위 시험(211.35s)과 `TestTrimpathBuiltCollectorProcessHoldsBeforeExternalReadOnMissingCache`(203.20s)가 같은 문구로 실패(load 최대 약 15).
  - 1·2 차 원인: a112 rebuild 가 실제 시각 120 초 시한(`main.go:82` `context.WithDeadline(… time.Now().Add(overallBudget))`) 안에서 돈다. 그 rebuild 오류가 `identity.go:488` 에서 digest 불일치와 같은 "does not match" 문구로 덮인다. 병렬 `make test` 아래에서 시한을 넘긴 것으로 판단한다. a175 diff(`internal/console/*`·문서) 밖이고, 이 패키지는 a174 PASS 뒤 커밋 0 이다. 원인을 구분하는 수정은 a178 에서 한다.
- 3 차 gate(06:11~06:15, 시작 load 3.31) FAIL ⑥ `make sdd-check`: `CodeGraph status probe failed: … codegraph status . timed out after 15 seconds`. 직후 단독 실행은 3.38s rc 0. 원인은 gate 직전 CodeGraph 워밍업을 빠뜨린 것이다(a174 HANDOFF §4 에 기록된 함정, Manager 브리프 누락).
- 4 차 gate(사용자 승인 "1 진행"):
  - 조용함 확인: load 2.82, 다른 gate·go test 0.
  - 워밍업 `codegraph status .`: 30.57s·37.77s·46.71s·16.84s 로 기준(15 초 미만)을 못 넘기다가 6.84s·8.95s 로 연속 2 회 통과. 전부 rc 0, Pending 0.
  - 곧바로 `rtk proxy make -C /mnt/D/Axipient/workspace/TossOS.worktrees/a175-close gate CHANGE=a175-…`(07:53:37~08:22:12, 로그 `repo:` 는 격리 worktree) → **GATE PASS 11/11, rc 0**.
  - ⑦ 의 a112 패키지는 `(cached)` 다. 같은 입력의 앞선 단독 진단 PASS 를 Go 테스트 캐시가 재사용했다. ⑧ 에서는 실제로 돌아 251.583s PASS.
  - 1 초 간격 기록 1707 표본: `/tmp/.git` 1707/1707 있음, load 최대 10.82.
- 원문 로그(잘리지 않음): `/tmp/a175-close-logs/{check_analysis,sdd-infra,sdd-sync{,2,3},sdd-check{,2},codex-save-solo,gate,gate2,gate3,gate4,a112-mb-solo,codegraph-status-diag,warmup.summary,watch3,watch4,archive}.log` — 세션 스크래치, 커밋 안 함.
- archive: `openspec archive a175-optimization-and-runtime-matrix-views --yes`(1.4.1) → `operator-console: + 2 added`, `2026-10-26-a175-optimization-and-runtime-matrix-views`.
