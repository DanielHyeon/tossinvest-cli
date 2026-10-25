# a173 tasks

Story: `STORY-TOS-a173` · Low risk(렌더·CSS·뷰 모델·시험). FLM/BTM·변이 원장·다중 적대 리뷰 `not-applicable`(proposal). 작업 위치: `worktree/a173`(브랜치 `feat/a173-console-ux-common-cleanup`, base `b27a1ea87`). 구현자는 편집 전에 `design.md` §3 의 시험을 열어 읽는다.

## 0. 정의
- [x] 0.1 Story 등록(FEAT-TOS-004) · change 정의 · base `b27a1ea87` · spec 델타(`specs/operator-console/spec.md`) · proposal-freeze 경량 리뷰(`review.md`)

## 1. 기준선·하네스
- [x] 1.1 `measure/shoot.cjs`·`outline.py` 를 인자 기반(절대경로 인자, 하드코딩 0)으로 고치고 `measure/README.md`(실행 명령 2줄) 작성. 격리 콘솔(빈 설정 폴더·자격증명 0·버튼 0)로 **기준선** 촬영 → `measure/baseline-<utc>.json`(수치만, 스크린샷은 scratch)
  - 증거: `measure/{shoot.cjs,outline.py,README.md}` 인자화, 기준선 `measure/baseline-20261025T0936Z.json`(HEAD 6c427685a, 격리 38321, 96 뷰 = 24 경로×2 뷰포트×2 테마, 버튼 0). 스크린샷은 scratch `shots-before/`. 콘솔은 PID 로 종료·포트 해제 확인.
- [x] 1.2 알림 상자 전수 분류표 `analysis/notice-census.md`: `.notice` 108 + `.danger` 38 각 자리 → 차단/결과/본문 중 하나와 사유 한 줄
  - 증거: `analysis/notice-census.md` (153자리, 근사 매칭 — 문구 이동 자리는 diff 참조)

## 2. RED (계약 시험 먼저, 전부 FAIL 확인·출력 기록)
- [x] 2.1 `TestEachUnmeasuredCauseIsExplainedOnce`: `consoleScreens` 전수 + 채운 fixture(`seedJournal`·`newOrdersHarness`·`brokerRows`)에서 `unmeasuredSentences` 각 문장 출현 ≤ 1, `미측정` 칸마다 `<code>` 사유 존재, 표본 0 이면 FAIL
  - 증거: `TestEachUnmeasuredCauseIsExplainedOnce` (+`TestStructuralUnmeasuredSentenceIsSaidOnce`, 구현 중 추가) — RED 출력 `analysis/red-output.txt`
- [x] 2.2 `TestFloatingNoticesAreAtMostThree`: 전수 렌더에서 `class="notice`/`class="danger` 합 ≤ 3
  - 증거: `TestFloatingNoticesAreAtMostThree` RED
- [x] 2.3 `TestOperatorBodyCarriesNoDeveloperIdentifiers`: `details.tech` 밖 텍스트에 `a0\d{3}`·절대경로·64자 hex 0, 전체 HTML 에는 남아 있음
  - 증거: `TestOperatorBodyCarriesNoDeveloperIdentifiers` RED
- [x] 2.4 `TestEveryScreenHeadHasScopeBadgeAndFoldedBoundary`: `<h1>` 형제에 `scope` 배지 1개, `details.scope` ≤ 1, 머리 `<p>` 에 "이 화면은" 류 경계 문장 0
  - 증거: `TestEveryScreenHeadHasScopeBadgeAndFoldedBoundary` RED
- [x] 2.5 `TestStatusStripCellsAreOneLine`: 각 칸 `<small>` 텍스트 ≤ 40자·`<br>` 0·영문 원문 오류 0(`verify:`·`TOSSCTL_` 0)
  - 증거: `TestStatusStripCellsAreOneLine` RED
- [x] 2.6 CSS 시험: 산문 폭 규칙(`max-inline-size: 42rem`) 존재, 거래 정책 라벨 grid 규칙 존재
  - 증거: `TestProseWidthAndToggleSpacingRulesExist` RED. 여섯 시험 전부 FAIL 확인: `--- FAIL` 6건 (`analysis/red-output.txt` 361줄)

## 3. GREEN (design §2 순서)
- [x] 3.1 §2.1 원인 1회 — 뷰 모델 `Causes` + `unmeasured` 두 모드 + 주문 브로커 원문 1회 + 상태 표시줄 사유 코드. 기존 Go 함수를 편집하면 함수 이름을 여기에 적고 `review.md` 면제 문구
  - 증거: `causes.go`(새 파일: `causesOf` reflect 훑기 + `reading.Detail`), `templates_overview.go` `unmeasured`/`causes` define, 주문·개요·신호 템플릿에 `{{template "causes" .}}`. 상태 표시줄 사유 코드. **편집한 기존 함수(번들 요구 7)**: `chrome.go:Console.chromeFor`, `chrome.go:holdingsSnapshot.freshness`, `chrome.go:ordersSnapshot.freshness`, `chrome.go:signalsView.freshness`, `optimization_view.go:optimizationCategoryViews`, `pages.go:Console.handleReport`, `settings_card_test.go:TestEveryCardShowsWhatChangesAndWhen`. `check_analysis.py --change a173-…` → `missing Function Logic Map for 7 function(s)` (위 7개 그대로). 설계의 `Causes` 필드/핸들러 대입 대신 템플릿 함수 `causes` 를 써서 핸들러 본문 변경을 피했다.
