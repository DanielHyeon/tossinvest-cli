# Function Logic Map: `Store.Promote`

- Source: `internal/candidate/store.go`
- Change: `fix-chase-veto-measurement`
- AST evidence: `ast.json` (revision `current`, L1035–1146, 분기 9개)
- Risk scan: `risk-pattern-report.md`

후보를 승격하거나 되살린다. 이 change가 **만료 reset 절에 두 줄**을 더했다 —
`first_rank_newly_listed`와 `first_rank_requested`도 만료 시 NULL로 돌아간다.

방향이 위험한 쪽이라서 그렇다. 만료가 남기고 간 자격 표식은 **새 생명의 순위가 나오지 않은
읽기**를 서술하게 되고, 특히 "소스가 직전 읽기를 갖고 있었다"라는 낡은 `no`는 아무도 자격
부여할 수 없는 최초 순위를 측정 가능하게 만든다.

placeholder는 전부 위치 인자다. 이 문장은 과거에 `?1`과 bare `?`를 섞어 market="0",
symbol="KR"을 쓴 적이 있고, 그 주석이 지금도 위에 있다. 이 change는 CASE 절 2개와 인자
2개를 **같은 수로** 늘렸다(13 reset).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market`/`symbol` | 비어 있지 않은 값 | 호출자 | 빈 값은 오류 |
| `at` | non-zero, 기존 `LastSeenAt` 이상 | `Collect` | 역행은 오류 — first_seen_at은 이 패키지가 지키는 필드다 |
| `expired` | `stateAt(...) == StateExpired` | `stateAt` | true면 provenance·baseline·first_rank·**두 자격 칼럼** 전부 NULL |
| placeholder 수 | CASE 13개 + VALUES 4개 = 17 | 이 문장 | 인자 수 불일치는 SQL 오류 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `market == "" || symbol == ""` | 없음 | 오류 | **커버 없음** |
| B2 | `at.IsZero()` | 없음 | 오류 | **커버 없음** |
| B3 | `BeginTx` 실패 | 없음 | 오류 | 커버 없음 — I/O |
| B4 | `readCandidate` 오류 | rollback | 오류 | 커버 없음 — I/O |
| B5 | `found && at.Before(existing.LastSeenAt)` | 없음 | 오류 — 시계 역행 거부 | `TestABackwardClockStepDoesNotEndTheDiscoveryLoop` · `TestOneRejectedSymbolDoesNotAbortTheMarket` |
| B6 | `found && !expired` | `firstSeen = existing.FirstSeenAt` | — | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` |
| B7 | `ExecContext` 오류 | rollback | 오류 | 커버 없음 — I/O |
| B8 | `tx.Commit()` 오류 | — | 오류 | 커버 없음 — I/O |
| B9 | `found && !expired`(반환 조립) | 이전 생명의 provenance 유지 | — | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readCandidate(ctx, tx, ...)` | 기존 행 | 오류 상승 | ast.json calls |
| `stateAt(existing, s.cooling, s.staleness, at)` | 만료 판정 | 순수 | ast.json calls |
| `tx.ExecContext` | UPSERT + reset 절 | 실패는 defer rollback | ast.json calls |
| `boolToInt(expired)` | 13개 CASE의 단일 스위치 | — | ast.json calls |

## State mutations and fallbacks

- `candidates` 한 행: `first_seen_at`, `last_seen_at`, `cooled_at=NULL`, 그리고 만료면 provenance 4개 + baseline 3개 + first_rank 4개 + **자격 2개**를 NULL/0으로.
- 만료가 아니면 이 13개 칼럼은 전부 보존된다 — D1의 우회로(목록을 한 스캔 떠났다 돌아온 종목이 새 baseline을 얻는 것)를 막는 절이다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: SET 절 2개 + 인자 2개 가산. 기존 11개 절과 조건(`reset`)은 무변경.
- High-risk impact: **yes 인접** — `first_seen_at`은 이 패키지가 존재하는 이유이고 이 문장이 그것을 쓴다. 이 change의 편집은 만료 시 NULL로 만드는 칼럼을 두 개 늘린 것뿐이며, 보존 방향이 아니라 **초기화 방향**이라 측정 가능성을 줄이는 쪽이다.
