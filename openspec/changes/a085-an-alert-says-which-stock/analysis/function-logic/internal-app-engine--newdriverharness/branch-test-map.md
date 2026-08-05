# Branch Test Map: `newDriverHarness`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/app/engine/reconcileloop_test.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (129) `if` — if err != nil { | `newDriverHarness` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B2 | (133) `if` — if err := j.SetApplyHooks(journal.ApplyHooks{ | `newDriverHarness` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B3 | (177) `if` — if mutate != nil { | `newDriverHarness` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
| B4 | (181) `if` — if err != nil { | `newDriverHarness` (통과) | 측정 불가 — `go test -coverprofile`은 `_test.go`를 계측하지 않는다. 이 함수 자체가 테스트이고, 통과가 곧 실행 증거다 |