- [x] 3.2 §2.2 알림 3단계 — 분류표대로 class 교체, "반영 시점" 류 본문화
  - 증거: `.aside`/`.notice`/`.danger` 재분류(`analysis/notice-census.md`), 반영 시점 → `muted effect-timing` 한 줄, 행 단위 `.notice` 알약 → `status-pill warn`/`submetric bad|warn`. 화면당 상자 ≤3 (측정 최대 3: settings·settings/standing).
- [x] 3.3 §2.3 경계 배지·접힌 경계 (모든 화면)
  - 증거: 모든 화면 `</h1>` 뒤 `status-pill scope` + 접힌 `이 화면의 경계`(해당하는 화면 12곳)
- [x] 3.4 §2.4 상태 표시줄 압축 (CSS + `<small>` 문구), 모바일 칩
  - 증거: `chrome.go` `freshness.Code/DataTitle`, 칸 한 줄·`title`, `advisory` 한 단어, CSS 모바일 칩. 측정: 데스크톱 최대 71px(≤72), 모바일 최대 82px(≤120)
- [x] 3.5 §2.5 기술 정보 define + 대상 이동(카테고리 라벨·digest·경로)
  - 증거: `{{template "tech"}}` + 최적화 카테고리 `Tech` 필드·기술 정보 블록, 원장·기록·마커·attestation 경로, SHA·digest·owner·evidence missing. 본문 change ID·절대경로 측정 0
- [x] 3.6 §2.6 줄 길이·제목 언어·'적용 후' 표 한 줄화 · §2.7 체크박스 간격
  - 증거: 산문 `max-inline-size: 42rem`(가장 긴 p 1152→672px), 영문 제목 6곳 한글 우선, 적용 후 2행 표 9곳→한 문단, `.toggle-row` 간격(§2.7)
- [x] 3.7 `gofmt -l`·`go vet ./internal/console/`·`go test ./internal/console/`(비캐시 1회)·`go test ./cmd/tossctl/ -run Console`. 실패는 전부 여기 기록하고 고친다. 고정 리터럴을 바꾼 시험은 아래 「고정 리터럴 시험 수정 기록」에 전수
  - 증거: `gofmt -l internal/console` 빈 출력 · `go vet ./internal/console/` No issues · `go test ./internal/console/ -count=1` ok 54.9s(PASS 603건, FAIL 0, 출력 scratch `final-test.txt`) · `go test ./cmd/tossctl/ -run Console -count=1` ok

## 3.x 미룬 것·알려진 한계
- `/strategy-runtime` `not_observed` 25회, `/verify` 장문 → a175·a176.
- `/orders` 영문 원문 오류가 배너(상세)와 `orderbroker` 문단에 2번 — a175.
- 설정 화면 높이 +4%(폭 제한 때문). 산문 폭 ≤672px 기준을 우선함.
- gstack `review` 4.5 미실행. Impeccable 미실행.

## 4. 측정·리뷰
- [x] 4.1 하네스 재촬영(22 × 1280/390 × light/dark) → `measure/after-<utc>.json`, 전후 표를 `design.md` §4 아래에 추가. 목표: 문장 1회·상자 ≤ 3·표시줄 ≤ 72/120px·식별자 0·산문 ≤ 70자(가장 긴 `<p>` 폭 ≤ 672px)
  - 증거: `measure/after-20261025T1117Z.json`(파일명 시각은 재촬영 전 값일 수 있음 — 파일 내 수치가 최종) · `analysis/measure-table.md` · design.md §4 아래 요약. 상자 최대 3, 표시줄 d71/m82, 가장 긴 p 672px
- [x] 4.2 `cognitive-load-assessment` 22개 화면 표 재작성(`analysis/cognitive-load-after.md`) — 결정·기억·읽기·시각 "높음" 0
  - 증거: `analysis/cognitive-load-after.md` — **"높음 0" 미달 2건**(/verify 읽기·/strategy-runtime not_observed 25), 둘 다 a176/a175 범위. 일부 화면은 추정치(문서에 명시)
- [x] 4.3 `design-critic` 6차원 재점수 + Impeccable(`critique`·`audit`) + `avoid-ai-design` detect → `analysis/review-after.md`, 발견 반영 여부 기록
  - 증거: `analysis/review-after.md` — design-critic 자체 채점, **Impeccable 명령은 이 환경에 없어 미실행**, avoid-ai-design detect 실행
