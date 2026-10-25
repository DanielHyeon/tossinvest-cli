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
