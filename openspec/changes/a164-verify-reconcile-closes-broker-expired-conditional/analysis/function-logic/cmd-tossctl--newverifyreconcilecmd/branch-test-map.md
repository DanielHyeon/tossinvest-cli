# Branch Test Map: `newVerifyReconcileCmd`

- Source: `cmd/tossctl/verify_reconcile.go` (44-79)
- AST branches 0 — 시험 칸은 시험별 coverage 측정(본문 블록 실행)으로 채움

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 | `TestVerifyReconcileExpiredIsAFlagOfTheSameMutatingCommand` · `TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand` · `TestVerifyReconcileOffersNoIdentifierSymbolOrBypassFlag` | 해당 없음 — a164 GREEN 재추출 — 새 갈래 RED 는 analysis/red-green/red-2026-10-22.md | 기준선 통과(측정, 아래 「측정」) |

## 측정

- 기준선: 구현 base `aec9bde67` worktree, `go test -run '^<Test>$' -coverprofile` 을 시험 하나씩(verifylive 의
  `reconcile_*_test.go`·`record_filled_test.go`·`record_replay_test.go`·`report_test.go`, official 의 `reconcile_reads*_test.go`,
  cmd 의 `verify_reconcile*_test.go` 를 `-tags tossos_testseams` 로) 돌려 279 시험 프로필을 얻고, 분기 줄의 본문 `{` 블록 실행 수가
  0 보다 큰 시험만 적었다. 기준선 스위트는 verifylive 634 · official 455 · cmd(Reconcile 계열) 62 전부 통과.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 0 에서 편집 뒤 0 로. `source_sha256` 82787649cc41… → 4008a00b6a32….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: 분기 없음(둘 다 0).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
