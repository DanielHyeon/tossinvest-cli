# Function Logic Map: `Store.NoteFirstRank`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1466–1563, 분기 15개)
- Risk scan: `risk-pattern-report.md`

최초 관측 위치를 **한 번만** 기록한다. 이 change가 시그니처를
`(ctx, market, symbol, rank, total, at, source)`에서 `(ctx, market, symbol, first FirstRank)`로
바꿨다(issues.md I1).

위치 인자를 아홉 개로 늘리지 않으려는 것이 아니다. **읽기 전체**를 받는 이유는, 그 위치가
`seen_late`에 답할 수 있는지를 결정하는 두 사실 — 소스가 직전 읽기를 갖고 있었는가, 몇 행을
요청했는가 — 이 **위치와 같은 문장에서** 쓰여야 하기 때문이다. 별도 setter는 "출처를 모르는
순위"가 존재하는 창을 만들고, 그 창의 정직한 해석이 바로 이 change가 없애려는 상태다.

새 거부 하나: `first.Requested < 0`. 절대값이 아니라 **부호**가 문제다 — `truncationOf`는
`requested <= 0`을 `unknown`으로 접으므로, 음수가 통과하면 아무것도 만들 수 없는 수가
"아무도 재지 않았다"로 위장한 채 write-once 칼럼에 앉는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `first.Rank`/`first.Total` | 둘 다 양수, `Rank <= Total` | `recordFirsts` | 위반은 오류 |
| `first.At` | non-zero, first_seen_at의 ±TTL 안 | 관측 instant | zero는 오류, 창 밖은 조용한 no-op |
| `first.Requested` | 0(미기록) 또는 양수 | 소스 어댑터 | **음수는 오류** |
| `first.NewlyListed` | 3-상태 | 소스 어댑터 | unknown은 NULL로 저장 |
| `first.Source` | 정규화 가능 | `normaliseSource` | 실패는 오류 |
| `first_rank IS NULL` (UPDATE 조건) | write-once | SQL | 이미 있으면 저장된 값을 그대로 돌려준다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 인자 검증 switch | 없음 | — | `TestARankOfZeroIsNotAFirstSighting` |
| B2 | `rank <= 0 || total <= 0` | 없음 | 오류 | `TestARankOfZeroIsNotAFirstSighting`(0/150, 12/0, -1/150, 0/0) |
| B3 | `rank > total` | 없음 | 오류 | `TestARankOfZeroIsNotAFirstSighting`(151/150) |
| B4 | `at.IsZero()` | 없음 | 오류 | `TestARankOfZeroIsNotAFirstSighting` |
| B5 | `first.Requested < 0` | 없음 | 오류 | `TestANegativeRequestedCountIsRefusedByTheFirstRankWrite` |
| B6 | `normaliseSource` 실패 | 없음 | 오류 | `TestARankOfZeroIsNotAFirstSighting`(공백 source) |
| B7 | `BeginTx` 실패 | 없음 | 오류 | 커버 없음 — I/O |
| B8 | 기존 행 SELECT switch | — | — | `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext` |
| B9 | `sql.ErrNoRows` — 후보가 없다 | 없음 | `ErrNoCandidate` | `TestARankOfZeroIsNotAFirstSighting`(999999) |
| B10 | SELECT 오류 | 없음 | 오류 | 커버 없음 — I/O |
| B11 | `storedRank.Valid && storedRank.Int64 > 0` — 이미 있다 | **쓰지 않는다** | 저장된 `FirstRank`(자격 칼럼 포함) | `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext` · `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` |
| B12 | `parseStamp(firstSeen)` 실패 | 없음 | 오류 | 커버 없음 — 손상된 행 |
| B13 | `!nearFirstSighting(at, seen)` | **쓰지 않는다** | `FirstRank{}, nil` — 오류가 아니다 | `TestARankFromOutsideTheIdentityWindowIsNotStored` |
| B14 | UPDATE 오류 | rollback | 오류 | 커버 없음 — I/O |
| B15 | `tx.Commit()` 오류 | — | 오류 | 커버 없음 — I/O |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `normaliseSource(first.Source)` | id 정규화 | 오류 래핑 | ast.json calls |
| `nearFirstSighting(at, seen)` | ±TTL 동일성 창 | 순수 | ast.json calls |
| `decodeFirstRank(...)` | 이미 있는 값을 자격 칼럼까지 복원 | 파싱 실패는 오류 | ast.json calls |
| `newlyListedToStore(first.NewlyListed)` / `positive(first.Requested)` | 두 자격 칼럼 값 | 오류 없음 | ast.json calls |
| `tx.ExecContext` (UPDATE ... WHERE first_rank IS NULL) | write-once 쓰기 | 동시 writer는 SQL 조건이 막는다 | ast.json calls |

## State mutations and fallbacks

- `candidates` 한 행의 `first_rank`, `first_rank_total`, `first_rank_at`, `first_rank_source`, **`first_rank_newly_listed`, `first_rank_requested`** — 전부 같은 UPDATE 문에서.
- `WHERE first_rank IS NULL`이 write-once를 SQL 수준에서 보장하므로, B11의 early return은 최적화이지 유일한 가드가 아니다.
- fallback 없음. 창 밖 읽기는 오류가 아니라 no-op이고, 그것이 '순위를 싣지 않는 소스가 처음 올린 후보'의 일상 상태다.

## Safety conclusion

- Safe edit boundary: 시그니처 변경(구조체 1개) + 검증 case 1개 + SELECT/UPDATE 칼럼 2개씩. 기존 4개 검증과 창 규칙 무변경.
- High-risk impact: **yes 인접** — write-once 칼럼이고, 잘못 저장하면 그 후보의 남은 생명 전체가 그 값으로 답한다. 이 change의 방향은 저장을 **더 까다롭게** 만드는 쪽(음수 거부)이고, 자격 칼럼은 가산·nullable이다.