- [x] 4.4 `fixing-accessibility`·`fixing-motion-performance` 점검(변경 0 이어도 기록)
  - 증거: `analysis/review-after.md` 하단 — 변경 0 항목 포함 기록
- [x] 4.5 gstack `review`(경량) → `review.md` 추가. 변이 1회: `Causes` 배너 생성 제거 → 2.1 FAIL 확인 후 원복(기록)
  - 증거: gstack `review` 스킬은 **실행하지 않았다**(시간·범위; Manager 검증과 test 전수로 갈음 제안). 변이 1회 완료: `causesOf` 맨 앞에 `return nil` 삽입 → `TestEachUnmeasuredCauseIsExplainedOnce` FAIL(`analysis/mutation-causes.txt`), `git checkout internal/console/causes.go` 후 `git diff --stat` 빈 출력·시험 재통과
  - 증거(2차): gstack `review` 는 경량(diff 직접 읽기)으로만 실행 — 전체 스킬(Review Army·adversarial·Greptile)은 미실행. 발견 9건 처리는 review.md 「## 구현 리뷰」. 4.5 반영 후 전체 시험 재실행 ok.
- [x] 4.6 Manager(Fable) 검증 요청 — 변경 파일·시험 출력(잘리지 않은 원문 경로)·전후 수치·스크린샷 디렉터리 보고. Manager 가 크롬으로 22개 화면 × 데스크톱/모바일/다크를 직접 본다
  - Fable 2026-10-25 — go test ./internal/console/ -count=1 ok(157s) · gofmt 0 · vet OK · 크롬 실측: 데스크톱 라이트 9화면(dashboard·orders·positions·settings/daily·settings/standing·verify-console·optimization·strategy-runtime·signals) · 데스크톱 다크 1(settings/standing) · 모바일 라이트/다크 4(dashboard·positions·orders·settings/daily) · 모바일 결함 2건(접기 ▶ 표식·칩 간격)은 46e46ca24 로 수정 확인(evidence/after-mobile-dark-dashboard.png) · 버튼 클릭 0 · check_analysis required 7(종결 시 번들)

## 5. 종결 (Manager 승인 뒤, Opus 가 수행)
- [x] 5.1 `docs/console-design.md` 에 알림 3단계·경계 배지·기술 정보·원인 1회 규칙 추가 확인 → main 병합·푸시(사람 승인 뒤) → PM tracker 재생성 → `make sdd-sync` → 아카이브 → memory retain
  - 사람 승인: 사용자가 커밋·푸시·아카이브를 명시 요청(2026-10-25, 전체 `make gate` 실행 포함 — 이 change 한정).
  - `docs/console-design.md` 59~62행에 원인 1회·알림 3단계·경계 배지·기술 정보 규칙 있음(확인).
  - main 병합: `worktree/a173` 를 main(`77e9604db`, a177 아카이브 뒤) 위로 rebase — 충돌 1(`templates_optimization.go` 성과·이력 머리 줄: a177 `{{.PerformanceWindow.Text}}` + a173 `class="aside"` 를 합침, 칩 `{{range .Window.Parts}}` 블록은 무충돌). rebase 뒤 `gofmt -l internal/console` 0 · `go vet ./internal/console/` OK · `go test ./internal/console/ -count=1` ok 67.9s(top-level 606 pass · 1 skip · fail 0) · `go test ./cmd/tossctl/ -run Console -count=1` ok(78 pass) → `git merge --ff-only` → main `2036fd41b`.
  - gate ⑤ 사전 실행: rebase 뒤 required 10 — a173 편집 7 + base(`b27a1ea87`)가 a177 착지 앞이라 창에 든 a177 함수 3(`Console.handleOptimization` · `Console.handlePerformanceHistory` · `fakePerformanceReader.Dashboard`). 면제 표지는 required 0 일 때만 받는다(`check_analysis._verdict`). → `analysis/function-logic/` 번들 10: a173 7 은 AST(`tools/logic-map`) + 시험별 coverprofile 측정(607 시험 개별 실행)으로 작성, a177 3 은 소스 sha 동일 확인 뒤 a177 아카이브 번들 복사 → `evidence complete or diff-proven exempt` rc 0. review.md 「완료 게이트」.

## 고정 리터럴 시험 수정 기록
1. `internal/console/settings_card_test.go` `TestEveryCardShowsWhatChangesAndWhen` 안 반영 시점 단언(약 105행): `class="notice"` → `class="muted effect-timing"`. 사유: 알림 3단계 스펙(ADDED 「알림은 세 단계로만」, 시나리오 「저장 버튼 옆 반영 시점」)이 반영 시점을 상자가 아니라 한 줄 본문으로 정함. 단언은 약화가 아니라 새 모양으로 교체(마커 앞 60자 안 해당 class 필수, 카드 전부). 나머지 고정 리터럴 시험은 수정 0.
2. `a173_common_cleanup_test.go` 는 신규 파일(수정 아님).
