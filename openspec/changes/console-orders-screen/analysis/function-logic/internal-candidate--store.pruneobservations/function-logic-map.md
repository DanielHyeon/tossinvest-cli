# Function Logic Map: `Store.PruneObservations`

- Source: `internal/candidate/store.go`
- AST evidence: `ast.json` (revision: base, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

**본문 무변경**. 바로 뒤에 삽입된 `PruneExpiredCandidates`의 diff hunk가 이 함수와 교차해 evidence가 요구되었다. base L719-730의 본문은 현재 L828-839와 byte 동일하고, ast.json은 base revision이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `before` | 컷오프 순간 | 호출자(`PruneStale`이 `clk.Now()-retention`으로) | `stamp`는 고정폭(`2006-01-02T15:04:05.000000000Z07:00`)이다. TEXT 컬럼이라 모든 `<`·`>=`·`ORDER BY`가 BINARY 문자열 비교이고, `time.RFC3339Nano`는 소수부 뒤 0을 떼어내 텍스트 순서가 시간 순서가 아니게 만든다 — §1 리뷰가 실행으로 재현한 P0다(`TestTheStoredInstantOrdersLikeTime`). |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ExecContext` 에러 | 없음 | `0, wrap` | 직접 테스트 없음 |
| B2 | `RowsAffected` 에러 | 삭제는 이미 커밋 | `0, wrap` | 직접 테스트 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.db.ExecContext` | `DELETE FROM observations WHERE observed_at < ?` | 단일 문장, busy_timeout이 감싼다 | ast.json calls |
| `stamp` | 고정폭 직렬화 | — | `stamp`는 고정폭(`2006-01-02T15:04:05.000000000Z07:00`)이다. TEXT 컬럼이라 모든 `<`·`>=`·`ORDER BY`가 BINARY 문자열 비교이고, `time.RFC3339Nano`는 소수부 뒤 0을 떼어내 텍스트 순서가 시간 순서가 아니게 만든다 — §1 리뷰가 실행으로 재현한 P0다(`TestTheStoredInstantOrdersLikeTime`). |

## State mutations and fallbacks

- `observations` 테이블만 삭제한다. `candidates`는 문장의 **모양**에서 이미 사정권 밖이다 — D11을 경고가 아니라 구조로 표현한 것.
- `first_seen_at`·`first_price`·`first_rank`는 이 스윕이 절대 가져갈 수 없다. `TestPruningRawObservationsLeavesTheCandidateSummary`·`...LeavesTheBaselineToo`·`...LeavesTheFirstRankToo` 셋이 각각 고정한다.
- 본문 무변경이므로 이 change가 만든 상태 변화는 없다. 이 change가 더한 것은 **다른 함수**(`PruneExpiredCandidates`)로, D11의 두 번째 층에 처음으로 집행자를 준 것이다.

## Safety conclusion

- Safe edit boundary: 무변경 — 인접 삽입만 존재. `candidates`를 이 문장의 사정권에 넣는 것은 이 패키지의 주장 전체를 실패 없이 삭제하는 방향이다
- High-risk impact: no — 발굴 저장소의 원 관측 삭제. 원장 파일과 무관하고 본문은 byte 동일하다.
