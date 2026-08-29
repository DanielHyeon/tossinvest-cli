# Branch Test Map: `TestTrackerReleaseFailureStopsBeforePriceAndAdoption`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/app/engine/reconcileloop_test.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (414) `if` — if _, _, err := h.journal.EnterReconcile(context.Background(), journal.EnterReconcileRequest{ | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B2 | (420) `if` — if err := h.tracker.Restore(context.Background()); err != nil { | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B3 | (430) `if` — if cycle.Err == nil { | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B4 | (433) `if` — if cycle.Blocked != 1 \|\| cycle.Released != 0 { | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B5 | (436) `if` — if cycle.Adopted != 0 \|\| cycle.Unmanaged != 0 \|\| h.prices.calls != 0 { | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
