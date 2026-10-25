# a173 tasks

Story: `STORY-TOS-a173` · Low risk(렌더·CSS·뷰 모델·시험). FLM/BTM·변이 원장·다중 적대 리뷰 `not-applicable`(proposal). 작업 위치: `worktree/a173`(브랜치 `feat/a173-console-ux-common-cleanup`, base `b27a1ea87`). 구현자는 편집 전에 `design.md` §3 의 시험을 열어 읽는다.

## 0. 정의
- [x] 0.1 Story 등록(FEAT-TOS-004) · change 정의 · base `b27a1ea87` · spec 델타(`specs/operator-console/spec.md`) · proposal-freeze 경량 리뷰(`review.md`)

## 1. 기준선·하네스
- [ ] 1.1 `measure/shoot.cjs`·`outline.py` 를 인자 기반(절대경로 인자, 하드코딩 0)으로 고치고 `measure/README.md`(실행 명령 2줄) 작성. 격리 콘솔(빈 설정 폴더·자격증명 0·버튼 0)로 **기준선** 촬영 → `measure/baseline-<utc>.json`(수치만, 스크린샷은 scratch)
- [ ] 1.2 알림 상자 전수 분류표 `analysis/notice-census.md`: `.notice` 108 + `.danger` 38 각 자리 → 차단/결과/본문 중 하나와 사유 한 줄

## 2. RED (계약 시험 먼저, 전부 FAIL 확인·출력 기록)
- [ ] 2.1 `TestEachUnmeasuredCauseIsExplainedOnce`: `consoleScreens` 전수 + 채운 fixture(`seedJournal`·`newOrdersHarness`·`brokerRows`)에서 `unmeasuredSentences` 각 문장 출현 ≤ 1, `미측정` 칸마다 `<code>` 사유 존재, 표본 0 이면 FAIL
- [ ] 2.2 `TestFloatingNoticesAreAtMostThree`: 전수 렌더에서 `class="notice`/`class="danger` 합 ≤ 3
- [ ] 2.3 `TestOperatorBodyCarriesNoDeveloperIdentifiers`: `details.tech` 밖 텍스트에 `a0\d{3}`·절대경로·64자 hex 0, 전체 HTML 에는 남아 있음
- [ ] 2.4 `TestEveryScreenHeadHasScopeBadgeAndFoldedBoundary`: `<h1>` 형제에 `scope` 배지 1개, `details.scope` ≤ 1, 머리 `<p>` 에 "이 화면은" 류 경계 문장 0
- [ ] 2.5 `TestStatusStripCellsAreOneLine`: 각 칸 `<small>` 텍스트 ≤ 40자·`<br>` 0·영문 원문 오류 0(`verify:`·`TOSSCTL_` 0)
- [ ] 2.6 CSS 시험: 산문 폭 규칙(`max-inline-size: 42rem`) 존재, 거래 정책 라벨 grid 규칙 존재

## 3. GREEN (design §2 순서)
- [ ] 3.1 §2.1 원인 1회 — 뷰 모델 `Causes` + `unmeasured` 두 모드 + 주문 브로커 원문 1회 + 상태 표시줄 사유 코드. 기존 Go 함수를 편집하면 함수 이름을 여기에 적고 `review.md` 면제 문구
- [ ] 3.2 §2.2 알림 3단계 — 분류표대로 class 교체, "반영 시점" 류 본문화
- [ ] 3.3 §2.3 경계 배지·접힌 경계 (모든 화면)
- [ ] 3.4 §2.4 상태 표시줄 압축 (CSS + `<small>` 문구), 모바일 칩
- [ ] 3.5 §2.5 기술 정보 define + 대상 이동(카테고리 라벨·digest·경로)
- [ ] 3.6 §2.6 줄 길이·제목 언어·'적용 후' 표 한 줄화 · §2.7 체크박스 간격
- [ ] 3.7 `gofmt -l`·`go vet ./internal/console/`·`go test ./internal/console/`(비캐시 1회)·`go test ./cmd/tossctl/ -run Console`. 실패는 전부 여기 기록하고 고친다. 고정 리터럴을 바꾼 시험은 아래 「고정 리터럴 시험 수정 기록」에 전수

## 4. 측정·리뷰
- [ ] 4.1 하네스 재촬영(22 × 1280/390 × light/dark) → `measure/after-<utc>.json`, 전후 표를 `design.md` §4 아래에 추가. 목표: 문장 1회·상자 ≤ 3·표시줄 ≤ 72/120px·식별자 0·산문 ≤ 70자(가장 긴 `<p>` 폭 ≤ 672px)
- [ ] 4.2 `cognitive-load-assessment` 22개 화면 표 재작성(`analysis/cognitive-load-after.md`) — 결정·기억·읽기·시각 "높음" 0
- [ ] 4.3 `design-critic` 6차원 재점수 + Impeccable(`critique`·`audit`) + `avoid-ai-design` detect → `analysis/review-after.md`, 발견 반영 여부 기록
- [ ] 4.4 `fixing-accessibility`·`fixing-motion-performance` 점검(변경 0 이어도 기록)
- [ ] 4.5 gstack `review`(경량) → `review.md` 추가. 변이 1회: `Causes` 배너 생성 제거 → 2.1 FAIL 확인 후 원복(기록)
- [ ] 4.6 Manager(Fable) 검증 요청 — 변경 파일·시험 출력(잘리지 않은 원문 경로)·전후 수치·스크린샷 디렉터리 보고. Manager 가 크롬으로 22개 화면 × 데스크톱/모바일/다크를 직접 본다

## 5. 종결 (Manager 승인 뒤, Opus 가 수행)
- [ ] 5.1 `docs/console-design.md` 에 알림 3단계·경계 배지·기술 정보·원인 1회 규칙 추가 확인 → main 병합·푸시(사람 승인 뒤) → PM tracker 재생성 → `make sdd-sync` → 아카이브 → memory retain

## 고정 리터럴 시험 수정 기록
(없으면 "없음" 이라고 쓴다)
