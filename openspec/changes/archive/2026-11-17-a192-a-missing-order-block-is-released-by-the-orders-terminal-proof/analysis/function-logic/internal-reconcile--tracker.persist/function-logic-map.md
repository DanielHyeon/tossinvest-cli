# Function Logic Map: `Tracker.persist` (reconcile)

- Source: `internal/reconcile/mismatch.go` (878-954)
- 편집 전: `../../pre-edit-ast/reconcile--tracker.persist.ast.json`(857~919, sha f72bcbc7d288, 분기 8).
- AST evidence: `ast.json`(편집 뒤 sha 9895f31184e5, 분기 11, return: return@880 return@909 return@915 return@927 return@943 return@947 return@953). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/reconcile/mismatch.go --func Tracker.persist`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 정렬. 손 재번호 안 함).

## 재번호

```
B1:if@858 -> B1:if@879
B2:range@869 -> B2:range@890
B3:if@870 -> B3:if@891
B4:if@886 -> B4:if@908
B5:if@890 -> B5:if@912
B6:range@899 -> B6:range@921
(new/changed, insert) -> B7:if@922  | if block.terminalRelease != nil {
(new/changed, insert) -> B8:if@926  | if err != nil {
(new/changed, insert) -> B9:if@930  | if released {
B7:if@907 -> B10:if@942
B8:if@911 -> B11:if@946
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `out.Added` | pending 블록 전부 | `pendingBlocksForPersistence` | 첫 오류에서 반환(기존) |
| `out.Cleared` | 크레딧 해제 + 종결 증명 해제 후보 | `Observe` | 종결 증명 후보는 tx 재판정이 유지면 해제 없음(오류 아님) |
| 불변식 | 장부 쓰기 내용(EnterReconcile 요청 필드)·크레딧 해제 요청 무변경 | design D11 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@879) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B2 (range@890) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B3 (if@891) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B4 (if@908) | `enterReconcileWithBinding` 오류(편집: 호출만 감쌈) | — | 오류 반환(기존) | 기존 + `TestA192AnUnprovenInsertIsNeverBound` |
| B5 (if@912) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B6 (range@921) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B7 (if@922) | Cleared 블록에 `terminalRelease` 표시가 있음 | `t.releaseOnOrderTerminal` — 같은 판정을 tx 안에서 다시 통과할 때만 장부 해제 | B8/B9 뒤 continue | `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof` |
| B8 (if@926) | 종결 증명 해제의 장부 오류 | — | 오류 반환 — 블록·게이트 유지 | `TestA192ReleaseWriteFailureKeepsTheGate` |
| B9 (if@930) | tx 재판정이 해제함 | `result.released` append | —(continue) | `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof`, `TestA192AFillCommittedBetweenVerdictAndReleaseKeepsTheBlock`(거짓 쪽) |
| B10 (if@942) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B11 (if@946) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.enterReconcileWithBinding`(call@902) | `EnterReconcile` 을 감싸 삽입 증명(entered)으로 결속 생성 | 요청·반환 그대로 전달 | AST |
| `t.releaseOnOrderTerminal`(call@925) | 종결 증명 해제(tx 재판정) | 장부 오류 그대로 반환 | AST |
| `t.Journal.ReleaseReconcile`(call@935) | 크레딧 해제(기존) | 기존 | AST |

## State mutations and fallbacks

- 새 부수효과: 결속 생성(기억), 종결 증명 해제(장부 1행 UPDATE — 새 journal 메서드).
- 크레딧 해제 갈래(B10·B11)는 편집 전 B7·B8 과 같은 문장.

## Safety conclusion

- Safe edit boundary: EnterReconcile 호출을 감싼 함수로 바꿈(같은 요청·같은 반환), Cleared 루프 앞에 갈래 B7~B9.
- High-risk impact: yes. `Terminal == nil` 이면 감싼 함수는 `EnterReconcile` 한 번과 같고 B7 은 표시가 없어 거짓.
