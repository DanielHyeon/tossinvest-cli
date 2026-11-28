# a205 CodeGraphContext context

- CodeGraphContext: not-applicable — 바뀌는 것은 상수 3개 · 주석 2곳 · 시험이고 기존 함수 본문(생산) 편집 0. 관련 문맥 후보 찾기는
  `rg DefaultExitObservationInterval|\.Interval\(\)|lineRefreshInterval|engineRuntimeInterval` 전수 + `tools/logic-map` AST 7(`../ast-evidence/`)로 대신함.
- 그 전수가 찾은 CodeGraph 밖 후보: `Tracer.Run`(tracer.go:315, 필드 경유 호출) · 시험 3곳(`tracer_test.go:195`, `a163_source_age_panic_test.go:66,69` — 모두 `observer.Interval()` 을 읽어
  시계를 그만큼 돌리므로 값에 무관).
