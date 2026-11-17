# Function Logic Map: `Tracker.Restore` (reconcile)

- Source: `internal/reconcile/mismatch.go` (966-1012)
- 편집 전: `../../pre-edit-ast/reconcile--tracker.restore.ast.json`(931~975, sha f72bcbc7d288, 분기 6).
- AST evidence: `ast.json`(편집 뒤 sha 9895f31184e5, 분기 6, return: return@968 return@972 return@1011). 명령 `GOFLAGS=-trimpath go run ./tools/logic-map --file internal/reconcile/mismatch.go --func Tracker.Restore`.
- Risk scan: `risk-pattern-report.md`(일치 0).
- 재번호: `renumber.txt`(`../renumber.py` — 옛/새 분기를 (종류, 그 줄의 소스 문장)으로 정렬. 손 재번호 안 함).

## 재번호

```
B1:if@932 -> B1:if@967
B2:if@936 -> B2:if@971
B3:range@943 -> B3:range@978
B4:if@944 -> B4:if@979
B5:if@967 -> B5:if@1004
B6:if@970 -> B6:if@1007
```

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 장부 활성 행 | — | `ActiveReconcileStates` | 오류 → return(기존) |
| 불변식 | 복원된 행은 결속 없음(삽입 증명 없음) | design D2·D6 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (if@967) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B2 (if@971) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B3 (range@978) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B4 (if@979) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B5 (if@1004) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |
| B6 (if@1007) | 기존 분기(편집 없음 — 편집 전과 같은 문장, 재번호 표) | 기존과 같음 | 기존과 같음 | 기존 시험(패키지 GREEN) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Journal.ActiveReconcileStates`(call@970) | 복원 | 기존 | AST |
| `t.Gate.RebuildReconcileProjection`(call@1008) | 게이트 재구성 | 기존 | AST |

## State mutations and fallbacks

- 편집: 잠금 안 대입 `t.terminal = terminalState{}` 1줄(분기 아님) — 같은 객체 재사용에도 결속·무장을 버림(리뷰 3 #3).

## Safety conclusion

- Safe edit boundary: 대입 1줄. 분기 6개 문장 불변.
- High-risk impact: yes(복원) — 블록·게이트 복원 그대로, 결속만 비움(보수 방향).
