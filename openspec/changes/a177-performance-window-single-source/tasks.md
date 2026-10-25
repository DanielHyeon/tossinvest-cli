# a177 tasks

Story: `STORY-TOS-a177` · Low risk(표시 문자열 출처). FLM/BTM·변이 원장·다중 적대 리뷰 `not-applicable`(proposal). 작업 위치: `worktree/a177`(브랜치 `feat/a177-performance-window-single-source`, base `b27a1ea87`).

## 0. 정의
- [x] 0.1 Story 등록(FEAT-TOS-004) · change 정의 · base `b27a1ea87` · proposal-freeze 경량 리뷰(`review.md`)

## 1. 근본 원인 (코드로 확정, 추측 금지)
- [x] 1.1 `templates_optimization.go:217` 리터럴의 데이터 타입·핸들러와 `/performance-history` 기본값 칩의 필드·출처를 file:line 으로 적는다(design.md 「조사 순서」)
- [x] 1.2 "0일" 이 의도된 기본값인지 빈 값의 가면인지 판정하고 proposal 「가설」을 H1/H2 로 확정. H2 면 `issues.md` 에 사람 결정 항목(일수 변경)을 적고 값은 바꾸지 않는다

## 2. RED → GREEN
- [x] 2.1 RED: 두 화면 렌더 비교 시험(같은 fixture → 같은 조회 창 문구; 빈 값 fixture → 같은 미측정 형식; 렌더 길이 > 0 단언). 실패 출력을 여기에 기록
- [x] 2.2 GREEN: 공용 헬퍼 + 카테고리 뷰 모델 필드, 템플릿 리터럴 삭제. `gofmt -l`·`go vet ./internal/console/`·`go test ./internal/console/ -run 'Performance|Optimization'` 통과
- [x] 2.3 변이: 리터럴 되돌림 → 시험 FAIL 확인 후 원복(심볼 수 대조). 기록

## 3. 검증·리뷰
- [x] 3.1 `go test ./internal/console/` 전체(비캐시 1회) + `cmd/tossctl` 콘솔 관련 시험. 실행 범위·미실행 범위 기록
- [x] 3.2 렌더 확인: 격리 콘솔(빈 설정 폴더, 자격증명 0, 버튼 0)로 두 화면을 띄워 같은 문구인지 스크린샷 2장(`openspec/changes/a177-…/evidence/`)
- [x] 3.3 gstack `review`(경량) → 결과를 `review.md` 에 추가. 발견 사항 반영 여부 기록
- [ ] 3.4 Manager(Fable) 검증 요청 — 변경 파일 목록·시험 출력·스크린샷 경로를 보고

## 4. 종결 (Manager 승인 뒤, Opus 가 수행)
- [ ] 4.1 main 병합·푸시(사람 승인 뒤) · PM tracker 재생성 · `make sdd-sync` · 아카이브 · memory retain

## 구현 증거 (Sonnet 5.5, 2026-10-25)
- 1.1 확정: 리터럴 `templates_optimization.go:217`(데이터 미참조, 구 템플릿 `optimizationPage`). 칩 `templates_optimization.go:306,337,352` 가 `.View.Query.PeriodDays` 를 찍고 `handlePerformanceHistory`(`performance_history.go:35-56`) 는 Performance==nil→`Unwired`, 오류→`LoadErr` 로 `View` 영값 → 0일. 성공 경로 `performance.DefaultQuery`(`query.go:36`) PeriodDays 30.
- 1.2 판정: H1(30일이 정본, 의도된 기본값) + 미배선 시 0일 가면(spec 706 위반)이라는 둘째 결함. proposal 가설 절 갱신. 일수 불변. H2 아님 → issues.md 불필요.
- 2.1 RED(생산 변경을 stash 한 상태, `go test ./internal/console/ -run TestPerformanceWindowIsSingleSource -count=1`):
  - wired: history=`최근 30일 · 전체 시장 · 전체 lane · complete lineage only` optimization=`… complete lineage`
  - unwired: history=`최근 0일 · 전체 시장 · 전체 lane · complete lineage only` optimization=`최근 30일 · … complete lineage` want=`미측정 · seam_unwired`. FAIL. (렌더 길이>0 단언 포함)
- 2.2 GREEN: `performanceWindow`/`newPerformanceWindow`(`performance_history.go`), 두 화면이 필드로 받음, 템플릿 리터럴 삭제. 사유 코드는 기존 `reasonSeamUnwired`(`overview.go`) 재사용. 편차: LoadErr 는 창을 미측정으로 바꾸지 않음(review.md 구현 리뷰 #1).
- 2.3 변이: 템플릿에 리터럴 복구 → unwired 하위시험 FAIL(`optimization="최근 30일 · … only"` vs want `미측정 · seam_unwired`). `cp` 복구 후 `sha256sum -c` OK, `PerformanceWindow.Text` 심볼 1개 확인, 시험 3 passed.
- 3.1 `gofmt -l ./internal/console` 출력 없음; `go vet ./internal/console/` 이상 없음; `go test ./internal/console/ -count=1` 786 passed; `go test ./cmd/tossctl/ -run 'Console|Optimization|Performance' -count=1` 106 passed. 미실행: make gate·전체 스위트.
- 3.2 격리 콘솔(빈 설정 폴더, port 38322, PID 종료 확인 `gone`): 1280·390 모두 두 화면 `미측정 · seam_unwired`. 증거 `evidence/{history,optimization}-{1280,390}.png`(4장). 버튼 클릭 0.
- 3.3 review.md 「구현 리뷰」.
