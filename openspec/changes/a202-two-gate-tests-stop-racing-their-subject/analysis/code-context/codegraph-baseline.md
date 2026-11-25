# CodeGraph baseline — a202

- 기준: base `7bc35e5d1`(origin/main). 도구 CodeGraph **1.6.0**. 질의 원문 `codegraph-queries.txt`.
- 색인은 주 체크아웃의 것을 읽기만 했다(그 순간 HEAD `65a274aec`, 병행 a166 커밋 7개 위). 대상 파일 4개는 두 커밋에서 같다(빈 diff — 원문 첫 줄).

| 질의 | 결과 | 쓰임 |
|---|---|---|
| callers `TestChildLockAndCrash` · `TestTheReplacedReaderIsClosed` | 없음 | 시험 함수 — 생산 호출자 0 |
| callers `closeEvictedStrategyReader` | `attempt`(`httpapi_strategy_attach.go:180`) 하나 | 닫기는 시도 goroutine 한 자리에서만 — 시험이 기다릴 대상이 하나 |
| callees `TestTheReplacedReaderIsClosed` | `a109Attachment` · `a109WaitFor` · `attach` · `inFlight` · `wake` · 픽스처 타입 | `attach`·`inFlight`·`wake` 는 `console_lifecycle_attach.go` 로 해소됨 — **오해소**(아래 reconciliation) |
| callees `TestChildLockAndCrash` | `Begin` · `fixture` · `readState` · `Root` | 생산 `Begin` 은 부르기만 함 |
| callers `a109WaitFor` | 20(잘림 — 13 표시) | 헬퍼는 고치지 않는다(호출자 20 의 대기 계약 불변) |
| affected 두 파일 `--filter '*_test.go'` | 그 두 파일 자신 | 영향 시험 = 두 패키지 |
